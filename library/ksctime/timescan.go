package ksctime

import (
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
)

var (
	ScannerID = "TimeScan"
)

type timeScan struct {
	config toolkit.M
}

func NewScanner(config toolkit.M) kdp.Scanner {
	ts := new(timeScan)
	ts.config = config
	return ts
}
