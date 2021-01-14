package kdp

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Valve struct {
	ID       string
	Scope    string
	RefID    string
	Name     string
	KeyField string
	Data     map[string]toolkit.M
}

func (v *Valve) DataAsArray(keys []string) []toolkit.M {
	tms := []toolkit.M{}
	for k, v := range v.Data {
		if len(keys) == 0 {
			tms = append(tms, v)
		} else {
			keyFound := false
		CHECK_KEY:
			for _, key := range keys {
				if key == k {
					keyFound = true
					break CHECK_KEY
				}
			}

			if keyFound {
				tms = append(tms, v)
			}
		}
	}
	return tms
}

type DataBankBranch struct {
	valves map[string]*Valve
}

type DataBankWriteRequest struct {
	Scope     string
	RefID     string
	WriteType string
	Items     []toolkit.M
}

func DataBankWriteToTable(h *datahub.Hub, request *DataBankWriteRequest) error {
	db := new(DataBank)
	if e := h.GetByFilter(db, dbflex.Eqs("Scope", request.Scope, "RefID", request.RefID)); e != nil {
		return e
	}
	if request.WriteType == "Overwrite" {
		h.DeleteQuery(new(DataBankItem), dbflex.Eq("DataBankID", db.ID))
	}

	KeyField := db.KeyField
	for _, item := range request.Items {
		itemKey := item.GetString(KeyField)
		dbi := new(DataBankItem)
		f := dbflex.Eqs("DataBankID", db.ID, "Key", itemKey)
		h.GetByFilter(dbi, f)
		if dbi.ID == primitive.NilObjectID {
			dbi.DataBankID = db.ID
			dbi.Key = itemKey
		}
		dbi.Data = item
		if e := h.Save(dbi); e != nil {
			return fmt.Errorf("SavingError: %s %s", itemKey, e.Error())
		}
	}

	return nil
}

type DataBankReadRequest struct {
	DataBankID string
	Scope      string
	RefID      string
	Name       string
	Keys       []string
}

func DataBankReadFromTable(h *datahub.Hub, request *DataBankReadRequest) ([]toolkit.M, error) {
	items := []DataBankItem{}
	res := []toolkit.M{}

	var f *dbflex.Filter
	if request.Scope != "" && request.RefID != "" {
		f = dbflex.Eqs("Scope", request.Scope, "RefID", request.RefID, "Name", request.Name)
	} else {
		f = dbflex.Eq("Name", request.Name)
	}
	if len(request.Keys) > 0 {
		f = dbflex.And(f, dbflex.In("Key", request.Keys))
	}

	if err := h.PopulateByFilter(new(DataBankItem).TableName(), f, 0, &items); err != nil {
		return res, err
	}

	res = make([]toolkit.M, len(items))
	for k, item := range items {
		res[k] = item.Data
	}

	return res, nil
}

func (obj *DataBankBranch) ClearValve(valveID string) {
	if obj.valves == nil {
		obj.valves = make(map[string]*Valve)
	}

	delete(obj.valves, valveID)
}

func (obj *DataBankBranch) Valve(scope, refid, name string) *Valve {
	if obj.valves == nil {
		obj.valves = map[string]*Valve{}
	}

	for _, v := range obj.valves {
		if v.Scope == scope && v.RefID == refid && v.Name == name {
			return v
		}
	}

	return nil
}

func (obj *DataBankBranch) ValveItem(vid string, keys []string) []toolkit.M {
	if obj.valves == nil {
		obj.valves = map[string]*Valve{}
	}

	tms := []toolkit.M{}
	v, ok := obj.valves[vid]
	if !ok {
		return tms
	}

	for key, item := range v.Data {
		if len(keys) == 0 {
			tms = append(tms, item)
		} else {
			if toolkit.HasMember(keys, key) {
				tms = append(tms, item)
			}
		}
	}

	return tms
}

func (obj *DataBankBranch) Load(ctx *kaos.Context, request *DataBankReadRequest) ([]toolkit.M, error) {
	v := obj.Valve(request.Scope, request.RefID, request.Name)
	if v != nil {
		return v.DataAsArray(request.Keys), nil
	}

	h, _ := ctx.DefaultHub()
	ms, e := DataBankReadFromTable(h, request)
	if e != nil {
		return ms, e
	}
	v.UpdateData(ms)

	return v.DataAsArray(request.Keys), nil
}

func (v *Valve) UpdateData(items []toolkit.M) {
	if v.Data == nil {
		v.Data = make(map[string]toolkit.M)
	}

	for _, item := range items {
		keyItem := item.GetString(v.KeyField)
		v.Data[keyItem] = item
	}
}
