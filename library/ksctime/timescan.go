package ksctime

import (
	"time"

	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ScannerName = "TimeScan"
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
	return ScannerName
}

func (o *timeScan) Scan(request toolkit.M, sess *kdp.ScannerSession) ([]toolkit.M, bool, error) {
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

func (o *timeScan) Close(id string) {
}
