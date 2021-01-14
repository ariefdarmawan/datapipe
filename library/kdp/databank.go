package kdp

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type DataBank struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	KeyField          string
	Description       string
	Scope             string
	RefID             string
	Name              string
	Count             int `readonly:"1"`
}

func (o *DataBank) TableName() string {
	return "DPDataBanks"
}

func (o *DataBank) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *DataBank) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}
