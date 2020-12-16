package ksctime

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ScannerID = "TimeScan"
)

type timeScan struct {
	kdp.BaseScanner
	id string
}

func NewScanner(id string) kdp.Scanner {
	ts := new(timeScan)
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}
	ts.id = id
	return ts
}

func (o *timeScan) ID() string {
	return o.id
}

func (o *timeScan) Name() string {
	return ScannerID
}

func (o *timeScan) Scan(request toolkit.M, h *datahub.Hub, ev kaos.EventHub) ([]toolkit.M, bool, error) {
	tick := time.Duration(request.GetInt("Every"))
	switch request.GetString("Unit") {
	case "ms":
		tick = tick * time.Millisecond
	default:
		tick = tick * time.Second
	}

	if tick < time.Duration(100*time.Millisecond) {
		tick = 100 * time.Millisecond
	}

	time.Sleep(tick)
	return []toolkit.M{
		toolkit.M{}.Set("Code", primitive.NewObjectID().Hex()),
	}, true, nil
}
