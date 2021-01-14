package kdp

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PipeLog struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" grid-show:"hide"`
	PipeID            string
	ScannerID         string `grid-show:"hide"`
	JobID             string
	WorkItemID        string
	WorkerID          string `grid-show:"hide"`
	NodeID            string `grid-show:"hide"`
	SessionID         string `grid-show:"hide"`
	LogType           string `label:"Log Type"`
	Message           string
	Created           time.Time
}

func (o *PipeLog) TableName() string {
	return "DPPipeLogs"
}

func (o *PipeLog) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PipeLog) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}

func (o *PipeLog) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
		o.Created = time.Now()
	}
	if o.LogType == "" {
		o.LogType = "INFO"
	}
	return nil
}

func NewPipeLog(logType, pipeID, message string) *PipeLog {
	p := new(PipeLog)
	p.LogType = logType
	p.PipeID = pipeID
	p.Message = message
	return p
}
