package kdp

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DataBankItem struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                primitive.ObjectID `bson:"_id" json:"_id" key:"1"`
	DataBankID        string
	Key               string
	Data              toolkit.M
}

func (o *DataBankItem) TableName() string {
	return "DPDataBankItems"
}

func (o *DataBankItem) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *DataBankItem) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		var ok bool
		if o.ID, ok = keys[0].(primitive.ObjectID); !ok {
			if str, ok := keys[0].(string); ok {
				o.ID, _ = primitive.ObjectIDFromHex(str)
			} else {
				o.ID = primitive.NewObjectID()
			}
		}
	}
}
