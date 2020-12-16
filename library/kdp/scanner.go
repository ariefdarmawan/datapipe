package kdp

import (
	"context"
	"time"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScannerOptions struct {
	Tick time.Duration
	Data toolkit.M

	ctx context.Context
	h   *datahub.Hub
	ev  kaos.EventHub
	log *toolkit.LogEngine
}

func NewScannerOptions(ctx context.Context, h *datahub.Hub, ev kaos.EventHub, log *toolkit.LogEngine, tick time.Duration, data toolkit.M) *ScannerOptions {
	so := new(ScannerOptions)
	so.ctx = ctx
	so.h = h
	so.ev = ev
	so.Tick = tick
	so.Data = data
	so.log = log
	return so
}

type Scanner interface {
	ID() string
	Name() string
	Secret() string
	SetSecret(s string) Scanner
	SetLog(l *toolkit.LogEngine) Scanner
	Log() *toolkit.LogEngine
	SetOptions(o *ScannerOptions) Scanner
	Options() *ScannerOptions
	SetDatahub(h *datahub.Hub) Scanner
	Datahub() *datahub.Hub
	SetEventHub(ev kaos.EventHub) Scanner
	EventHub() kaos.EventHub
	Scan(request toolkit.M, data *datahub.Hub, ev kaos.EventHub) ([]toolkit.M, bool, error)
}

type BaseScanner struct {
	secret string
	opts   *ScannerOptions
	h      *datahub.Hub
	ev     kaos.EventHub
	logger *toolkit.LogEngine
}

func (b *BaseScanner) ID() string {
	panic("ID is not implemented")
}

func (b *BaseScanner) Name() string {
	panic("Name is not implemented")
}

func (b *BaseScanner) Secret() string {
	if b.secret == "" {
		b.secret = primitive.NewObjectID().Hex()
	}
	return b.secret
}

func (b *BaseScanner) SetSecret(s string) Scanner {
	b.secret = s
	return b
}

func (b *BaseScanner) SetLog(l *toolkit.LogEngine) Scanner {
	b.logger = l
	return b
}

func (b *BaseScanner) Log() *toolkit.LogEngine {
	if b.logger == nil {
		b.logger = appkit.LogWithPrefix("Scanner")
	}
	return b.logger
}

func (b *BaseScanner) SetOptions(o *ScannerOptions) Scanner {
	b.opts = o
	return b
}

func (b *BaseScanner) Options() *ScannerOptions {
	if b.opts == nil {
		b.opts = new(ScannerOptions)
	}
	return b.opts
}

func (b *BaseScanner) SetDatahub(h *datahub.Hub) Scanner {
	b.h = h
	return b
}

func (b *BaseScanner) Datahub() *datahub.Hub {
	return b.h
}

func (b *BaseScanner) SetEventHub(ev kaos.EventHub) Scanner {
	b.ev = ev
	return b
}

func (b *BaseScanner) EventHub() kaos.EventHub {
	return b.ev
}

func (b *BaseScanner) Scan(request toolkit.M, h *datahub.Hub, ev kaos.EventHub) ([]toolkit.M, bool, error) {
	panic("not implemented") // TODO: Implement
}
