package engine

import (
	"git.kanosolution.net/kano/kaos"
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

func (c *restCoordinator) Workers(ctx *kaos.Context, parm string) (toolkit.M, error) {
	res := make([]*model.WorkerInfo, len(c.c.workers))
	idx := 0
	for _, si := range c.c.workers {
		res[idx] = si
		idx++
	}
	return toolkit.M{}.Set("data", res).Set("count", len(res)), nil
}
