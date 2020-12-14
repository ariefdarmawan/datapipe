package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kext/mod/kmdb"
	"git.kanosolution.net/kano/kext/mod/kmui"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe/engine"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/ariefdarmawan/kconfigurator"
	"github.com/google/uuid"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"

	"github.com/ariefdarmawan/datapipe/library/kdp"
	_ "github.com/ariefdarmawan/datapipe/library/kfslocal"
	_ "github.com/ariefdarmawan/datapipe/library/kfsmn"
	_ "github.com/ariefdarmawan/flexmgo"
)

var (
	config      = flag.String("config", "config/app.yml", "path to config file")
	appConfig   = new(kconfigurator.AppConfig)
	serviceName = "backend"
	version     = "v1"
	log         = appkit.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()

	// log and config
	e := appkit.ReadConfig(*config, appConfig)
	if e != nil {
		log.Error(e.Error())
		os.Exit(1)
	}

	evServer := appConfig.EventServer
	if evServer.Group == "" {
		evServer.Group = uuid.New().String()
	}
	appConfig.EventServer = evServer

	// service
	ev := knats.NewEventHub(evServer.Server, byter.NewByter("")).SetSignature(appConfig.EventServer.Group)
	defer ev.Close()

	s := kaos.NewService().SetLogger(log).
		SetBasePoint(version).
		RegisterEventHub(ev, "default", evServer.Group)

	// datahub
	h, err := kconfigurator.MakeHub(appConfig, "default")
	if err != nil {
		log.Errorf("InvalidHub: %s", e.Error())
		os.Exit(1)
	}
	s.RegisterDataHub(h, "default")

	// config
	configurator := kconfigurator.NewConfigurator(appConfig)
	s.RegisterModel(configurator.EventModel(), "/config").SetEvent(true).SetDeploy(false)

	// model registration
	mdb := kmdb.New()
	mly := kmui.New()

	coordinator := engine.NewCoordinator(h)
	defer coordinator.CloseNodes(ev)

	s.RegisterModel(new(model.Storage), "storage").SetMod(mdb, mly)
	s.RegisterModel(new(model.Connection), "connection").SetMod(mdb, mly)
	s.RegisterModel(new(model.Variable), "variable").SetMod(mdb, mly)
	s.RegisterModel(new(model.ScannerInfo), "scanner").SetMod(mdb, mly).
		RegisterHook(func(ctx *kaos.Context, req *model.ScannerInfo) error {
			coordinator.RegisterScanner(ctx, &model.ScannerNode{ScannerID: req.ID})
			return nil
		}, "PostSave").
		RegisterHook(func(ctx *kaos.Context, req *model.ScannerInfo) error {
			coordinator.DeregisterScanner(ctx, &model.ScannerNode{ScannerID: req.ID})
			return nil
		}, "PostDelete")
	s.RegisterModel(new(model.WorkerInfo), "worker").SetMod(mdb, mly).
		RegisterHook(func(ctx *kaos.Context, req *model.WorkerInfo) error {
			coordinator.RegisterWorker(ctx, &model.WorkerNode{WorkerID: req.ID})
			return nil
		}, "PostSave").
		RegisterHook(func(ctx *kaos.Context, req *model.ScannerInfo) error {
			coordinator.DeregisterWorker(ctx, &model.WorkerNode{WorkerID: req.ID})
			return nil
		}, "PostDelete")

	s.RegisterModel(new(kdp.Pipe), "pipe").SetMod(mdb, mly)
	s.RegisterModel(new(kdp.PipeItem), "pipeitem").SetMod(mly)
	s.RegisterModel(new(kdp.PipeItemRoute), "pipeitemroute").SetMod(mly)

	s.RegisterModel(coordinator, "coordinator").SetEvent(true).SetDeploy(false)
	s.RegisterModel(coordinator.RESTEngine(), "coordinator").SetDeploy(true)
	s.RegisterModel(new(engine.StorageEngine), "storage")

	// deploy
	if e = s.ActivateEvent(); e != nil {
		log.Errorf("unable to deploy events: %s", e.Error())
		os.Exit(1)
	}
	mux := http.NewServeMux()

	if e = hd.NewHttpDeployer().Deploy(s, mux); e != nil {
		log.Errorf("unable to deploy. %s", e.Error())
		os.Exit(1)
	}

	// run service
	csign := make(chan os.Signal)
	hostName := appConfig.Hosts[serviceName]
	if hostName == "" {
		log.Errorf("unable to start service %s, hostname is not defined", serviceName)
		os.Exit(1)
	}
	go func() {
		s.Log().Infof("starting %v service on %s", serviceName, hostName)
		err := http.ListenAndServe(hostName, mux)
		if err != nil {
			csign <- syscall.SIGINT
		}
	}()

	// grace shutdown
	signal.Notify(csign, os.Interrupt, os.Kill)
	<-csign
	log.Infof("stopping %v service", serviceName)
}
