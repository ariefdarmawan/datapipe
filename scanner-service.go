package datapipe

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/eaciit/toolkit"
)

type NewScannerFn func(service, topic string, opt toolkit.M) (Scanner, error)

type ScannerService interface {
	Start() error
	RegisterScanProvider(name string, fn NewScannerFn) ScannerService
	RegisterScanner(provider, name, topic string, opts interface{}) (Scanner, error)
	UpdateResultStatus(ctx *kaos.Context, req ScanResultUpdateStatusRequest) (string, error)
	Scanners() []Scanner
	CloseScanner(clearData bool, names ...string)
	GetScannerStat(name string) *ScannerStat
	Stop()
}

type ScanResultUpdateStatusRequest struct {
	ID, Status, WorkerID, WorkerProvider string
}
