package kdp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

var DefaultWorkerTick = 100 * time.Millisecond

type workerEngine struct {
	id       string
	s        *kaos.Service
	sc       Worker
	cstop    chan bool
	sessions []*workerSession
	stopFlag bool
}

func NewKxWorker(s *kaos.Service, sc Worker) *workerEngine {
	se := new(workerEngine)
	se.id = sc.ID()
	se.s = s
	se.sc = sc
	se.cstop = make(chan bool)
	go se.Listen()
	return se
}

func (se *workerEngine) ID() string {
	return se.id
}

func (se *workerEngine) StopEngine(ctx *kaos.Context, req string) (string, error) {
	se.cstop <- true
	return "OK", nil
}

func (se *workerEngine) Listen() {
	ev, _ := se.s.EventHub("default")

	// /${WorkerName}/create -- create new worker session
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

			ss, err := NewWorkerSession(se.sc, NewWorkerOptions(ctx.Context(), h, ev, ctx.Log(), data))
			if err != nil {
				return res, fmt.Errorf("FailCreateSession")
			}
			se.sessions = append(se.sessions, ss)
			go ss.run()

			res.Set("NodeID", se.sc.ID()).Set("SessionID", ss.ID)
			return res, nil
		})

	// /${WorkerName}/stop -- stop session
	ev.Subscribe(fmt.Sprintf("/%s/stop", se.sc.Name()), se.s, nil,
		func(ctx *kaos.Context, data toolkit.M) (string, error) {
			sid := data.GetString("SessionID")
			sessions := []*workerSession{}
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

func (se *workerEngine) PingCoordinator() error {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/v1/coordinator/registerworker",
		model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: "TimeScan", Secret: nodeID},
		&res)
	se.s.Log().Infof("registered to coordinator")

	/*
		go func() {
			for {
				select {
				case <-time.After(5 * time.Second):
					ev.Publish("/v1/coordinator/workerbeat",
						model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: "TimeScan", Secret: nodeID},
						&res)

				case <-se.s.Context().Done():
					se.cstop <- true
				}
			}
		}()
	*/

	return nil
}

func (se *workerEngine) UnpingCoordinator() {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/v1/coordinator/deregisterworker",
		model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: "TimeScan"},
		&res)
	se.s.Log().Infof("deregister from coordinator")
}
