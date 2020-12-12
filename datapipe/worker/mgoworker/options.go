package mgoworker

import (
	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
)

type options struct {
	cmd            dbflex.ICommand
	hr             *datahub.Hub
	hw             *datahub.Hub
	writeTableName string
	vars           map[string]string
}

func NewOptions(cmd dbflex.ICommand, hr, hw *datahub.Hub, writeTableName string) *options {
	o := new(options)
	o.cmd = cmd
	o.hr = hr
	o.hw = hw
	o.vars = map[string]string{}
	o.writeTableName = writeTableName
	return o
}

func (o *options) RegisterVar(name string, mapFieldId string) *options {
	o.vars[name] = mapFieldId
	return o
}

func (o *options) VarData(data toolkit.M) toolkit.M {
	m := toolkit.M{}
	for k, v := range o.vars {
		datav, _ := data.PathGet(v)
		m.Set(k, datav)
	}
	return m
}
