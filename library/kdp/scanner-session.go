package kdp

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type scannerSession struct {
	ID     string
	pipeID string
	sc     Scanner
	cstop  chan bool
	opts   *ScannerOptions
}

func NewScannerSession(sc Scanner, opts *ScannerOptions) (*scannerSession, error) {
	ss := new(scannerSession)
	ss.sc = sc
	ss.pipeID = opts.Data.GetString("PipeID")
	if ss.pipeID == "" {
		return nil, fmt.Errorf("InvalidPipeID")
	}
	ss.opts = opts
	ss.cstop = make(chan bool)
	ss.ID = primitive.NewObjectID().Hex()
	return ss, nil
}

func (ss *scannerSession) stop() {
	ss.cstop <- true
}

func (ss *scannerSession) run() {
ExitLoop:
	for {
		select {
		case <-time.After(ss.opts.Tick):
			ms, trigerred, err := ss.sc.Scan(ss.opts.Data, ss.opts.h, ss.opts.ev)
			if err != nil {
				ss.opts.log.Errorf("%s ScanFail: %s", ss.ID, err.Error())
			}
			if trigerred {
				j := new(Job)
				j.InitialData = ms
				j.PipeID = ss.pipeID
				if err = ss.opts.h.Save(j); err != nil {
					ss.opts.log.Errorf("%s JobInitFail: %s", ss.ID[len(ss.ID)-5:], err.Error())
				} else {
					ss.opts.log.Infof("%s JobInitated %s", ss.ID[len(ss.ID)-5:], j.ID[len(j.ID)-5:])
				}
			}

		case <-ss.cstop:
			break ExitLoop

		case <-ss.opts.ctx.Done():
			break ExitLoop
		}
	}
}
