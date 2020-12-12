package kfs

import (
	"errors"
	"os"
	"time"

	"github.com/eaciit/toolkit"
)

type Item struct {
	storage   Storage
	readIndex int
	contents  []byte

	Path       string
	FileName   string
	FileSize   int64
	FileMode   os.FileMode
	LastUpdate time.Time
	DirFlag    bool
}

func (it *Item) Name() string {
	return it.FileName
}

func (it *Item) Size() int64 {
	return it.FileSize
}

func (it *Item) Mode() os.FileMode {
	return it.FileMode
}

func (it *Item) ModTime() time.Time {
	return it.LastUpdate
}

func (it *Item) IsDir() bool {
	return it.DirFlag
}

func (it *Item) Sys() interface{} {
	return toolkit.M{}
}

func (it *Item) ResetRead(resetContent bool) *Item {
	it.readIndex = 0
	if resetContent {
		it.contents = []byte{}
	}
	return it
}

func (it *Item) Read(data []byte) (n int, err error) {
	if it.storage == nil {
		return 0, errors.New("NilStorageObject")
	}

	if it.readIndex == 0 {
		it.contents, err = it.storage.Read(it.Path)
	}

	n = copy(data, it.contents[it.readIndex:])
	it.readIndex += n
	return n, err
}

func (it *Item) ReadAll() (content []byte, err error) {
	if it.storage == nil {
		return []byte{}, errors.New("NilStorageObject")
	}

	if len(it.contents) == 0 {
		it.contents, err = it.storage.Read(it.Path)
	}

	return it.contents, err
}

func (it *Item) Write(data []byte) (n int, err error) {
	if it.storage == nil {
		return 0, errors.New("NilStorageObject")
	}
	err = it.storage.Write(it.Path, data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (it *Item) SetStorage(s Storage) *Item {
	it.storage = s
	return it
}

func (it *Item) Storage() *Storage {
	return &it.storage
}
