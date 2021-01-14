package mgowr

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const WorkerName = "MongoWriter"

type writer struct {
	kdp.BaseWorker
	id string

	conns map[string]*datahub.Hub
}

func NewEngine(id string) *writer {
	w := new(writer)
	w.id = id
	w.conns = make(map[string]*datahub.Hub)
	return w
}

func (w *writer) ID() string {
	return w.id
}

func (w *writer) Name() string {
	return WorkerName
}

func (w *writer) Work(request toolkit.M, ws *kdp.WorkerSession, inputName string, cres chan<- toolkit.M) error {
	res := make(chan toolkit.M)
	defer close(res)

	tableName := ws.Data.GetString("TableName")
	if tableName == "" {
		return fmt.Errorf("InvalidAttribute: %s = %s", "TableName", tableName)
	}

	mapToID := ws.Data.GetString("MapToID")
	if mapToID != "" {
		request.Set("_id", request[mapToID])
	}

	conn, ok := w.conns[ws.ID]
	if !ok {
		if !ws.Data.Has("DBConn") {
			return errors.New("NoDBConn")
		}
		dbConnTxt := ws.Data.GetString("DBConn")
		conn = datahub.NewHub(datahub.GeneralDbConnBuilder(dbConnTxt), true, 10)
	}

	if !request.Has("_id") {
		request.Set("_id", primitive.NewObjectID().Hex())
	}
	cmd := dbflex.From(tableName).Save()
	if _, e := conn.Execute(cmd, request); e != nil {
		return fmt.Errorf("DBError: %s", e.Error())
	}

	return nil
}

func (w *writer) Close(jobID string) {
	for _, conn := range w.conns {
		conn.Close()
	}
}
