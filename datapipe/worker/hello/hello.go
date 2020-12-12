package hello

import (
	"sync"

	"git.kanosolution.net/kano/appkit"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe/datapipe"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ProviderID = "HelloWorker"
)

type worker struct {
	bt   byter.Byter
	log  *toolkit.LogEngine
	name string

	counts map[string]int
	mtx    *sync.RWMutex
}

func NewWorker(name string, opt interface{}) (datapipe.Worker, error) {
	w := new(worker)
	w.log = appkit.Log()
	w.name = primitive.NewObjectID().Hex()
	w.counts = map[string]int{}
	w.mtx = new(sync.RWMutex)
	return w, nil
}

func (w *worker) Name() string {
	return w.name
}

func (w *worker) SetByter(b byter.Byter) datapipe.Worker {
	w.bt = b
	return w
}

func (w *worker) Work(data toolkit.M) (interface{}, error) {
	msg := data.GetString("Message")
	op := data.GetString("Op")
	switch op {
	case "Reverse":
		out := ""
		for i := 0; i < len(msg); i++ {
			out += string(msg[len(msg)-1-i])
		}
		return out, nil

	case "Char":
		cs := []string{}
		for _, c := range msg {
			cs = append(cs, string(c))
		}
		return cs, nil

	case "Int":
		cs := []int{}
		for _, c := range msg {
			cs = append(cs, int(c))
		}
		return cs, nil

	case "GetCount":
		w.mtx.Lock()
		defer w.mtx.Unlock()

		processID := data.GetString("ProcessID")
		cnow := w.counts[processID]
		cnow++
		w.counts[processID] = cnow
		return cnow, nil

	default:
		return msg, nil
	}
}

func (w *worker) Collect(parm toolkit.M) (interface{}, error) {
	op := parm.GetString("Op")
	pid := parm.GetString("ProcessID")
	switch op {
	case "GetCount":
		w.mtx.Lock()
		defer w.mtx.Unlock()

		count := w.counts[pid]
		delete(w.counts, pid)
		return count, nil

	default:
		return toolkit.M{}, nil
	}
}

func (w *worker) Close() {
}
