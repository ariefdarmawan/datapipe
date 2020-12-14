package engine

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

type coordinator struct {
	scanners map[string]*model.ScannerInfo
	workers  map[string]*model.WorkerInfo
}

func NewCoordinator() *coordinator {
	c := new(coordinator)
	c.scanners = make(map[string]*model.ScannerInfo)
	c.workers = make(map[string]*model.WorkerInfo)
	return c
}

func (c *coordinator) RegisterScanner(ctx *kaos.Context, node *model.ScannerNode) (toolkit.M, error) {
	res := toolkit.M{}

	sc, ok := c.scanners[node.ScannerID]
	if !ok {
		sc = new(model.ScannerInfo)
		sc.ID = node.ScannerID
		c.scanners[node.ScannerID] = sc
	}
	sc.RegisterNode(node)

	ctx.Log().Infof("register scanner %s node %s", node.ScannerID, node.ID)
	return res, nil
}

func (c *coordinator) DeregisterScanner(ctx *kaos.Context, node *model.ScannerNode) (toolkit.M, error) {
	res := toolkit.M{}

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
	}
	sc.RegisterNode(node)

	ctx.Log().Infof("register worker %s node %s", node.WorkerID, node.ID)
	return res, nil
}

func (c *coordinator) DeregisterWorker(ctx *kaos.Context, node *model.WorkerNode) (toolkit.M, error) {
	res := toolkit.M{}

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
		//ctx.Log().Infof("scanner %s-%s healthcheck %s", node.ScannerID, node.ID, wn.LastUpdate.String())
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
