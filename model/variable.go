package model

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type Variable struct {
	orm.DataModelBase `json:"-" bson:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" required:"true"`
	Kind              string `kf-list:"Text|Number|Date"`
	Value             string
}

func (o *Variable) TableName() string {
	return "DPVars"
}

func (o *Variable) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Variable) SetID(keys ...interface{}) {
	if len(keys) > 00 {
		o.ID = keys[0].(string)
	}
}
