package engine

import (
	"fmt"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
)

type restCoordinator struct {
	c *coordinator
}

func (c *restCoordinator) Scanners(ctx *kaos.Context, parm string) (toolkit.M, error) {
	res := make([]*model.ScannerInfo, len(c.c.scanners))
	idx := 0
	for _, si := range c.c.scanners {
		res[idx] = si
		idx++
	}
	return toolkit.M{}.Set("data", res).Set("count", len(res)), nil
}

func (c *restCoordinator) FindScanner(ctx *kaos.Context, parm string) ([]*model.ScannerInfo, error) {
	res := make([]*model.ScannerInfo, len(c.c.scanners))
	idx := 0
	for _, si := range c.c.scanners {
		nsi := new(model.ScannerInfo)
		*nsi = *si
		nsi.Nodes = make(map[string]*model.ScannerNode)
		nsi.NodeCount = 0
		res[idx] = nsi
		idx++
	}
	return res, nil
}

func (c *restCoordinator) Workers(ctx *kaos.Context, parm string) (toolkit.M, error) {
	res := make([]*model.WorkerInfo, len(c.c.workers))
	idx := 0
	for _, si := range c.c.workers {
		res[idx] = si
		idx++
	}
	return toolkit.M{}.Set("data", res).Set("count", len(res)), nil
}

func (c *restCoordinator) FindWorker(ctx *kaos.Context, parm string) ([]*model.WorkerInfo, error) {
	res := make([]*model.WorkerInfo, len(c.c.workers))
	idx := 0
	for _, si := range c.c.workers {
		nsi := new(model.WorkerInfo)
		*nsi = *si
		nsi.Nodes = make(map[string]*model.WorkerNode)
		nsi.NodeCount = 0
		res[idx] = nsi
		idx++
	}
	return res, nil
}

func (c *restCoordinator) StartPipe(ctx *kaos.Context, pipeID string) (toolkit.M, error) {
	res := toolkit.M{}

	h, _ := ctx.DefaultHub()
	ev, _ := ctx.DefaultEvent()

	p := new(kdp.Pipe)
	p.ID = pipeID
	e := h.Get(p)
	if e != nil {
		return res, nil
	}
	p.Running = "Running"

	scanner, ok := c.c.scanners[p.ScannerID]
	if !ok {
		return res, fmt.Errorf("InvalidScanner: %s", p.ScannerID)
	}
	if len(scanner.Nodes) == 0 {
		return res, fmt.Errorf("InvalidScanner: %s", p.ScannerID)
	}

	startM := toolkit.M{}
	if e = ev.Publish(strings.ToLower(fmt.Sprintf("/%s/create", p.ScannerID)),
		p.ScannerConfig.Set("PipeID", pipeID),
		&startM); e != nil {
		return res, fmt.Errorf("RunPipeFail: %s", e.Error())
	}

	p.ScanNodeID = startM.GetString("NodeID")
	p.ScanSessID = startM.GetString("SessionID")
	if e = h.Save(p); e != nil {
		return res, nil
	}
	ctx.Log().Infof("%s is started, scanned by %s %s", p.ID, p.ScannerID, startM.GetString("NodeID"))

	return res, nil
}

func (c *restCoordinator) StopPipe(ctx *kaos.Context, pipeID string) (toolkit.M, error) {
	res := toolkit.M{}

	h, _ := ctx.DefaultHub()
	ev, _ := ctx.DefaultEvent()

	p := new(kdp.Pipe)
	p.ID = pipeID
	e := h.Get(p)
	if e != nil {
		return res, nil
	}
	p.Running = ""

	scanner, ok := c.c.scanners[p.ScannerID]
	if !ok {
		return res, fmt.Errorf("InvalidScanner: %s", p.ScannerID)
	}
	if len(scanner.Nodes) == 0 {
		return res, fmt.Errorf("InvalidScanner: %s", p.ScannerID)
	}
	h.Save(p)

	if e = ev.Publish(strings.ToLower(fmt.Sprintf("/%s/stop", p.ScannerID)),
		p.ScannerConfig.Set("SessionID", p.ScanSessID), nil); e != nil {
		return res, fmt.Errorf("StopPipeFails: %s", e.Error())
	}
	ctx.Log().Infof("%s is stopped", p.ID)

	return res, nil
}
