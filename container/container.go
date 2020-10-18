package container

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
)

type container struct {
	Name    string
	Cluster string

	ks       *kaos.Service
	ev       kaos.EventHub
	log      *toolkit.LogEngine
	tick     time.Duration
	ctx      context.Context
	cancelFn context.CancelFunc

	scannerProviders map[string]datapipe.NewScannerFn
	workerProviders  map[string]datapipe.NewWorkerFn

	scanners map[string]datapipe.Scanner
	workers  map[string]datapipe.Worker
}

func NewContainer(name, cluster string, ks *kaos.Service) *container {
	c := new(container)
	c.Name = name
	c.ctx, c.cancelFn = context.WithCancel(context.Background())
	c.Cluster = cluster

	c.tick = 100 * time.Millisecond
	c.scannerProviders = make(map[string]datapipe.NewScannerFn)
	c.workerProviders = make(map[string]datapipe.NewWorkerFn)
	c.scanners = make(map[string]datapipe.Scanner)
	c.workers = make(map[string]datapipe.Worker)
	c.ev, _ = ks.EventHub("default")
	c.log = ks.Log()
	c.ks = ks
	ks.RegisterModel(c, c.Name).SetDeploy(false).SetEvent(true)
	newServiceTopic := path.Join(ks.BasePoint(), cluster, "newservice")
	if c.ev != nil {
		c.ev.Publish(newServiceTopic,
			&datapipe.PipeService{ID: name, ServiceType: "Scanner"},
			nil)
	}
	return c
}

func (c *container) SetLogger(l *toolkit.LogEngine) *container {
	c.log = l
	return c
}

func (c *container) Log() *toolkit.LogEngine {
	return c.log
}

func (c *container) SetTick(t time.Duration) *container {
	if int(t) == 0 {
		t = 100 * time.Millisecond
	}
	c.tick = t
	return c
}

func (c *container) RegisterScanProvider(providerName string, fn datapipe.NewScannerFn) error {
	if c.ev == nil {
		e := fmt.Errorf("fail to register scan provider %s. invalid event hub", providerName)
		c.Log().Error(e.Error())
		return e
	}
	c.scannerProviders[providerName] = fn

	topic := path.Join(c.ks.BasePoint(), c.Cluster, "create", providerName)
	c.ev.SubscribeEx(topic, c.ks, nil,
		func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
			mres := toolkit.M{}
			opt := req
			fn, ok := c.scannerProviders[providerName]
			if !ok {
				return mres, fmt.Errorf("invalid scan provider: %s", providerName)
			}
			sc, err := fn(c.Name, topic, opt)
			if err != nil {
				return mres, fmt.Errorf("fail to create new scanner. %s", err.Error())
			}
			name := sc.Name()
			c.scanners[name] = sc
			c.Log().Infof("New scanner is created. ID: %s, Provider: %s, Options: %s", sc.Name(), providerName, toolkit.JsonString(opt))
			mres.Set("_id", name).Set("ContainerID", c.Name)
			if err := c.runScanScheduler(sc, providerName); err != nil {
				return mres, err
			}
			return mres, nil
		})

	newServiceTopic := path.Join(c.ks.BasePoint(), c.Cluster, "newserviceprovider")
	if err := c.ev.Publish(newServiceTopic,
		&datapipe.PipeServiceProvider{ServiceID: c.Name, Provider: providerName, Add: true},
		nil); err != nil {
		return err
	}
	return nil
}

func (c *container) runScanScheduler(sc datapipe.Scanner, providerName string) error {
	h, _ := c.ks.GetDataHub("default")
	ev, _ := c.ks.EventHub("default")

	if h == nil {
		return fmt.Errorf("no valid datahub connection")
	}

	if ev == nil {
		return fmt.Errorf("no valid eventhubs")
	}

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.Log().Infof("closing scanner %s", sc.Name())
				return

			case <-time.After(c.tick):
				ok, out, e := sc.Scan()

				// is scan error, log the error
				if e != nil {
					c.Log().Errorf("scan error: %s", e.Error())
				}

				// if scan triggers something, dispatch the activity
				if ok {
					// create scan result
					sr := datapipe.NewScanResult(c.Name, sc.Name(), out)
					sr.ScanProvider = providerName

					// get payload
					payload := &datapipe.WorkerMessage{
						ServiceID: c.Name, ScannerID: sc.Name(), ID: sr.ID, ScanProvider: providerName}

					// save result to db
					if e = h.Save(sr); e != nil {
						c.Log().Errorf("saving scan result error: %s", e.Error())
					}

					// dispatch the scan result ID and footprint to worker
					//ns.nc.Publish(fmt.Sprintf("%s|%s", ns.key, name), bs)
					resultTopic := path.Join(c.ks.BasePoint(), sc.Name(), "result")
					if e = ev.Publish(resultTopic, payload, nil); e != nil {
						c.Log().Errorf("fail sending payload: %s", e.Error())
					}
				}
			}
		}
	}()

	return nil
}

func (c *container) RegisterWorkerProvider(providerName string, fn datapipe.NewWorkerFn) error {
	if c.ev == nil {
		e := fmt.Errorf("fail to register scan provider %s. invalid event hub", providerName)
		c.Log().Error(e.Error())
		return e
	}
	c.workerProviders[providerName] = fn

	topic := strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, "create", providerName))
	c.ev.Subscribe(topic, c.ks, nil,
		func(ctx *kaos.Context, req toolkit.M) (string, error) {
			pipeID := req.GetString("Pipe")
			taskID := req.GetString("WorkerName")

			fn, ok := c.workerProviders[providerName]
			if !ok {
				return "", fmt.Errorf("invalid worker provider: %s", providerName)
			}
			wrk, err := fn(providerName, c.Name)
			if err != nil {
				return "", fmt.Errorf("fail to create new worker. %s", err.Error())
			}
			workerID := wrk.Name()
			c.workers[workerID] = wrk

			workerIDNotifTopic := path.Join(c.ks.BasePoint(), c.Cluster, pipeID, taskID, "notify")
			c.ev.Publish(workerIDNotifTopic,
				toolkit.M{}.Set("Pipe", pipeID).Set("Container", c.Name).Set("Provider", providerName).Set("WorkerName", taskID).Set("_id", workerID),
				nil)
			return workerID, nil
		})

	newServiceTopic := path.Join(c.ks.BasePoint(), c.Cluster, "newserviceprovider")
	c.ev.Publish(newServiceTopic,
		&datapipe.PipeServiceProvider{ServiceID: c.Name, Provider: providerName, Add: true},
		nil)
	return nil
}

func (c *container) Scanners() map[string]datapipe.Scanner {
	return c.scanners
}

func (c *container) Workers() map[string]datapipe.Worker {
	return c.workers
}

func (c *container) RemoveScanner(id string) {
	s, ok := c.scanners[id]
	if !ok {
		return
	}
	s.Close()
	delete(c.scanners, id)
}

func (c *container) RemoveWorker(id string) {
	s, ok := c.workers[id]
	if !ok {
		return
	}
	s.Close()
	delete(c.workers, id)
}

func (c *container) Close() {
	for _, s := range c.scanners {
		s.Close()
	}

	for _, w := range c.workers {
		w.Close()
	}
}
