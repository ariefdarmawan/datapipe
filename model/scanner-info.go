package model

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type ScannerNode struct {
	ID         string
	ScannerID  string
	Secret     string
	LastUpdate time.Time
}

type ScannerInfo struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                  `bson:"_id" json:"_id" key:"1"`
	NodeCount         int                     `label:"Nodes"`
	Nodes             map[string]*ScannerNode `grid-show:"hide" form-show:"hide"`
}

func (s *ScannerInfo) RegisterNode(node *ScannerNode) {
	if s.Nodes == nil {
		s.Nodes = map[string]*ScannerNode{}
	}
	s.Nodes[node.ID] = node
	s.NodeCount = len(s.Nodes)
}

func (s *ScannerInfo) DeregisterNode(id string) {
	if s.Nodes == nil {
		s.Nodes = map[string]*ScannerNode{}
	}
	delete(s.Nodes, id)
	s.NodeCount = len(s.Nodes)
}

func (o *ScannerInfo) TableName() string {
	return "DPScanners"
}

func (o *ScannerInfo) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ScannerInfo) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}
