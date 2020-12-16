package kdp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobItem struct {
	PipeItem  PipeItem
	NodeID    string
	SessionID string
	Started   time.Time
	Completed time.Time
}

type Job struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string              `bson:"_id" json:"_id" key:"1" readonly:"1"`
	PipeID            string              `readonly:"1"`
	Status            string              `readonly:"1"`
	Created           time.Time           `readonly:"1"`
	InitialData       []toolkit.M         `grid-show:"hide" readonly:"1"`
	Workers           map[string]*JobItem `grid-show:"hide"`
	Data              toolkit.M           `grid-show:"hide"`
}

func (o *Job) TableName() string {
	return "DPJobs"
}

func (o *Job) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Job) SetID(keys ...interface{}) {
	if len(keys) > 00 {
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

func (o *Job) updateWorkerMeta(h *datahub.Hub, ev kaos.EventHub, log *toolkit.LogEngine) error {
	var e error

	// get pipe
	p := new(Pipe)
	p.ID = o.PipeID
	if e = h.Get(p); e != nil {
		return fmt.Errorf("InvalidPipe: %s", e.Error())
	}

	// load workers
	workers := make(map[string]*JobItem)
	for k, w := range p.Items {
		item := new(JobItem)
		item.PipeItem = w
		workers[k] = item

		if w.CollectProcess {
			topic := fmt.Sprintf("/%s/create", w.WorkerID)
			res := toolkit.M{}
			if e = ev.Publish(topic, toolkit.M{}.Set("JobID", o.ID), res); e != nil {
				return errors.New("GetWorerrSessionError: " + w.ID)
			}
			item.NodeID = res.GetString("NodeID")
			item.SessionID = res.GetString("SessionID")
		}
	}
	o.Workers = workers

	if e = h.Save(o); e != nil {
		return errors.New("JobSaveErr: " + o.ID + ", " + e.Error())
	}
	return nil
}
