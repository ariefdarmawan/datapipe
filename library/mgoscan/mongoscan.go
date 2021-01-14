package mgoscan

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ScannerName = "MongoScan"
)

type opts struct {
	Connection string
	OutputTo   kdp.OutputTo
	Filters    []*dbflex.Filter
	Sort       []string
	TableName  string
	Fields     []string
	Mapping    map[string]string
}

type mgoScan struct {
	kdp.BaseScanner
	id   string
	hubs map[string]*datahub.Hub
}

func NewScanner(id string) kdp.Scanner {
	ts := new(mgoScan)
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}
	ts.id = id
	ts.hubs = make(map[string]*datahub.Hub)
	return ts
}

func (o *mgoScan) ID() string {
	return o.id
}

func (o *mgoScan) Name() string {
	return ScannerName
}

func (o *mgoScan) Scan(request toolkit.M, sess *kdp.ScannerSession) ([]toolkit.M, bool, error) {
	res := []toolkit.M{}
	opt := new(opts)
	if e := toolkit.Serde(request, opt, ""); e != nil {
		return res, false, e
	}

	h, ok := o.hubs[sess.ID]
	if !ok {
		h = datahub.NewHub(datahub.GeneralDbConnBuilder(opt.Connection), true, 25)
		o.hubs[sess.ID] = h
	}

	cmd := dbflex.From(opt.TableName)
	if len(opt.Fields) > 0 {
		cmd.Select(opt.Fields...)
	} else {
		cmd.Select()
	}

	if len(opt.Filters) == 1 {
		cmd.Where(opt.Filters[0])
	} else if len(opt.Filters) > 1 {
		cmd.Where(dbflex.And(opt.Filters...))
	}

	//sess.Opts.Log().Infof("connection: %s command: %v", opt.Connection, *opt.Filters[0])
	if _, e := h.Populate(cmd, &res); e != nil {
		return res, false, e
	}

	if opt.OutputTo.Name != "" && len(res) > 0 {
		if err := kdp.MapsOutputTo(res, sess.Opts.Hub(), sess.PipeID, "", opt.OutputTo); err != nil {
			return []toolkit.M{}, true, fmt.Errorf("MapOutputErr: %s", err.Error())
		}
	}

	return res, len(res) > 0, nil
}

func (o *mgoScan) Close(id string) {
}
