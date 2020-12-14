package engine

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/eaciit/toolkit"
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

func (s *StorageEngine) CreateFolder(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
	res := toolkit.M{}
	h, e := ctx.DefaultHub()
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}

	kfsid := req.GetString("KfsID")
	basePath := req.GetString("Path")
	name := req.GetString("Name")
	if name == "" {
		return res, errors.New("PathEmpty")
	}

	storage := new(model.Storage)
	storage.ID = kfsid
	h.Get(storage)
	if storage.Driver == "" {
		return res, fmt.Errorf("InvalidKFSHandler: %s NotFound", kfsid)
	}

	kstorage, e := kfs.NewStorage(storage.Driver, storage.Connection, storage.Data)
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}
	if e = kstorage.Create(path.Join(basePath, name), true); e != nil {
		return res, e
	}

	return res, nil
}

func (s *StorageEngine) DeleteFiles(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
	res := toolkit.M{}
	h, e := ctx.DefaultHub()
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}

	kfsid := req.GetString("KfsID")
	basePath := req.GetString("Path")
	names := req.Get("Names", []interface{}{}).([]interface{})

	storage := new(model.Storage)
	storage.ID = kfsid
	h.Get(storage)
	if storage.Driver == "" {
		return res, fmt.Errorf("InvalidKFSHandler: %s NotFound", kfsid)
	}

	kstorage, e := kfs.NewStorage(storage.Driver, storage.Connection, storage.Data)
	if e != nil {
		return res, errors.New("InvalidKFSHandler: " + e.Error())
	}
	for _, name := range names {
		fullName := filepath.Join(basePath, name.(string))
		kstorage.Delete(fullName, true)
	}

	return res, nil
}
