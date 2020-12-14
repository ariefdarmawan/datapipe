package ksctime

import (
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
)

var (
	ScannerID = "TimeScan"
)

type timeScan struct {
	kdp.BaseScanner
	id     string
	config toolkit.M
}

func NewScanner(id string, config toolkit.M) kdp.Scanner {
	ts := new(timeScan)
	ts.config = config
	ts.id = id
	return ts
}

func (o *timeScan) ID() string {
	return o.ID()
}

func (o *timeScan) Scan(request toolkit.M) ([]toolkit.M, bool, error) {
	res := []toolkit.M{}
	return res, false, nil
}
