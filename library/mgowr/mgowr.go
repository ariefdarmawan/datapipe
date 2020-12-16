package mgowr

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
	"github.com/sugab/datahub"
)

const WorkerID = "MongoWriter"

type writer struct {
	kdp.BaseWorker
	id string
}

func (w *writer) ID() string {
	return w.id
}

func (w *writer) Name() string {
	return WorkerID
}

func (w *writer) Work(request toolkit.M, h *datahub.Hub, ev kaos.EventHub) ([]toolkit.M, error) {
	res := []toolkit.M{}

	tableName := request.GetString("TableName")
	if tableName == "" {
		return res, fmt.Errorf("InvalidAttribute: %s = %s", "TableName", tableName)
	}

	data, ok := request["Data"]
	if !ok {
		return res, errors.New("NoData")
	}
	if e := h.SaveAny(tableName, data); e != nil {
		return res, fmt.Errorf("DBErr: %s", e.Error())
	}

	return res, nil
}
