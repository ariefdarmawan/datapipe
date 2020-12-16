package kdp

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
)

type Pipe struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string              `bson:"_id" json:"_id" key:"1" kf-pos:"1,1"`
	Name              string              ` kf-pos:"1,2"`
	ScannerID         string              `kf-pos:"2,1" kf-lookup:"/coordinator/findscanner|_id|_id" kf-allow-add:"1"`
	Status            string              `kf-pos:"2,2" kf-list:"Inactive|Active"`
	Running           string              `readonly:"1" form-show:"hide"`
	ScannerConfig     toolkit.M           `grid-show:"hide" kf-pos:"3,1"`
	ScanNodeID        string              `readonly:"1" form-show:"hide"`
	ScanSessID        string              `readonly:"1" form-show:"hide"`
	Items             map[string]PipeItem `grid-show:"hide"  kf-pos:"4,1"`
}

func (o *Pipe) TableName() string {
	return "DPPipes"
}

func (o *Pipe) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Pipe) SetID(keys ...interface{}) {
	if len(keys) > 00 {
		o.ID = keys[0].(string)
	}
}
