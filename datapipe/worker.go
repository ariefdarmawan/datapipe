package datapipe

import (
	"github.com/ariefdarmawan/byter"
	"github.com/eaciit/toolkit"
)

type WorkerProcRequest struct {
	Data toolkit.M
}

type NewWorkerFn func(string, interface{}) (Worker, error)

type Worker interface {
	Name() string
	SetByter(b byter.Byter) Worker
	Work(data toolkit.M) (interface{}, error)
	Collect(parm toolkit.M) (interface{}, error)
	Close()
}
