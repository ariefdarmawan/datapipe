package mgoworker

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ProviderID = "MgoWorker"

type worker struct {
	opts *options
	bt   byter.Byter
	name string
}

func NewWorker(name string, opt interface{}) (datapipe.Worker, error) {
	opts, ok := opt.(*options)
	if !ok {
		return nil, fmt.Errorf("options is mandatory")
	}

	w := new(worker)
	w.opts = opts
	w.name = primitive.NewObjectID().Hex()
	w.bt = byter.NewByter("")
	return w, nil
}

func (mw *worker) Name() string {
	return mw.name
}

func (mw *worker) SetByter(b byter.Byter) datapipe.Worker {
	mw.bt = b
	return mw
}

func (mw *worker) Work(data toolkit.M) (interface{}, error) {
	results := []toolkit.M{}
	cmd := dbflex.CopyCommand(mw.opts.cmd)
	varData := mw.opts.VarData(data)
	dbflex.PushVarToCommand(cmd, varData)
	if _, e := mw.opts.hr.Populate(cmd, &results); e != nil {
		return nil, fmt.Errorf("fail run worker. %s", e.Error())
	}

	for _, res := range results {
		cmdw := dbflex.From(mw.opts.writeTableName).Save()
		mw.opts.hw.Execute(cmdw, res)
	}
	return len(results), nil
}

func (w *worker) Collect(parm toolkit.M) (interface{}, error) {
	return toolkit.M{}, nil
}

func (w *worker) Close() {
}
