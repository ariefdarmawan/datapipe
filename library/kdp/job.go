package kdp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos/kpx"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string              `bson:"_id" json:"_id" key:"1" readonly:"1"`
	PipeID            string              `readonly:"1"`
	Status            string              `readonly:"1"`
	Created           time.Time           `readonly:"1"`
	InitialData       []toolkit.M         `grid-show:"hide" readonly:"1"`
	Workers           map[string]*JobItem `grid-show:"hide"`
	Data              toolkit.M           `grid-show:"hide" form-show:"hide"`
}

func (o *Job) TableName() string {
	return "DPJobs"
}

func (o *Job) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Job) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}

func (o *Job) PreSave(conn dbflex.IConnection) error {
	if o.PipeID == "" {
		return fmt.Errorf("InvalidPipeID: %s", o.PipeID)
	}

	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
		o.Created = time.Now()
		o.Status = "New"
	}
	return nil
}

func (o *Job) Start(opts *kpx.ProcessContext) error {
	var e error

	h := opts.DataHub()
	ev := opts.Event()

	// get pipe
	p := new(Pipe)
	p.ID = o.PipeID
	if e = h.Get(p); e != nil {
		return fmt.Errorf("%s JobInvalidPipe: %s", o.ID, e.Error())
	}

	if len(p.Items) == 0 {
		return fmt.Errorf("%s JobInvalidPipe: NoWorkerItems", o.ID)
	}

	// load workers
	workers := make(map[string]*JobItem)
	for k, w := range p.Items {
		item := new(JobItem)
		item.PipeItem = w
		workers[k] = item
		inputItems := []string{}
		for _, pi := range p.Items {
			for _, r := range pi.Routes {
				if r.RouteItemID == w.ID {
					inputItems = append(inputItems, pi.ID)
				}
			}
		}

		wscr := new(WorkerSessionCreateRequest)
		/*
			translatedConfig := toolkit.M{}
			for k, v := range w.Config {
				rv := reflect.ValueOf(v)
				if rv.Kind() == reflect.String {
					translatedV := v.(string)
					translatedV, e = Translate(translatedV, opts.DataHub(), p.ID, o.ID)
					if e != nil {
						return fmt.Errorf("fail to translate %s: %s", k, v)
					}
					translatedConfig.Set(k, translatedV)
				} else {
					translatedConfig.Set(k, v)
				}
			}
		*/

		wscr.Data = w.Config
		wscr.PipeID = o.PipeID
		wscr.JobID = o.ID
		wscr.WorkerItem = w
		wscr.InputItems = inputItems

		/*
			if e = opts.DataHub().Save(item); e != nil {
				return fmt.Errorf("%s SaveWorkerSessionError: %s %s", o.ID, w.ID, e.Error())
			}
		*/
		item.ID = primitive.NewObjectID().Hex()
		wscr.SessionID = item.ID

		if w.CollectProcess {
			topic := fmt.Sprintf("/%s/assign", w.WorkerID)
			res := toolkit.M{}
			if e = ev.Publish(topic, wscr, res); e != nil {
				return fmt.Errorf("%s GetWorkerSessionError: %s %s", o.ID, w.ID, e.Error())
			}
			item.NodeID = res.GetString("NodeID")
			item.SessionID = res.GetString("SessionID")
		} else {
			topic := fmt.Sprintf("/%s/create", w.WorkerID)
			ev.Publish(topic, wscr, nil)
		}

		opts.Log().Infof("worker %s has been activated for %s - %s, job %s",
			w.WorkerID, o.PipeID, w.ID, o.ID)
	}
	o.Workers = workers
	o.Status = "Running"

	if e = h.Save(o); e != nil {
		return opts.Log().Errorf("%s JobSaveError: %s", o.ID, e.Error())
	}

	for _, item := range o.Workers {
		f := dbflex.Eqs("JobID", o.ID, "PipeItem.ID", item.PipeItem.ID)
		existingItem := new(JobItem)
		if e = h.GetByFilter(existingItem, f); e == nil {
			item.ID = existingItem.ID
		}
		item.JobID = o.ID
		h.Save(item)
	}

	// logging
	go func() {
		log := NewPipeLog("INFO", o.PipeID, "job is started")
		log.JobID = o.ID
		h.Save(log)
	}()

	go o.Fetch(opts)

	return nil
}

func (o *Job) Stop(opts *kpx.ProcessContext, status, msg string) error {
	for _, item := range o.Workers {
		stopReq := fmt.Sprintf("/%s/stop", item.ID)
		opts.Event().Publish(stopReq, toolkit.M{}, nil)
	}
	o.Status = status
	l := NewPipeLog("INFO", o.PipeID, fmt.Sprintf("Job %s is stopped with status: %s, message: %s", status, msg))
	l.JobID = o.ID
	go opts.DataHub().Save(l)
	go opts.DataHub().Save(o)
	return nil
}

type WorkerDoRequest struct {
	SourceItemID string
	Data         toolkit.M
}

func (o *Job) Fetch(opts *kpx.ProcessContext) {
	//-- get genesys worker
	genWorkers := []string{}
	for _, w := range o.Workers {
		if w.PipeItem.ReceiveScannerData {
			genWorkers = append(genWorkers, w.ID)
		}
	}

	//-- send data
	for _, data := range o.InitialData {
		for _, w := range genWorkers {
			topic := fmt.Sprintf("/%s/do", w)
			parm := new(WorkerDoRequest)
			parm.SourceItemID = "Genesys"
			parm.Data = data
			opts.Event().Publish(topic, parm, "")
		}
	}

	//-- inform worker that respective data transmit has been done
	for _, w := range genWorkers {
		topic := fmt.Sprintf("/%s/stop", o.ID, w)
		parm := new(WorkerDoRequest)
		parm.SourceItemID = w
		opts.Event().Publish(topic, parm, "")
	}
}

func (o *Job) RaiseErr(pcx *kpx.ProcessContext, msg string) error {
	pcx.Log().Error(msg)
	pl := NewPipeLog("ERROR", o.PipeID, msg)
	pl.JobID = o.ID
	go pcx.DataHub().Save(pl)
	return errors.New(msg)
}
