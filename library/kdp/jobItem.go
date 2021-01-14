package kdp

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobItem struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	PipeItem          PipeItem
	Genesys           bool
	NodeID            string
	SessionID         string
	JobID             string
	Started           time.Time
	Completed         time.Time
}

func (o *JobItem) TableName() string {
	return "DPJobItems"
}

func (o *JobItem) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *JobItem) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}

func (o *JobItem) PreSave(_ dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}
