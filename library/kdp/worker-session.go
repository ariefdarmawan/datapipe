package kdp

import (
	"fmt"
	"sync"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kaos/kpx"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type workerInput struct {
	done bool
	c    chan *WorkerInputData
}

type WorkerSession struct {
	ID         string
	JobID      string
	PipeID     string
	NodeID     string
	Data       toolkit.M
	WorkerItem PipeItem

	wk Worker

	cstop  chan bool
	inputs map[string]*workerInput
	opts   *kpx.ProcessContext
}

func NewWorkerSession(worker Worker, opts *kpx.ProcessContext, req *WorkerSessionCreateRequest) (*WorkerSession, error) {
	ss := new(WorkerSession)
	ss.wk = worker

	if req.SessionID == "" {
		req.SessionID = primitive.NewObjectID().Hex()
	}
	ss.ID = req.SessionID
	ss.Data = req.Data
	ss.PipeID = req.PipeID
	if ss.PipeID == "" {
		return nil, fmt.Errorf("InvalidPipeID")
	}
	ss.JobID = req.JobID
	ss.NodeID = req.NodeID
	ss.WorkerItem = req.WorkerItem

	ss.opts = opts
	ss.cstop = make(chan bool)
	ss.inputs = map[string]*workerInput{"Genesys": &workerInput{c: make(chan *WorkerInputData)}}
	return ss, nil
}

func (ss *WorkerSession) stop() {
	ss.opts.Log().Infof("%s %s %s sess: %s is going to be stopped", ss.PipeID, ss.JobID, ss.WorkerItem.ID, ss.ID)
	ss.cstop <- true
}

func (ss *WorkerSession) run() {
	// routine to close
	go func() {
		for {
			select {
			case <-ss.cstop:
				ss.doStop()
			case <-ss.opts.Context().Done():
				ss.doStop()
			}
		}
	}()

	cres := make(chan toolkit.M)
	// routine to do the work
	go func() {
		wg := new(sync.WaitGroup)
		wg.Add(len(ss.inputs))

		for nm, input := range ss.inputs {
			go func(nm string, input *workerInput, wg *sync.WaitGroup) {
				defer wg.Done()

				for mInput := range input.c {
					err := ss.wk.Work(mInput.Data, ss, nm, cres)
					if err != nil {
						ss.opts.Log().Errorf("%s WorkerError: %s", ss.ID[len(ss.ID)-5:], err.Error())
						mInput.Status = "Error"
						mInput.Message = err.Error()
						mInput.Processed = time.Now()
						go ss.opts.DataHub().Save(mInput)

						ss.opts.Event().Publish("/coordinator/stopjob",
							toolkit.M{}.Set("JobID", ss.JobID).Set("Status", "Stopped"),
							nil)
						continue
					}

					mInput.Status = "Processed"
					mInput.Processed = time.Now()
					go ss.opts.DataHub().Save(mInput)
				}

				input.done = true
			}(nm, input, wg)
		}

		wg.Wait()
		if ss.WorkerItem.CollectProcess {
			ss.wk.Collect(cres)
		}

		if ss.WorkerItem.CloseWhenDone {
			ss.opts.Event().Publish("/coordinator/stopjob",
				toolkit.M{}.Set("JobID", ss.JobID).Set("Status", "Done"),
				nil)
		}

		// inform next route that input has been completed
		for _, route := range ss.WorkerItem.Routes {
			ss.opts.Event().Publish(fmt.Sprintf("/worker/%s/stop", ss.JobID),
				toolkit.M{}.Set("WorkerItemID", route.RouteItemID),
				nil)
		}
	}()

	// retrieve the work
	for mo := range cres {
		for _, r := range ss.WorkerItem.Routes {
			pass := false
			if r.ConditionField == "" {
				pass = true
			} else if mo.Get(r.ConditionField, nil) == r.ConditionValue {
				pass = true
			}
			if pass {
				nextTaskID := fmt.Sprintf("/job/%s/%s/do", ss.JobID, r.RouteItemID)
				ss.opts.Event().Publish(nextTaskID, mo, nil)
			}
		}
	}
}

func (ss *WorkerSession) Listen(s *kaos.Service) {
	// listener to retrieve input
	//ss.opts.Event().SubscribeEx(fmt.Sprintf("/%s/%s/do", ss.JobID, ss.WorkerItem.ID), s, nil,
	ss.opts.Event().SubscribeEx(fmt.Sprintf("/%s/do", ss.ID), s, nil,
		func(ctx *kaos.Context, req *WorkerDoRequest) (toolkit.M, error) {
			if req.SourceItemID == "" {
				req.SourceItemID = "Genesys"
			}

			ss.sendInput(req.SourceItemID, req.Data)

			// record log
			go func() {
				h, _ := ctx.DefaultHub()
				log := NewPipeLog("INFO", ss.PipeID,
					fmt.Sprintf("jobid: %s step: %s workerid: %s do", ss.JobID, ss.WorkerItem.ID, ss.WorkerItem.WorkerID))
				log.JobID = ss.JobID
				log.WorkerID = ss.WorkerItem.WorkerID
				log.WorkItemID = ss.WorkerItem.ID
				log.NodeID = ss.NodeID
				log.SessionID = ss.ID
				h.Save(log)
			}()

			return toolkit.M{}, nil
		})

	// listener to stop session
	//ss.opts.Event().Subscribe(fmt.Sprintf("/%s/%s/stop", ss.JobID, ss.WorkerItem.ID), s, nil,
	ss.opts.Event().Subscribe(fmt.Sprintf("/%s/stop", ss.ID), s, nil,
		func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
			res := toolkit.M{}

			inputID := req.GetString("InputID")
			if inputID == "" {
				ss.stop()

				// record log
				go func() {
					h, _ := ctx.DefaultHub()
					log := NewPipeLog("INFO", ss.PipeID,
						fmt.Sprintf("jobid: %s step: %s workerid: %s stop", ss.JobID, ss.WorkerItem.ID, ss.WorkerItem.WorkerID))
					log.JobID = ss.JobID
					log.WorkerID = ss.WorkerItem.WorkerID
					log.WorkItemID = ss.WorkerItem.ID
					log.NodeID = ss.NodeID
					log.SessionID = ss.ID
					h.Save(log)
				}()
			} else {
				ss.closeInput(inputID)
			}

			return res, nil
		})
}

func (ss *WorkerSession) doStop() {
	for inputName := range ss.inputs {
		ss.closeInput(inputName)
	}

	if ss.WorkerItem.CloseWhenDone {
		ss.opts.Event().Publish("/coordinator/closejob", toolkit.M{}.Set("ID", ss.JobID).Set("Status", "Done"), nil)
	}
}

func (ss *WorkerSession) closeInput(name string) {
	inp, ok := ss.inputs[name]
	if !ok {
		return
	}

	if !inp.done {
		inp.done = true
		close(inp.c)
	}

	// record log
	go func() {
		log := NewPipeLog("INFO", ss.PipeID, "stop input channel "+name)
		log.JobID = ss.JobID
		log.WorkItemID = ss.WorkerItem.ID
		log.NodeID = ss.NodeID
		log.SessionID = ss.ID
		ss.opts.DataHub().Save(log)
	}()
}

func (ss *WorkerSession) sendInput(name string, input toolkit.M) {
	inp, ok := ss.inputs[name]
	if !ok {
		return
	}

	if inp.done {
		return
	}

	inputData := NewWorkerInputData(ss.JobID, ss.WorkerItem.ID, ss.NodeID, ss.ID)
	inputData.Data = input
	go ss.opts.DataHub().Save(inputData)
	inp.c <- inputData
}
