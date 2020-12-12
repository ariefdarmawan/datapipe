package kfsmn_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ariefdarmawan/datapipe/library/kfs"
	"github.com/eaciit/toolkit"
	cv "github.com/smartystreets/goconvey/convey"
)

func TestStorage(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	wd = filepath.Join(wd, "storage-test-data")
	os.Mkdir(wd, 0744)
	defer os.RemoveAll(wd)

	s, err := kfs.NewStorage("LocalStorage", wd, toolkit.M{}.Set("DefaultPerm", os.FileMode(0744)))
	if err != nil {
		t.Error(err)
		return
	}

	cv.Convey("list", t, func() {
		items, err := s.List("")
		cv.So(err, cv.ShouldBeNil)
		cv.So(len(items), cv.ShouldEqual, 0)

		cv.Convey("make file and dir", func() {
			e1 := s.Create("file.txt", false)
			e2 := s.Create("subfoldeer", true)
			cv.So(e1, cv.ShouldBeNil)
			cv.So(e2, cv.ShouldBeNil)

			items, err := s.List("")
			cv.So(err, cv.ShouldBeNil)
			cv.So(len(items), cv.ShouldEqual, 2)

			cv.Convey("write file", func() {
				var it *kfs.Item
				if items[1].IsDir() {
					it = items[0]
				} else {
					it = items[1]
				}
				txt := "Ini hanya test saja"
				_, err = it.Write([]byte(txt))
				cv.So(err, cv.ShouldBeNil)

				cv.Convey("read file", func() {
					data, err := it.ResetRead(false).ReadAll()
					cv.So(err, cv.ShouldBeNil)
					cv.So(len(data), cv.ShouldBeGreaterThan, 0)
					cv.So(string(data), cv.ShouldEqual, txt)
				})
			})
		})
	})
}
