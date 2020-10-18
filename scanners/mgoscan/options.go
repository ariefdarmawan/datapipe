package mgoscan

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
)

type variable struct {
	Name         string
	Val          interface{}
	MapFromIndex int
	MapFromAttr  string
}

type options struct {
	tick time.Duration
	h    *datahub.Hub
	cmd  dbflex.ICommand

	vars map[string]*variable
}

func NewOptions(hub *datahub.Hub, tickInMs int) *options {
	opt := new(options)
	opt.h = hub
	if tickInMs == 0 {
		opt.tick = 500 * time.Millisecond
	} else {
		opt.tick = time.Duration(tickInMs) * time.Millisecond
	}
	opt.vars = make(map[string]*variable)
	return opt
}

func (o *options) SetCmd(cmd dbflex.ICommand) *options {
	o.cmd = cmd
	return o
}

func (o *options) RegisterVar(name string, initialVal interface{}, mapFromIndex int, mapFromAttr string) *options {
	o.vars[name] = &variable{name, initialVal, mapFromIndex, mapFromAttr}
	return o
}
