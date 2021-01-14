package kdp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

var DefaultScannerTick = 200 * time.Millisecond

type ScannerEngine struct {
	id       string
	s        *kaos.Service
	scanner  Scanner
	cstop    chan bool
	sessions []*ScannerSession
	stopFlag bool
}

func NewKxScanner(s *kaos.Service, sc Scanner) *ScannerEngine {
	se := new(ScannerEngine)
	se.id = sc.ID()
	se.s = s
	se.scanner = sc
	se.cstop = make(chan bool)
	go se.Listen()
	return se
}

func (se *ScannerEngine) ID() string {
	return se.id
}

func (se *ScannerEngine) StopEngine() (string, error) {
	se.cstop <- true
	return "OK", nil
}

func (se *ScannerEngine) Listen() {
	ev, _ := se.s.EventHub("default")

	// /${ScannerName}/create -- create new scanner session
	ev.SubscribeEx(fmt.Sprintf("/%s/create", se.scanner.Name()), se.s, nil,
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

			ss, err := NewScannerSession(se.scanner, NewScannerOptions(ctx.Context(), h, ev, ctx.Log(), DefaultScannerTick, data))
			if err != nil {
				return res, fmt.Errorf("FailCreateSession")
			}
			se.sessions = append(se.sessions, ss)
			go ss.run()

			res.Set("NodeID", se.scanner.ID()).Set("SessionID", ss.ID)

			go func() {
				log := NewPipeLog("INFO", pipeID, "new scanner session is created")
				log.ScannerID = se.scanner.Name()
				log.NodeID = se.scanner.ID()
				log.SessionID = ss.ID
				h.Save(log)
			}()

			return res, nil
		})

	// /${ScannerName}/stop -- stop session
	ev.Subscribe(fmt.Sprintf("/%s/stop", se.scanner.Name()), se.s, nil,
		func(ctx *kaos.Context, data toolkit.M) (string, error) {
			sid := data.GetString("SessionID")
			sessions := []*ScannerSession{}
			for _, ss := range se.sessions {
				if ss.ID == sid {
					ss.stop()
					continue
				}
				sessions = append(sessions, ss)
			}
			se.sessions = sessions

			h, _ := ctx.DefaultHub()
			log := NewPipeLog("INFO", "", "scanner session is stopped")
			log.ScannerID = se.scanner.Name()
			log.NodeID = se.scanner.ID()
			log.SessionID = sid
			h.Save(log)

			return "", nil
		})

	go func() {
		<-se.cstop
		for _, sess := range se.sessions {
			sess.stop()
		}

		ev.Unsubscribe(fmt.Sprintf("/%s/create", se.scanner.Name()), se.s, nil)
		ev.Unsubscribe(fmt.Sprintf("/%s/stop", se.scanner.Name()), se.s, nil)
		se.s.Log().Infof("unregistered from coordinator")
	}()
}

func (se *ScannerEngine) PingCoordinator() error {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/coordinator/registerscanner",
		model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: se.scanner.Name(), Secret: nodeID},
		&res)
	se.s.Log().Infof("registered to coordinator")

	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				/*
					ev.Publish("/coordinator/scannerbeat",
						model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: se.scanner.Name(), Secret: nodeID},
						&res)
				*/

			case <-se.s.Context().Done():
				se.cstop <- true
			}
		}
	}()

	return nil
}

func (se *ScannerEngine) UnpingCoordinator() {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/coordinator/deregisterscanner",
		model.ScannerNode{ID: nodeID[len(nodeID)-4:], ScannerID: se.scanner.Name()},
		&res)
	se.s.Log().Infof("deregister from coordinator")
}
