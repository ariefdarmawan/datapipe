package datapipe

import (
	"github.com/ariefdarmawan/byter"
	"github.com/eaciit/toolkit"
)

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
