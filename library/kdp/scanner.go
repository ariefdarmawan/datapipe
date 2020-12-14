package kdp

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScannerOptions struct {
	TickDuration time.Duration
	Data         toolkit.M
}

type Scanner interface {
	ID() string
	Secret() string
	SetSecret(s string) Scanner
	SetOptions(o *ScannerOptions) Scanner
	Options() *ScannerOptions
	SetDatahub(h *datahub.Hub) Scanner
	Datahub() *datahub.Hub
	SetEventHub(ev kaos.EventHub) Scanner
	EventHub() kaos.EventHub
	Scan(request toolkit.M) ([]toolkit.M, bool, error)
	CreateSession(config toolkit.M) *scannerSession
}

type BaseScanner struct {
	secret string
	opts   *ScannerOptions
	h      *datahub.Hub
	ev     kaos.EventHub
}

func (b *BaseScanner) CreateSession(config toolkit.M) *scannerSession {
	panic("not implemented")
}

func (b *BaseScanner) ID() string {
	panic("not implemented")
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

func (b *BaseScanner) Scan(request toolkit.M) ([]toolkit.M, bool, error) {
	panic("not implemented") // TODO: Implement
}
