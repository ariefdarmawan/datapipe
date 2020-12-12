package kfs

import (
	"errors"

	"github.com/eaciit/toolkit"
)

type WriteMode string

var (
	connects = map[string]func(connTxt string, parm toolkit.M) (Storage, error){}
)

const (
	WriteTruncate WriteMode = "Truncate"
	WriteAppend             = "Append"
)

type Storage interface {
	Close()

	List(searchPath string) ([]*Item, error)
	Create(objPath string, dir bool) error
	Delete(objPath string, recursive bool) error

	Read(objPath string) ([]byte, error)
	Write(objPath string, data []byte) error
}

func NewStorage(driver, txt string, parm toolkit.M) (Storage, error) {
	fn, ok := connects[driver]
	if !ok {
		return nil, errors.New("InvalidDriver: " + driver)
	}
	return fn(txt, parm)
}

func RegisterDriver(name string, fn func(txt string, parm toolkit.M) (Storage, error)) {
	connects[name] = fn
}
