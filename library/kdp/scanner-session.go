package kdp

import (
	"fmt"
	"time"

	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScannerSession struct {
	ID     string
	PipeID string
	Opts   *ScannerOptions

	sc    Scanner
	cstop chan bool
}

func NewScannerSession(sc Scanner, opts *ScannerOptions) (*ScannerSession, error) {
	ss := new(ScannerSession)
	ss.sc = sc
	ss.PipeID = opts.Data.GetString("PipeID")
	if ss.PipeID == "" {
		return nil, fmt.Errorf("InvalidPipeID")
	}
	ss.Opts = opts
	ss.cstop = make(chan bool)
	ss.ID = primitive.NewObjectID().Hex()
	return ss, nil
}

func (ss *ScannerSession) stop() {
	ss.cstop <- true
}

func (ss *ScannerSession) run() {
ExitLoop:
	for {
		select {
		case <-time.After(ss.Opts.Tick):
			trData, err := TranslateM(ss.Opts.Data, ss.Opts.Hub(), ss.PipeID, "")
			//ss.sc.Log().Infof("TR Data: %s", toolkit.JsonString(trData))
			if err != nil {
				ss.sc.Log().Errorf("TranslateError: %s", err.Error())
				continue
			}

			ms, trigerred, err := ss.sc.Scan(trData, ss)
			if err != nil {
				ss.Opts.Log().Errorf("%s ScanFail: %s", ss.ID, err.Error())
				pl := NewPipeLog("ERROR", ss.PipeID, err.Error())
				pl.ScannerID = ss.sc.Name()
				pl.NodeID = ss.ID
				go ss.Opts.Hub().Save(pl)
			}

			if trigerred && len(ms) > 0 {
				createJobTopic := "/coordinator/createjob"
				createdID := ""
				if ss.Opts.Event().Publish(createJobTopic, toolkit.M{}.Set("PipeID", ss.PipeID).Set("Data", ms), &createdID); createdID != "" {
					ss.Opts.Log().Infof("%s JobInitiated %s", ss.ID, createdID)
				}
			}

		case <-ss.cstop:
			break ExitLoop

		case <-ss.Opts.Context().Done():
			break ExitLoop
		}
	}

	ss.Opts.Log().Infof("%s has been stopped", ss.ID[len(ss.ID)-5:])
}
