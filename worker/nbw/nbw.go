package nbw

import (
	"context"
	"fmt"
	"path"
	"time"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NatsBaseWorker struct {
	tick   time.Duration
	chStop chan bool

	h      *datahub.Hub
	ev     kaos.EventHub
	logger *toolkit.LogEngine
	ed     byter.Byter

	providers map[string]datapipe.NewWorkerFn
	workers   map[string]datapipe.Worker
	key       string
	ks        *kaos.Service
	ctx       context.Context
}

func NewNatsBaseWorker(key string, ks *kaos.Service) *NatsBaseWorker {
	if key == "" {
		key = primitive.NewObjectID().Hex()
	}

	w := new(NatsBaseWorker)
	w.ks = ks
	w.h, _ = ks.GetDataHub("default")
	w.ev, _ = ks.EventHub("default")
	w.tick = 100 * time.Millisecond
	w.key = key
	w.ed = byter.NewByter("")
	w.chStop = make(chan bool)
	w.providers = make(map[string]datapipe.NewWorkerFn)
	w.workers = make(map[string]datapipe.Worker)
	return w
}

func (w *NatsBaseWorker) Start() error {
	_, e := w.ks.GetDataHub("default")
	if e != nil {
		return fmt.Errorf("worker servicee %s has no valid data hub. %s", w.key, e.Error())
	}
	_, found := w.ks.EventHub("default")
	if !found {
		return fmt.Errorf("worker servicee %s has no valid event hub", w.key)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-w.chStop
		cancelFn()
	}()
	w.ctx = ctx

	return nil
}

func (w *NatsBaseWorker) Log() *toolkit.LogEngine {
	if w.logger == nil {
		w.logger = appkit.Log()
	}
	return w.logger
}

func (w *NatsBaseWorker) RegisterWorkerProvider(name string, fn datapipe.NewWorkerFn) datapipe.WorkerService {
	w.providers[name] = fn
	return w
}

func (w *NatsBaseWorker) RegisterWorker(provider, name, scanServiceID, scannerID string, opts interface{}) (datapipe.Worker, error) {
	fn, found := w.providers[provider]
	if !found {
		return nil, fmt.Errorf("worker provider %s is not found")
	}

	if name == "" {
		name = primitive.NewObjectID().Hex()
	}

	worker, e := fn(name, opts)
	worker.SetByter(w.ev.Byter())
	if e != nil {
		return nil, fmt.Errorf("fail create new worker. %s", e.Error())
	}
	w.workers[name] = worker
	w.ev.SubscribeEx(path.Join(w.ks.BasePoint(), scannerID), w.ks, nil,
		func(ctx *kaos.Context, req *datapipe.WorkerMessage) (string, error) {
			// get sr
			sr := new(datapipe.ScanResult)
			sr.ID = req.ID
			if e := w.h.Get(sr); e != nil {
				w.Log().Errorf("%s.%s fail to get request data %s: %s", w.key, name, req.ID, e.Error())
				return "", fmt.Errorf("%s.%s fail to get request data %s: %s", w.key, name, req.ID, e.Error())
			}

			for _, data := range sr.Data {
				//w.Log().Infof("running worker %s.%s data %s", w.key, name, toolkit.JsonString(data))
				_, err := worker.Work(data)
				if err != nil {
					w.Log().Errorf("%s.%s fail to run worker: %s", w.key, name, err.Error())
					return "", fmt.Errorf("%s.%s fail to run worker: %s", w.key, name, err.Error())
				}
			}
			updReq := datapipe.UpdateRequest{ID: req.ID, Status: "Done"}
			if e := w.ev.Publish(fmt.Sprintf("%s|update", scanServiceID), updReq, nil); e != nil {
				w.Log().Errorf("%s.%s fail to update status: %s", w.key, name, e.Error())
				return "", fmt.Errorf("%s.%s fail to update status: %s", w.key, name, e.Error())
			}
			return "OK", nil
		})
	return worker, nil
}

func (w *NatsBaseWorker) Workers() []datapipe.Worker {
	ws := []datapipe.Worker{}
	for _, item := range w.workers {
		ws = append(ws, item)
	}
	return ws
}

func (w *NatsBaseWorker) Stop() {
	w.chStop <- true
}
