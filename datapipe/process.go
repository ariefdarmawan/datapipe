package datapipe

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DataProcess struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	Created           time.Time
	ScannerID         string
	ServiceID         string
	ScanProvider      string
	PipeID            string
	Status            string
	Data              []toolkit.M
}

func NewDataProcess(serviceID, scannerID, pipeID string, data []toolkit.M) *DataProcess {
	sr := new(DataProcess)
	sr.ID = primitive.NewObjectID().Hex()
	sr.ScannerID = scannerID
	sr.Data = data
	sr.ServiceID = serviceID
	sr.PipeID = pipeID
	sr.Created = time.Now()
	sr.Status = "New"
	return sr
}

func (sr *DataProcess) TableName() string {
	return "DataProcesses"
}

func (sr *DataProcess) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{sr.ID}
}

func (sr *DataProcess) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		if id, ok := keys[0].(string); ok {
			sr.ID = id
		}
	}
}
