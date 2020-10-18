package datapipe

import (
	"github.com/ariefdarmawan/byter"
	"github.com/eaciit/toolkit"
)

type NewWorkerFn func(string, interface{}) (Worker, error)

type Worker interface {
	Name() string
	SetByter(b byter.Byter) Worker
	Work(data toolkit.M) (interface{}, error)
	Close()
}
