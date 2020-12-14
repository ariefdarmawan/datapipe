package kfsmn

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/eaciit/toolkit"
	"github.com/minio/minio-go/v7"
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
	if searchPath != "" && !strings.HasSuffix(searchPath, "/") {
		searchPath += "/"
	}

	res := []*kfs.Item{}
	cObj := ls.mc.ListObjects(context.Background(), ls.bucketName, minio.ListObjectsOptions{Recursive: false, Prefix: searchPath})

	for fi := range cObj {
		if fi.Err != nil {
			continue
		}

		isFolder := fi.Key[len(fi.Key)-1] == '/'
		fullName := fi.Key
		if isFolder {
			fullName = fi.Key[:len(fi.Key)-1]
		}
		names := strings.Split(fullName, "/")
		name := names[len(names)-1]
		item := new(kfs.Item)
		item.SetStorage(ls)
		item.Path = filepath.Join(searchPath, name)
		item.FileName = name
		item.FileSize = fi.Size
		item.FileMode = 0744
		item.DirFlag = isFolder
		item.LastUpdate = fi.LastModified
		res = append(res, item)
	}

	return res, nil
}

func (ls *minioStorage) Create(objPath string, dir bool) error {
	if dir && !strings.HasSuffix(objPath, "/") {
		objPath += "/"
	}

	var e error
	if !dir {
		bs := []byte("")
		rdr := bytes.NewReader(bs)
		_, e = ls.mc.PutObject(context.Background(), ls.bucketName, objPath, rdr, 0, minio.PutObjectOptions{})
	} else {
		_, e = ls.mc.PutObject(context.Background(), ls.bucketName, objPath, nil, 0, minio.PutObjectOptions{})
	}
	if e != nil {
		return e
	}
	return nil
}

func (ls *minioStorage) Delete(objPath string, recursive bool) error {
	return ls.mc.RemoveObject(context.Background(), ls.bucketName, objPath, minio.RemoveObjectOptions{})
}

func (ls *minioStorage) Read(objPath string) ([]byte, error) {
	fullpath := filepath.Join("", objPath)
	return ioutil.ReadFile(fullpath)
}

func (ls *minioStorage) Write(objPath string, data []byte) error {
	fullpath := filepath.Join("", objPath)
	return ioutil.WriteFile(fullpath, data, 0744)
}
