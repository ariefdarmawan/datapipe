package kdp

import (
	"context"
	"time"

	"git.kanosolution.net/kano/appkit"
	"github.com/eaciit/toolkit"
)

type scannerSession struct {
	log     *toolkit.LogEngine
	ctx     context.Context
	scanner Scanner
}

func NewScannerSession(ctx context.Context, log *toolkit.LogEngine, scanner Scanner) *scannerSession {
	ss := &scannerSession{log: log, ctx: ctx, scanner: scanner}
	return ss
}

func (ss *scannerSession) Log() *toolkit.LogEngine {
	if ss.log == nil {
		ss.log = appkit.LogWithPrefix(ss.scanner.ID())
	}
	return ss.log
}

func (ss *scannerSession) Run(req toolkit.M) {
	opts := ss.scanner.Options()
	for {
		select {
		case <-time.After(opts.TickDuration):
			res, scanned, err := ss.scanner.Scan(req)
			if err != nil {
				ss.Log().Errorf("ScanError: %s", err.Error())
				continue
			}

			if scanned {
				ss.scanner.Datahub().SaveAny("DPJobs", toolkit.M{}.Set("Data", res))
			}

		case <-ss.ctx.Done():
			return
		}
	}
}
