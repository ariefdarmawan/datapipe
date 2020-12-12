package mgoscan

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ProviderID = "MongoScan"

type mgoScan struct {
	name, service, topic string
	bt                   byter.Byter
	opts                 *options

	timeTracker time.Time
}

func NewMgoScan(service, topic string, mOpts toolkit.M) (datapipe.Scanner, error) {
	opts := new(options)
	if err := toolkit.Serde(mOpts, opts, ""); err != nil {
		return nil, fmt.Errorf("unable to serialize options: %s", err.Error())
	}
	scan := new(mgoScan)
	scan.service = service
	scan.name = primitive.NewObjectID().Hex()
	scan.timeTracker = time.Now()
	scan.topic = topic
	return scan, nil
}

func (ms *mgoScan) Name() string {
	return ms.name
}

func (ms *mgoScan) Service() string {
	return ms.service
}

func (ms *mgoScan) SetByter(b byter.Byter) datapipe.Scanner {
	ms.bt = b
	return ms
}

func (ms *mgoScan) Scan() (triggered bool, data []toolkit.M, err error) {
	tn := time.Now()
	if tn.Sub(ms.timeTracker) >= ms.opts.tick {
		ms.timeTracker = tn
		res := []toolkit.M{}
		//fmt.Println("cmd pre:", toolkit.JsonString(ms.opts.cmd.Items()[dbflex.QueryWhere]))
		cmd := ms.pushVarToCmd(ms.opts.cmd)
		//fmt.Println("cmd upd:", toolkit.JsonString(cmd.Items()[dbflex.QueryWhere]))
		if _, e := ms.opts.h.Populate(cmd, &res); e != nil {
			return false, nil, e
		}
		if len(res) > 0 {
			for vidx, v := range ms.opts.vars {
				if v.MapFromIndex >= 0 && len(res) > v.MapFromIndex {
					v.Val, _ = res[v.MapFromIndex].PathGet(v.MapFromAttr)
				} else if v.MapFromIndex < 0 && len(res) >= -v.MapFromIndex {
					idx := len(res) + v.MapFromIndex
					v.Val, _ = res[idx].PathGet(v.MapFromAttr)
				}
				ms.opts.vars[vidx] = v
			}
			//fmt.Println("res len:", len(res), "vars:", toolkit.JsonString(ms.opts.vars), res[len(res)-1])
			return true, res, nil
		}
	}
	return false, nil, nil
}

func (ms *mgoScan) pushVarToCmd(cmd dbflex.ICommand) dbflex.ICommand {
	newItems := dbflex.QueryItems{}
	for _, item := range cmd.Items() {
		newItem := item
		if newItem.Op == dbflex.QueryWhere {
			if filter, ok := newItem.Value.(*dbflex.Filter); ok {
				newFilter := new(dbflex.Filter)
				*newFilter = *filter
				ms.pushVarToFilter(newFilter)
				newItem.Value = newFilter
				//fmt.Println("flt update:", toolkit.JsonString(item))
			}
		}
		newItems[item.Op] = newItem
	}
	nc := &dbflex.CommandBase{}
	nc.SetItems(newItems)
	return nc
}

func (ms *mgoScan) pushVarToFilter(filter *dbflex.Filter) {
	if filter.Op == dbflex.OpAnd || filter.Op == dbflex.OpOr {
		filterItems := filter.Items
		newItems := []*dbflex.Filter{}
		for _, filterItem := range filterItems {
			newFilterItem := new(dbflex.Filter)
			*newFilterItem = *filterItem
			ms.pushVarToFilter(newFilterItem)
			newItems = append(newItems, newFilterItem)
		}
		filter.Items = newItems
	} else {
		if vs, ok := filter.Value.(string); ok {
			if len(vs) > 2 && vs[0] == '%' {
				vs = vs[1:]
				msvar, found := ms.opts.vars[vs]
				if found {
					//fmt.Println("filter psh:", vs, msvar.Val)
					filter.Value = msvar.Val
				}
			}
		}
	}
}

func (ms *mgoScan) MakePayload(data interface{}) (interface{}, error) {
	return nil, nil
}

func (ms *mgoScan) Close() {
}

func (ms *mgoScan) SetTopic(name string) datapipe.Scanner {
	ms.topic = name
	return ms
}

func (ms *mgoScan) Topic() string {
	if ms.topic == "" {
		ms.topic = primitive.NewObjectID().Hex()
	}
	return ms.topic
}
