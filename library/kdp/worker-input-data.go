package kdp

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkerInputData struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	JobID             string
	WorkItemID        string
	NodeID            string
	SessionID         string
	Status            string
	Data              toolkit.M
	Message           string
	Created           time.Time
	Processed         time.Time
}

func NewWorkerInputData(jobid, workItemID, nodeID, sessionID string) *WorkerInputData {
	wid := new(WorkerInputData)
	wid.ID = primitive.NewObjectID().Hex()
	wid.JobID = jobid
	wid.WorkItemID = workItemID
	wid.NodeID = nodeID
	wid.SessionID = sessionID
	return wid
}

func (o *WorkerInputData) TableName() string {
	return "DPWorkerAudits"
}

func (o *WorkerInputData) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkerInputData) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}

func (o *WorkerInputData) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
		o.Created = time.Now()
		o.Status = "New"
	}
	return nil
}
