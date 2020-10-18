package datapipe

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScanResult struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	Created           time.Time
	ScannerID         string
	ServiceID         string
	WorkerID          string
	ScanProvider      string
	WorkerProvider    string
	Status            string
	Data              []toolkit.M
}

func NewScanResult(serviceID, scannerID string, data []toolkit.M) *ScanResult {
	sr := new(ScanResult)
	sr.ID = primitive.NewObjectID().Hex()
	sr.ScannerID = scannerID
	sr.Data = data
	sr.ServiceID = serviceID
	sr.Created = time.Now()
	sr.Status = "New"
	return sr
}

func (sr *ScanResult) TableName() string {
	return "ScanResults"
}

func (sr *ScanResult) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{sr.ID}
}

func (sr *ScanResult) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		if id, ok := keys[0].(string); ok {
			sr.ID = id
		}
	}
}
