package datapipe_test

import (
	"testing"
	"time"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"

	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe"
	"github.com/ariefdarmawan/datapipe/container"
	"github.com/ariefdarmawan/datapipe/scanners/mgoscan"
	"github.com/ariefdarmawan/datapipe/scanners/timescan"
	"github.com/ariefdarmawan/datapipe/worker/hello"
	"github.com/ariefdarmawan/datapipe/worker/mgoworker"
	"github.com/eaciit/toolkit"
	"github.com/kanoteknologi/knats"
	"github.com/nats-io/nats.go"

	_ "github.com/ariefdarmawan/flexmgo"
	cv "github.com/smartystreets/goconvey/convey"
)

var (
	natsAddr  = nats.DefaultURL
	basePoint = "df"
	cluster   = "sambas"
	mgoAddr   = "mongodb://localhost:27017/ingestor"
)

func TestPipe(t *testing.T) {
	mgoHub := datahub.NewHub(func() (dbflex.IConnection, error) {
		c, e := dbflex.NewConnectionFromURI(mgoAddr, nil)
		if e != nil {
			return nil, e
		}
		if e = c.Connect(); e != nil {
			return nil, e
		}
		return c, nil
	}, true, 10)
	defer mgoHub.Close()

	secret := "my-secret"
	bt := byter.NewByter("")
	ev := knats.NewEventHub(natsAddr, bt).SetSecret(secret)
	defer ev.Close()

	cv.Convey("pipeline service", t, func() {
		log := appkit.LogWithPrefix("PipeEngine")
		//log.LogToStdOut = false

		s := kaos.NewService().
			SetBasePoint(basePoint).
			SetLogger(log).
			RegisterDataHub(mgoHub, "default").
			RegisterEventHub(ev, "default", secret)
		engine := datapipe.NewPipeEngine(cluster, s)
		e := s.ActivateEvent()
		cv.So(e, cv.ShouldBeNil)
		e = s.Start()
		cv.So(e, cv.ShouldBeNil)

		engine.AddConnVar("mgoSource", datapipe.ConnectionVariable{ConnTxt: mgoAddr, PoolSize: 10})

		cv.Convey("scanner service", func() {
			log2 := appkit.LogWithPrefix("ScannerService")
			log2.LogToStdOut = true
			ss := kaos.NewService().
				SetBasePoint(basePoint).
				SetLogger(log2).
				RegisterEventHub(ev, "default", secret).
				RegisterDataHub(mgoHub, "default")

			tc := container.NewContainer("ScannerService", cluster, ss)
			tc.RegisterScanProvider(timescan.ProviderID, timescan.NewTimeScan)
			tc.RegisterScanProvider(mgoscan.ProviderID, mgoscan.NewMgoScan)
			//e = ss.ActivateEvent()
			cv.So(e, cv.ShouldBeNil)
			time.Sleep(100 * time.Millisecond)

			cv.So(len(engine.Services()), cv.ShouldEqual, 1)
			cv.So(len(engine.Services()[0].Providers), cv.ShouldEqual, 2)
			log.Info("Scanner Service: " + toolkit.JsonString(engine.Services()[0]))

			cv.Convey("worker service", func() {
				log3 := appkit.LogWithPrefix("WorkerService")
				ws := kaos.NewService().SetBasePoint(basePoint).SetLogger(log3).RegisterEventHub(ev, "default", secret)
				wc := container.NewContainer("WorkerService", cluster, ws)
				wc.RegisterWorkerProvider(mgoworker.ProviderID, mgoworker.NewWorker)
				wc.RegisterWorkerProvider(hello.ProviderID, hello.NewWorker)
				time.Sleep(100 * time.Millisecond)
				log3.Info("Worker Services: " + toolkit.JsonString(engine.Services()[1]))
				cv.So(len(engine.Services()), cv.ShouldEqual, 2)
				cv.So(len(engine.Services()[1].Providers), cv.ShouldEqual, 2)

				cv.Convey("orchestrate datapipe 1", func() {
					engine.AddConnVar("mgoTarget", datapipe.ConnectionVariable{ConnTxt: mgoAddr, PoolSize: 10})

					pipe := datapipe.NewPipe("pipe-a")
					pipe.UseScanner(timescan.ProviderID, timescan.Options{500, "SimSalabim"})
					pipe.AddWorker("ReverseWord", datapipe.PipeWorkerConfig{
						Provider:  hello.ProviderID,
						ClosePipe: datapipe.CloseOnError,
						Options:   toolkit.M{}.Set("Op", "Reverse"),
					})
					pipe.AddWorker("GetCount", datapipe.PipeWorkerConfig{
						Input:     "ReverseWord",
						Provider:  hello.ProviderID,
						ClosePipe: datapipe.CloseOnError,
						RunMode:   "Collect",
						Options:   toolkit.M{}.Set("Op", "Count"),
					})
					pipe.AddWorker("WriteToDb", datapipe.PipeWorkerConfig{
						Input:     "ReverseWord",
						Provider:  mgoworker.ProviderID,
						ClosePipe: datapipe.CloseOnAny,
						Options:   toolkit.M{}.Set("Op", "Count"),
					})
					err := engine.Attach(pipe)
					cv.So(err, cv.ShouldBeNil)
					time.Sleep(200 * time.Millisecond)
					cv.Printf("\nDebug Pipe: %s\n", toolkit.JsonStringIndent(engine.Pipes(), "\t"))
					cv.Printf("\nDebug Pipe Objects: %s\n", toolkit.JsonStringIndent(engine.PipeObjects(), "\t"))

					cv.Convey("validate scanner object", func() {
						cv.So(len(tc.Scanners()), cv.ShouldEqual, 1)
						sc, ok := tc.Scanners()[pipe.ScannerConfig.ScannerID]
						cv.So(ok, cv.ShouldBeTrue)
						cv.So(sc.Name(), cv.ShouldEqual, pipe.ScannerConfig.ScannerID)

						cv.Convey("validate worker object", func() {
							cv.So(len(wc.Workers()), cv.ShouldEqual, 2)
							cv.So(len(pipe.WorkerConfigs[0].WorkerIDs), cv.ShouldBeGreaterThan, 0)
							wc, ok := wc.Workers()[pipe.WorkerConfigs[0].WorkerIDs[0]]
							cv.So(ok, cv.ShouldBeTrue)
							cv.So(wc.Name(), cv.ShouldEqual, pipe.WorkerConfigs[0].WorkerIDs[0])
						})
					})
					time.Sleep(1100 * time.Millisecond)
				})
			})
		})
	})
}
