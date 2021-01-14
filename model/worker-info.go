package model

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/eaciit/toolkit"
)

type WorkerNode struct {
	ID         string
	WorkerID   string
	Secret     string
	LastUpdate time.Time
}

type WorkerInfo struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                 `bson:"_id" json:"_id" key:"1" kf-pos:"1,1"`
	Description       string                 `grid-show:"hide" kf-multirow:"5" kf-pos:"2,1"`
	ConfigTemplate    toolkit.M              `grid-show:"hide" kf-pos:"3,1" kf-control:"json"`
	NodeCount         int                    `label:"Nodes" readonly:"1" form-show:"hide"`
	Nodes             map[string]*WorkerNode `grid-show:"hide" form-show:"hide"`
}

func (s *WorkerInfo) RegisterNode(node *WorkerNode) {
	if s.Nodes == nil {
		s.Nodes = map[string]*WorkerNode{}
	}
	s.Nodes[node.ID] = node
	s.NodeCount = len(s.Nodes)
}

func (s *WorkerInfo) DeregisterNode(id string) {
	if s.Nodes == nil {
		s.Nodes = map[string]*WorkerNode{}
	}
	delete(s.Nodes, id)
	s.NodeCount = len(s.Nodes)
}

func (o *WorkerInfo) TableName() string {
	return "DPWorkers"
}

func (o *WorkerInfo) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkerInfo) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}
