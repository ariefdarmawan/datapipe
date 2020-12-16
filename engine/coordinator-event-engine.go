package engine

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

type coordinator struct {
	scanners map[string]*model.ScannerInfo
	workers  map[string]*model.WorkerInfo
}

func NewCoordinator(h *datahub.Hub) *coordinator {
	c := new(coordinator)
	c.scanners = make(map[string]*model.ScannerInfo)
	c.workers = make(map[string]*model.WorkerInfo)

	scanners := []*model.ScannerInfo{}
	h.Gets(new(model.ScannerInfo), nil, &scanners)
	for _, sc := range scanners {
		sc.Nodes = make(map[string]*model.ScannerNode)
		sc.NodeCount = 0
		c.scanners[sc.ID] = sc
	}

	workers := []*model.WorkerInfo{}
	h.Gets(new(model.WorkerInfo), nil, &workers)
	for _, sc := range workers {
		sc.Nodes = make(map[string]*model.WorkerNode)
		sc.NodeCount = 0
		c.workers[sc.ID] = sc
	}
	return c
}

func (c *coordinator) RegisterScanner(ctx *kaos.Context, node *model.ScannerNode) (toolkit.M, error) {
	res := toolkit.M{}

	sc, ok := c.scanners[node.ScannerID]
	if !ok {
		sc = new(model.ScannerInfo)
		sc.ID = node.ScannerID
		c.scanners[node.ScannerID] = sc

		h, _ := ctx.DefaultHub()
		if h != nil {
			h.Save(sc)
		}
	}

	if node.ID != "" {
		sc.RegisterNode(node)
	}

	ctx.Log().Infof("register scanner %s node %s", node.ScannerID, node.ID)
	return res, nil
}

func (c *coordinator) DeregisterScanner(ctx *kaos.Context, node *model.ScannerNode) (toolkit.M, error) {
	res := toolkit.M{}

	if node.ID == "" {
		delete(c.scanners, node.ScannerID)
		return res, nil
	}

	sc, ok := c.scanners[node.ScannerID]
	if !ok {
		sc = new(model.ScannerInfo)
		sc.ID = node.ScannerID
		c.scanners[node.ScannerID] = sc
	}

	sc.DeregisterNode(node.ID)

	ctx.Log().Infof("deregister scanner %s node %s", node.ScannerID, node.ID)
	return res, nil
}

func (c *coordinator) RegisterWorker(ctx *kaos.Context, node *model.WorkerNode) (toolkit.M, error) {
	res := toolkit.M{}

	sc, ok := c.workers[node.WorkerID]
	if !ok {
		sc = new(model.WorkerInfo)
		sc.ID = node.WorkerID
		c.workers[node.WorkerID] = sc

		h, _ := ctx.DefaultHub()
		if h != nil {
			h.Save(sc)
		}
	}

	if node.ID != "" {
		sc.RegisterNode(node)
	}

	ctx.Log().Infof("register worker %s node %s", node.WorkerID, node.ID)
	return res, nil
}

func (c *coordinator) DeregisterWorker(ctx *kaos.Context, node *model.WorkerNode) (toolkit.M, error) {
	res := toolkit.M{}

	if node.ID == "" {
		delete(c.workers, node.WorkerID)
		return res, nil
	}

	sc, ok := c.workers[node.WorkerID]
	if !ok {
		sc = new(model.WorkerInfo)
		sc.ID = node.WorkerID
		c.workers[node.WorkerID] = sc
	}
	sc.DeregisterNode(node.ID)

	ctx.Log().Infof("Deregister worker %s node %s", node.WorkerID, node.ID)
	return res, nil
}

func (c *coordinator) ScannerBeat(ctx *kaos.Context, node *model.ScannerNode) (toolkit.M, error) {
	res := toolkit.M{}

	sc, ok := c.scanners[node.ScannerID]
	if !ok {
		return res, nil
	}
	if wn, ok := sc.Nodes[node.ID]; ok {
		wn.LastUpdate = time.Now()
	} else {
		sc.Nodes[node.ID] = node
		node.LastUpdate = time.Now()
	}

	for _, node := range sc.Nodes {
		if time.Now().Sub(node.LastUpdate) > (15 * time.Minute) {
			node.Status = "Error"
		} else {
			node.Status = "OK"
		}
	}
	return res, nil
}

func (c *coordinator) WorkerBeat(ctx *kaos.Context, node *model.WorkerNode) (toolkit.M, error) {
	res := toolkit.M{}

	sc, ok := c.workers[node.WorkerID]
	if !ok {
		return res, nil
	}
	if wn, ok := sc.Nodes[node.ID]; ok {
		wn.LastUpdate = time.Now()
		//ctx.Log().Infof("worker %s-%s healthcheck %s", node.WorkerID, node.ID, wn.LastUpdate.String())
	}
	return res, nil
}

func (c *coordinator) CloseNodes(ev kaos.EventHub) {
	ev.Publish("/node/close", "", nil)
}

func (c *coordinator) RESTEngine() *restCoordinator {
	rc := new(restCoordinator)
	rc.c = c
	return rc
}
