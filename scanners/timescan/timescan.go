package timescan

import (
	"errors"
	"fmt"
	"time"

	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ProviderID = "TimeScan"

type TimeScan struct {
	opts    *Options
	name    string
	topic   string
	service string

	timeTrack time.Time
	ed        byter.Byter
}

func NewTimeScan(service, topic string, options toolkit.M) (datapipe.Scanner, error) {
	if service == "" {
		return nil, errors.New("service is mandatory")
	}

	opts := new(Options)
	if err := toolkit.Serde(options, opts, ""); err != nil {
		return nil, fmt.Errorf("options is expected to be *Options, unable to serialize: %s", err.Error())
	}

	name := primitive.NewObjectID().Hex()
	ts := new(TimeScan)
	ts.opts = opts
	ts.service = service
	ts.name = name
	ts.timeTrack = time.Now()
	ts.topic = topic
	return ts, nil
}

func (ts *TimeScan) Name() string {
	return ts.name
}

func (ts *TimeScan) Service() string {
	return ts.service
}

func (ts *TimeScan) Scan() (triggered bool, data []toolkit.M, err error) {
	tn := time.Now()
	if tn.Sub(ts.timeTrack) < ts.opts.Tick() {
		return false, nil, nil
	}

	ts.timeTrack = tn
	return true, []toolkit.M{toolkit.M{}.Set("Messsage", ts.opts.Message)}, nil
}

func (ts *TimeScan) MakePayload(data interface{}) (interface{}, error) {
	return nil, nil
}

func (ts *TimeScan) Close() {
}

func (ts *TimeScan) SetByter(b byter.Byter) datapipe.Scanner {
	ts.ed = b
	return ts
}

func (ts *TimeScan) SetTopic(name string) datapipe.Scanner {
	ts.topic = name
	return ts
}

func (ts *TimeScan) Topic() string {
	if ts.topic == "" {
		ts.topic = primitive.NewObjectID().Hex()
	}
	return ts.topic
}
