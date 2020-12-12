package kfsmn

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/eaciit/toolkit"
	"github.com/minio/minio-go"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	host, key, secret string
	mc                *minio.Client
	bucketName        string
}

func NewStorage(host, key, secret, bucket string, secure bool) (*minioStorage, error) {
	ls := new(minioStorage)
	mc, err := minio.New(host, &minio.Options{Secure: secure, Creds: credentials.NewStaticV4(key, secret, "")})
	if err != nil {
		return nil, errors.New("ConnectError: " + err.Error())
	}
	ls.mc = mc
	exist, err := mc.BucketExists(context.Background(), bucket)
	if !exist || err != nil {
		return nil, errors.New("InvalidBucket: " + bucket)
	}
	ls.bucketName = bucket
	return ls, nil
}

func init() {
	kfs.RegisterDriver("Minio", func(txt string, parm toolkit.M) (kfs.Storage, error) {
		if parm == nil {
			parm = toolkit.M{}
		}
		key := parm.GetString("Key")
		secret := parm.GetString("Secret")
		bucket := parm.GetString("Bucket")
		secure := parm.GetBool("Secure")
		return NewStorage(txt, key, secret, bucket, secure)
	})
}

func (ls *minioStorage) Close() {
	if ls.mc != nil {
		ls.mc = nil
	}
}

func (ls *minioStorage) List(searchPath string) ([]*kfs.Item, error) {
	res := []*kfs.Item{}
	cObj := ls.mc.ListObjects(context.Background(), ls.bucketName, minio.ListObjectsOptions{Recursive: false, Prefix: searchPath})

	for fi := range cObj {
		if fi.Err != nil {
			continue
		}

		item := new(kfs.Item)
		item.SetStorage(ls)
		item.Path = filepath.Join(searchPath, fi.Key)
		item.FileName = fi.Key
		item.FileSize = fi.Size
		item.FileMode = 0744
		item.DirFlag = false
		item.LastUpdate = fi.LastModified
		res = append(res, item)
	}

	return res, nil
}

func (ls *minioStorage) Create(objPath string, dir bool) error {
	fullpath := filepath.Join("", objPath)
	if !dir {
		if _, e := os.Create(fullpath); e != nil {
			return e
		}
		return nil
	}

	if e := os.Mkdir(fullpath, 0644); e != nil {
		return e
	}
	return nil
}

func (ls *minioStorage) Delete(objPath string, recursive bool) error {
	fullpath := filepath.Join("", objPath)
	if recursive {
		return os.RemoveAll(fullpath)
	}

	return os.Remove(fullpath)
}

func (ls *minioStorage) Read(objPath string) ([]byte, error) {
	fullpath := filepath.Join("", objPath)
	return ioutil.ReadFile(fullpath)
}

func (ls *minioStorage) Write(objPath string, data []byte) error {
	fullpath := filepath.Join("", objPath)
	return ioutil.WriteFile(fullpath, data, 0744)
}
