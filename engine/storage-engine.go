package engine

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/ariefdarmawan/datapipe/model"
)

type StorageEngine struct {
}

type ExploreRequest struct {
	KfsID string
	Path  string
}

func (s *StorageEngine) Explore(ctx *kaos.Context, req *ExploreRequest) ([]*kfs.Item, error) {
	res := []*kfs.Item{}

	h, e := ctx.DefaultHub()
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}

	storage := new(model.Storage)
	storage.ID = req.KfsID
	h.Get(storage)
	if storage.Driver == "" {
		return res, fmt.Errorf("InvalidKFSHandler: %s NotFound", req.KfsID)
	}

	kstorage, e := kfs.NewStorage(storage.Driver, storage.Connection, storage.Data)
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}

	return kstorage.List(req.Path)
}
