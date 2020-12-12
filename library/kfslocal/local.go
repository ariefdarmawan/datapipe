package kfslocal

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/eaciit/toolkit"
)

type localStorage struct {
	RootPath    string
	DefaultPerm os.FileMode
}

func NewStorage(rootPath string, defaultPerm os.FileMode) (*localStorage, error) {
	fi, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("NotDirectory")
	}
	ls := new(localStorage)
	ls.RootPath = rootPath
	ls.DefaultPerm = defaultPerm
	return ls, nil
}

func init() {
	kfs.RegisterDriver("LocalStorage", func(txt string, parm toolkit.M) (kfs.Storage, error) {
		if parm == nil {
			parm = toolkit.M{}
		}
		defPerm := os.FileMode(parm.Get("DefaultPerm", 0744).(int))
		return NewStorage(txt, defPerm)
	})
}

func (ls *localStorage) Close() {
}

func (ls *localStorage) List(searchPath string) ([]*kfs.Item, error) {
	res := []*kfs.Item{}
	fullpath := filepath.Join(ls.RootPath, searchPath)
	fis, err := ioutil.ReadDir(fullpath)
	if err != nil {
		return res, err
	}
	for _, fi := range fis {
		item := new(kfs.Item)
		item.SetStorage(ls)
		item.Path = filepath.Join(searchPath, fi.Name())
		item.FileName = fi.Name()
		item.FileSize = fi.Size()
		item.FileMode = fi.Mode()
		item.DirFlag = fi.IsDir()
		item.LastUpdate = fi.ModTime()
		res = append(res, item)
	}
	return res, nil
}

func (ls *localStorage) Create(objPath string, dir bool) error {
	fullpath := filepath.Join(ls.RootPath, objPath)
	if !dir {
		if _, e := os.Create(fullpath); e != nil {
			return e
		}
		return nil
	}

	if e := os.Mkdir(fullpath, ls.DefaultPerm); e != nil {
		return e
	}
	return nil
}

func (ls *localStorage) Delete(objPath string, recursive bool) error {
	fullpath := filepath.Join(ls.RootPath, objPath)
	if recursive {
		return os.RemoveAll(fullpath)
	}

	return os.Remove(fullpath)
}

func (ls *localStorage) Read(objPath string) ([]byte, error) {
	fullpath := filepath.Join(ls.RootPath, objPath)
	return ioutil.ReadFile(fullpath)
}

func (ls *localStorage) Write(objPath string, data []byte) error {
	fullpath := filepath.Join(ls.RootPath, objPath)
	return ioutil.WriteFile(fullpath, data, ls.DefaultPerm)
}
