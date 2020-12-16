package kdp

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type workerSession struct {
	ID     string
	pipeID string
	jobID  string
	item   PipeItem
	wk     Worker
	cinput chan toolkit.M
	cstop  chan bool
	opts   *WorkerOptions
}

func NewWorkerSession(sc Worker, opts *WorkerOptions) (*workerSession, error) {
	ss := new(workerSession)
	ss.wk = sc
	ss.pipeID = opts.Data.GetString("PipeID")
	if ss.pipeID == "" {
		return nil, fmt.Errorf("InvalidPipeID")
	}
	ss.opts = opts
	ss.cinput = make(chan toolkit.M)
	ss.cstop = make(chan bool)
	ss.ID = primitive.NewObjectID().Hex()
	return ss, nil
}

func (ss *workerSession) stop() {
	ss.cstop <- true
}

func (ss *workerSession) run() {
	go func() {
		for {
			select {
			case <-ss.cstop:
				close(ss.cinput)
			case <-ss.opts.ctx.Done():
				close(ss.cinput)
			}
		}
	}()

	for m := range ss.cinput {
		mos, err := ss.wk.Work(m, ss.opts.h, ss.opts.ev)
		if err != nil {
			ss.opts.log.Errorf("%s WorkerError: %s", ss.ID[len(ss.ID)-5:], err.Error())
			continue
		}

		for mo := range mos {
			for _, r := range ss.item.Routes {
				pass := false
				if r.ConditionField == "" {
					pass = true
				} else if mo.Get(r.ConditionField, nil) == r.ConditionValue {
					pass = true
				}
				if pass {
					nextTaskID := fmt.Sprintf("/job/%s/%s/do", ss.jobID, r.RouteItemID)
					ss.opts.ev.Publish(nextTaskID, mo, nil)
				}
			}
		}
	}

	if ss.item.CloseWhenDone {
		ss.opts.ev.Publish("/job/close", toolkit.M{}.Set("ID", ss.jobID).Set("Status", "Done"), nil)
	}
}

func (ss *workerSession) Listen(s *kaos.Service) {
	ss.opts.ev.SubscribeEx(fmt.Sprintf("/job/%s/%s/do", ss.jobID, ss.item.ID), s, nil,
		func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
			res := toolkit.M{}
			ss.cinput <- req
			return res, nil
		})

	ss.opts.ev.Subscribe(fmt.Sprintf("/job/%s/%s/stop", ss.jobID, ss.item.ID), s, nil,
		func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
			res := toolkit.M{}
			ss.stop()
			return res, nil
		})
}
