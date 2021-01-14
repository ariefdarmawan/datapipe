package model

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
)

type Storage struct {
	orm.DataModelBase `json:"-" bson:"-"`
	ID                string    `bson:"_id" json:"_id" key:"true" required:"true" kf-pos:"1,1"`
	Description       string    `kf-pos:"3,1"`
	Driver            string    `kf-pos:"1,2" required:"true" kf-list:"LocalStorage|Minio|AWS S3|HDFS"`
	Connection        string    `kf-pos:"2,1" required:"true"`
	Data              toolkit.M `grid-show:"hide" kf-pos:"4,1" kf-control:"json"`
}

func (o *Storage) TableName() string {
	return "DPStorages"
}

func (o *Storage) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Storage) SetID(keys ...interface{}) {
	if len(keys) > 00 {
		o.ID = keys[0].(string)
	}
}
