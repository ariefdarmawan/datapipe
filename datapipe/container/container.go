package container

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
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
	mtx      *sync.RWMutex
	cancelFn context.CancelFunc

	scannerProviders map[string]datapipe.NewScannerFn
	workerProviders  map[string]datapipe.NewWorkerFn

	scanners map[string]datapipe.Scanner
	workers  map[string]datapipe.Worker
	connVars map[string]datapipe.ConnectionVariable
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
	c.mtx = new(sync.RWMutex)
	c.connVars = map[string]datapipe.ConnectionVariable{}
	ks.RegisterModel(c, c.Name).SetDeploy(false).SetEvent(true)
	newServiceTopic := path.Join(ks.BasePoint(), cluster, "newservice")
	if c.ev != nil {
		// notify engine for new scan has been created
		if e := c.ev.Publish(newServiceTopic,
			&datapipe.PipeService{ID: name},
			nil); e != nil {
			c.Log().Info("Notify pipe-engine for new service setup")
		}

		// load variables
		res := map[string]datapipe.ConnectionVariable{}
		if e := c.ev.Publish(strings.ToLower(path.Join(ks.BasePoint(), cluster, "getconnvars")), "", &res); e == nil {
			c.Log().Infof("Loading connection variables from pipe-engine. Connection variables found %d, %s", len(res), func() string {
				keys := make([]string, len(res))
				i := 0
				for k := range res {
					keys[i] = k
					i++
				}
				return strings.Join(keys, " ,")
			}())
			c.connVars = res
		}

		// listening for new connvar update
		c.ev.Subscribe(strings.ToLower(path.Join(ks.BasePoint(), cluster, "connvar", "add")), ks, nil,
			func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
				res := toolkit.M{}
				name := req.GetString("Name")
				if name != "" {
					if valAny, ok := req["ConnVar"]; ok {
						value := datapipe.ConnectionVariable{}
						if e := toolkit.Serde(valAny, &value, ""); e == nil {
							c.mtx.Lock()
							defer c.mtx.Unlock()
							c.Log().Infof("Receive instruction to update connection variable %s", name)
							c.connVars[name] = value
						} else {
							c.Log().Infof("Fail to update connection variable %s: %s", name, e.Error())
						}
					}
				} else {
					c.Log().Warning("Receive instruction to update connection variable but no name is specified")
				}
				return res, nil
			})
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
			pipeID := req.GetString("PipeID")
			opt := toolkit.M{}
			if _, ok := req["Options"]; ok {
				toolkit.Serde(req.Get("Options"), &opt, "")
			}
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
			if err := c.runScanScheduler(sc, providerName, pipeID); err != nil {
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

func (c *container) runScanScheduler(sc datapipe.Scanner, providerName, pipeID string) error {
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
					sr := datapipe.NewDataProcess(c.Name, sc.Name(), pipeID, out)
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
					mres := []toolkit.M{}
					resultTopic := path.Join(c.ks.BasePoint(), sc.Name(), "result")
					if e = ev.Publish(resultTopic, payload, &mres); e != nil {
						c.Log().Errorf("fail sending scan process %s: %s", sr.ID, e.Error())
					}
				}
			}
		}
	}()

	return nil
}

func (c *container) RegisterWorkerProvider(providerName string, fn datapipe.NewWorkerFn) error {
	var e error
	if c.ev == nil {
		e = fmt.Errorf("fail to register scan provider %s. invalid event hub", providerName)
		c.Log().Error(e.Error())
		return e
	}

	if e = c.listenForWorkerCreation(providerName, fn); e != nil {
		return e
	}

	return nil
}

func (c *container) listenForWorkerCreation(providerName string, fn datapipe.NewWorkerFn) error {
	c.workerProviders[providerName] = fn
	topic := strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, "create", providerName))
	if e := c.ev.Subscribe(topic, c.ks, nil,
		func(ctx *kaos.Context, req toolkit.M) (string, error) {
			pipeID := req.GetString("Pipe")
			taskID := req.GetString("WorkerName")
			runMode := req.GetString("RunMode")

			fn, ok := c.workerProviders[providerName]
			if !ok {
				return "", fmt.Errorf("invalid worker provider: %s", providerName)
			}
			wrk, err := fn(providerName, c.Name)
			if err != nil {
				return "", fmt.Errorf("fail to create new worker. %s", err.Error())
			}

			workerID := wrk.Name()
			if e := c.listenForWorkRequest(pipeID, taskID, runMode, wrk); e != nil {
				return "", fmt.Errorf("fail to create new worker request listener for worker %s. %s", workerID, e.Error())
			}

			c.workers[workerID] = wrk
			workerIDNotifTopic := strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, pipeID, taskID, "notify"))
			c.ev.Publish(workerIDNotifTopic,
				toolkit.M{}.Set("Pipe", pipeID).Set("Container", c.Name).Set("Provider", providerName).Set("WorkerName", taskID).Set("_id", workerID),
				nil)
			return workerID, nil
		}); e != nil {
		return e
	}

	newServiceTopic := path.Join(c.ks.BasePoint(), c.Cluster, "newserviceprovider")
	c.ev.Publish(newServiceTopic,
		&datapipe.PipeServiceProvider{ServiceID: c.Name, Provider: providerName, Add: true},
		nil)
	return nil
}

func (c *container) listenForWorkRequest(pipeID, taskID, runMode string, wrk datapipe.Worker) error {
	topic := ""

	// worker work
	if runMode != "CollectOnly" {
		if runMode == "Collect" {
			topic = strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, wrk.Name()))
		} else {
			topic = strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, pipeID, taskID))
		}
		c.ev.SubscribeEx(topic, c.ks, nil, func(ctx *kaos.Context, data toolkit.M) (interface{}, error) {
			res, err := wrk.Work(data)
			return res, err
		})
	}

	// worker collect
	if runMode == "Collect" || runMode == "CollectOnly" {
		topic = strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, wrk.Name(), "collect"))
	} else {
		topic = strings.ToLower(path.Join(c.ks.BasePoint(), c.Cluster, pipeID, taskID, "collect"))
	}
	c.ev.SubscribeEx(topic, c.ks, nil, func(ctx *kaos.Context, parm toolkit.M) (interface{}, error) {
		if runMode == "Collect" {
			return wrk.Collect(parm)
		}
		return toolkit.M{}, nil
	})

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
