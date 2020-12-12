package datapipe

import (
	"github.com/ariefdarmawan/byter"
	"github.com/eaciit/toolkit"
)

type NewScannerFn func(service, topic string, opt toolkit.M) (Scanner, error)

type Scanner interface {
	Name() string
	Service() string
	SetByter(b byter.Byter) Scanner
	Topic() string
	SetTopic(name string) Scanner
	Scan() (triggered bool, data []toolkit.M, err error)
	MakePayload(data interface{}) (interface{}, error)
	Close()
}
