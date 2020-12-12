package datapipe

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type TaskAudit struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id,omitempty" json:"_id,omitempty"`
	ProcessID         string
	Task              string
	TimeStart         time.Time
	TimeFinish        time.Time
	RunMode           string
	ClosePipe         string
	Status            string
	WorkerID          string
	Created           time.Time
}

func (p *TaskAudit) TableName() string {
	return "TaskAudits"
}

func (p *TaskAudit) GetID(c dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{p.ID}
}

func (p *TaskAudit) SetID(values ...interface{}) {
	p.ID = values[0].(string)
}
