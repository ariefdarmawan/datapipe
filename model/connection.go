package model

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
)

type Connection struct {
	orm.DataModelBase `json:"-" bson:"-"`
	ID                string    `bson:"_id" json:"_id" key:"true" required:"true" kf-pos:"1,1"`
	Description       string    `kf-pos:"3,1"`
	Driver            string    `kf-pos:"1,2" required:"true" kf-list:"MongoDB|Postgres|MSSQL|MySQL"`
	Connection        string    `kf-pos:"2,1" required:"true"`
	Data              toolkit.M `grid-show:"hide" kf-pos:"4,1" kf-multirow:"5"`
}

func (o *Connection) TableName() string {
	return "DPConnections"
}

func (o *Connection) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Connection) SetID(keys ...interface{}) {
	if len(keys) > 00 {
		o.ID = keys[0].(string)
	}
}
