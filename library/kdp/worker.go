package kdp

import (
	"context"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkerOptions struct {
	Data toolkit.M

	ctx context.Context
	h   *datahub.Hub
	ev  kaos.EventHub
	log *toolkit.LogEngine
}

func (w *WorkerOptions) Context() context.Context {
	return w.ctx
}

func (w *WorkerOptions) Hub() *datahub.Hub {
	return w.h
}

func (w *WorkerOptions) Event() kaos.EventHub {
	return w.ev
}

func (w *WorkerOptions) Log() *toolkit.LogEngine {
	return w.log
}

func NewWorkerOptions(ctx context.Context, h *datahub.Hub, ev kaos.EventHub, log *toolkit.LogEngine, data toolkit.M) *WorkerOptions {
	so := new(WorkerOptions)
	so.ctx = ctx
	so.h = h
	so.ev = ev
	so.Data = data
	so.log = log
	return so
}

type Worker interface {
	ID() string
	Name() string
	Secret() string
	SetSecret(s string) Worker
	SetLog(l *toolkit.LogEngine) Worker
	Log() *toolkit.LogEngine
	SetOptions(o *WorkerOptions) Worker
	Options() *WorkerOptions
	SetDatahub(h *datahub.Hub) Worker
	Datahub() *datahub.Hub
	SetEventHub(ev kaos.EventHub) Worker
	EventHub() kaos.EventHub
	Work(request toolkit.M, data *WorkerSession, inputName string, channelResult chan<- toolkit.M) error
	Collect(channelResult chan<- toolkit.M) error
	Close(jobID string)
}

type BaseWorker struct {
	secret string
	opts   *WorkerOptions
	h      *datahub.Hub
	ev     kaos.EventHub
	logger *toolkit.LogEngine
}

func (b *BaseWorker) ID() string {
	panic("ID is not implemented")
}

func (b *BaseWorker) Name() string {
	panic("Name is not implemented")
}

func (b *BaseWorker) Secret() string {
	if b.secret == "" {
		b.secret = primitive.NewObjectID().Hex()
	}
	return b.secret
}

func (b *BaseWorker) SetSecret(s string) Worker {
	b.secret = s
	return b
}

func (b *BaseWorker) SetLog(l *toolkit.LogEngine) Worker {
	b.logger = l
	return b
}

func (b *BaseWorker) Log() *toolkit.LogEngine {
	if b.logger == nil {
		b.logger = appkit.LogWithPrefix("Worker")
	}
	return b.logger
}

func (b *BaseWorker) SetOptions(o *WorkerOptions) Worker {
	b.opts = o
	return b
}

func (b *BaseWorker) Options() *WorkerOptions {
	if b.opts == nil {
		b.opts = new(WorkerOptions)
	}
	return b.opts
}

func (b *BaseWorker) SetDatahub(h *datahub.Hub) Worker {
	b.h = h
	return b
}

func (b *BaseWorker) Datahub() *datahub.Hub {
	return b.h
}

func (b *BaseWorker) SetEventHub(ev kaos.EventHub) Worker {
	b.ev = ev
	return b
}

func (b *BaseWorker) EventHub() kaos.EventHub {
	return b.ev
}

func (b *BaseWorker) Work(request toolkit.M, ws *WorkerSession, inputName string, res chan<- toolkit.M) error {
	panic("worker.Work is not implemented") // TODO: Implement
}

func (b *BaseWorker) Collect(chan<- toolkit.M) error {
	panic("worker.Collect is not implemented")
}

func (b *BaseWorker) Close(jobID string) {
	panic("not implemented") // TODO: Implement
}
