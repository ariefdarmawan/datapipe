package kdp

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kaos/kpx"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

var DefaultWorkerTick = 100 * time.Millisecond

type workerEngine struct {
	id       string
	s        *kaos.Service
	worker   Worker
	cstop    chan bool
	sessions []*WorkerSession
	stopFlag bool
}

func NewKxWorker(s *kaos.Service, sc Worker) *workerEngine {
	se := new(workerEngine)
	se.id = sc.ID()
	se.s = s
	se.worker = sc
	se.cstop = make(chan bool)
	go se.Listen()
	return se
}

func (se *workerEngine) ID() string {
	return se.id
}

func (se *workerEngine) StopEngine() (string, error) {
	se.cstop <- true
	return "OK", nil
}

func (se *workerEngine) Listen() {
	ev, _ := se.s.EventHub("default")

	// /${WorkerName}/create -- listen new worker session
	ev.Subscribe(fmt.Sprintf("/%s/create", se.worker.Name()), se.s, nil, se.initWorkerSession)

	ev.SubscribeEx(fmt.Sprintf("/%s/assign", se.worker.Name()), se.s, nil, se.initWorkerSession)

	// /${WorkerName}/stop -- stop session
	ev.Subscribe(fmt.Sprintf("/%s/stop", se.worker.Name()), se.s, nil,
		func(ctx *kaos.Context, data toolkit.M) (string, error) {
			found := false
			log := NewPipeLog("INFO", "", "worker session has been activated")
			sid := data.GetString("SessionID")
			sessions := []*WorkerSession{}

			for _, ss := range se.sessions {
				if ss.ID == sid {
					log.PipeID = ss.PipeID
					log.JobID = ss.JobID
					found = true
					ss.stop()
					continue
				}
				sessions = append(sessions, ss)
			}
			se.sessions = sessions

			if found {
				go func() {
					h, _ := ctx.DefaultHub()
					log.NodeID = se.worker.ID()
					log.SessionID = sid
					h.Save(log)
				}()
			}

			return "", nil
		})

	ev.Subscribe("/worker/stop", se.s, nil, se.stopWorker)

	go func() {
		<-se.cstop
		for _, sess := range se.sessions {
			sess.stop()
		}

		ev.Unsubscribe(fmt.Sprintf("/%s/start", se.worker.Name()), se.s, nil)
		ev.Unsubscribe(fmt.Sprintf("/%s/stop", se.worker.Name()), se.s, nil)
		se.s.Log().Infof("unregistered from coordinator")
	}()
}

func (se *workerEngine) PingCoordinator() error {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/coordinator/registerworker",
		model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: se.worker.Name(), Secret: nodeID},
		&res)
	se.s.Log().Infof("registered to coordinator")

	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				/*
					ev.Publish("/coordinator/workerbeat",
						model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: se.worker.Name(), Secret: nodeID},
						&res)
				*/

			case <-se.s.Context().Done():
				se.cstop <- true
			}
		}
	}()

	return nil
}

func (se *workerEngine) UnpingCoordinator() {
	nodeID := se.id
	ev, _ := se.s.EventHub("default")

	res := toolkit.M{}
	ev.Publish("/coordinator/deregisterworker",
		model.WorkerNode{ID: nodeID[len(nodeID)-4:], WorkerID: se.worker.Name()},
		&res)
	se.s.Log().Infof("deregister from coordinator")
}

type WorkerSessionCreateRequest struct {
	Data       toolkit.M
	SessionID  string
	PipeID     string
	JobID      string
	WorkerItem PipeItem
	NodeID     string
	InputItems []string
}

func (se *workerEngine) initWorkerSession(ctx *kaos.Context, req *WorkerSessionCreateRequest) (toolkit.M, error) {
	res := toolkit.M{}

	kpCtx := kpx.NewProcessContextFromKxC(ctx)
	req.NodeID = se.worker.ID()
	ss, err := NewWorkerSession(se.worker, kpCtx, req)
	if err != nil {
		return res, fmt.Errorf("WorkerSessionCreateError: %s", err.Error())
	}
	se.sessions = append(se.sessions, ss)

	ss.Listen(se.s)
	go ss.run()

	res.Set("NodeID", se.worker.ID()).Set("SessionID", ss.ID)

	go func() {
		h, _ := ctx.DefaultHub()
		log := NewPipeLog("INFO", "", "worker session is activated")
		log.PipeID = ss.PipeID
		log.JobID = ss.JobID
		log.WorkItemID = req.WorkerItem.ID
		log.NodeID = se.worker.ID()
		log.SessionID = ss.ID
		h.Save(log)
	}()

	ctx.Log().Infof("worker %s - %s has been activated for %s - %s, job %s", se.worker.Name(), ss.ID, ss.PipeID, req.WorkerItem.ID, ss.JobID)
	return res, nil
}

type stopWorkerRequest struct {
	Kind, Value string
}

func (we *workerEngine) stopWorker(ctx *kaos.Context, req *stopWorkerRequest) ([]string, error) {
	res := []string{}
	for _, ws := range we.sessions {
		match := false
		switch req.Kind {
		case "PipeID":
			match = ws.PipeID == req.Value

		case "JobID":
			match = ws.JobID == req.Value
		}
		if match {
			res = append(res, ws.ID)
			ws.stop()
		}
	}
	return res, nil
}
