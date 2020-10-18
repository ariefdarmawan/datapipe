package hello

import (
	"time"

	"git.kanosolution.net/kano/appkit"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ProviderID = "HelloWorker"
)

type worker struct {
	bt   byter.Byter
	log  *toolkit.LogEngine
	name string
}

func NewWorker(name string, opt interface{}) (datapipe.Worker, error) {
	w := new(worker)
	w.log = appkit.Log()
	w.name = primitive.NewObjectID().Hex()
	return w, nil
}

func (w *worker) Name() string {
	return w.name
}

func (w *worker) SetByter(b byter.Byter) datapipe.Worker {
	w.bt = b
	return w
}

func (w *worker) Work(data toolkit.M) (interface{}, error) {
	msg := data.GetString("message")
	w.log.Infof("Hello from data-pipe worker. Message: %s, Sent: %v", msg, time.Now())
	return "OK", nil
}

func (w *worker) Close() {
}
