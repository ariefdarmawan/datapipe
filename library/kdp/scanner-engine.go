package kdp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

var DefaultScannerTick = 100 * time.Millisecond

type scannerEngine struct {
	id       string
	s        *kaos.Service
	sc       Scanner
	cstop    chan bool
	sessions []*scannerSession
	stopFlag bool
}

func NewKxScanner(s *kaos.Service, sc Scanner) *scannerEngine {
	se := new(scannerEngine)
	se.id = sc.ID()
	se.s = s
	se.sc = sc
	se.cstop = make(chan bool)
	go se.Listen()
	return se
}

func (se *scannerEngine) ID() string {
	return se.id
}

func (se *scannerEngine) StopEngine(ctx *kaos.Context, req string) (string, error) {
	se.cstop <- true
	return "OK", nil
}

func (se *scannerEngine) Listen() {
	ev, _ := se.s.EventHub("default")

	// /${ScannerName}/create -- create new scanner session
	ev.SubscribeEx(fmt.Sprintf("/%s/create", se.sc.Name()), se.s, nil,
		func(ctx *kaos.Context, data toolkit.M) (toolkit.M, error) {
			res := toolkit.M{}

			pipeID := data.GetString("PipeID")
			if pipeID == "" {
				return res, errors.New("InvalidPipeID")
			}

			h, err := ctx.DefaultHub()
			if err != nil {
				return res, fmt.Errorf("InvalidDataHub")
			}

			ev, err := ctx.DefaultEvent()
			if err != nil {
				return res, fmt.Errorf("InvalidEventHub")
			}

			ss, err := NewScannerSession(se.sc, NewScannerOptions(ctx.Context(), h, ev, ctx.Log(), DefaultScannerTick, data))
			if err != nil {
				return res, fmt.Errorf("FailCreateSession")
			}
			se.sessions = append(se.sessions, ss)
			go ss.run()

			res.Set("NodeID", se.sc.ID()).Set("SessionID", ss.ID)
			return res, nil
		})

	// /${ScannerName}/stop -- stop session
	ev.Subscribe(fmt.Sprintf("/%s/stop", se.sc.Name()), se.s, nil,
		func(ctx *kaos.Context, data toolkit.M) (string, error) {
			sid := data.GetString("SessionID")
			sessions := []*scannerSession{}
			for _, ss := range se.sessions {
				if ss.ID == sid {
					ss.stop()
					continue
				}
				sessions = append(sessions, ss)
			}
			se.sessions = sessions
			return "", nil
		})

	go func() {
		<-se.cstop
		for _, sess := range se.sessions {
			sess.stop()
		}

		se.s.Log().Infof("unregistered from coordinator")
		ev.Unsubscribe(fmt.Sprintf("/%s/start", se.sc.Name()), se.s, nil)
		ev.Unsubscribe(fmt.Sprintf("/%s/stop", se.sc.Name()), se.s, nil)
	}()
}

func (se *scannerEngine) PingCoordinator() error {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/v1/coordinator/registerscanner",
		model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: "TimeScan", Secret: nodeID},
		&res)
	se.s.Log().Infof("registered to coordinator")

	/*
		go func() {
			for {
				select {
				case <-time.After(5 * time.Second):
					ev.Publish("/v1/coordinator/scannerbeat",
						model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: "TimeScan", Secret: nodeID},
						&res)

				case <-se.s.Context().Done():
					se.cstop <- true
				}
			}
		}()
	*/

	return nil
}

func (se *scannerEngine) UnpingCoordinator() {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/v1/coordinator/deregisterscanner",
		model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: "TimeScan"},
		&res)
	se.s.Log().Infof("deregister from coordinator")
}
