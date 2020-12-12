package datapipe

import "sync"

const (
	CloseNothing string = ""
	CloseOnError        = "CloseOnError"
	CloseOnOK           = "CloseOnOK"
	CloseOnAny          = "CloseOnAny"
)

type PipeScannerConfig struct {
	Provider  string
	Opts      interface{}
	ScannerID string
}

type PipeWorkerConfig struct {
	Name      string
	Provider  string
	Input     string
	ClosePipe string
	RunMode   string
	Options   interface{}
	WorkerIDs []string
}

type PipeObject struct {
	ID          string
	Kind        string
	PipeID      string
	ContainerID string
	Provider    string
}

type Pipe struct {
	ID            string
	ScannerConfig PipeScannerConfig
	WorkerConfigs []PipeWorkerConfig
	mtx           *sync.RWMutex
}

func NewPipe(name string) *Pipe {
	p := new(Pipe)
	p.ID = name
	p.mtx = new(sync.RWMutex)
	return p
}

func (p *Pipe) UseScanner(provider string, opts interface{}) {
	p.ScannerConfig = PipeScannerConfig{Provider: provider, Opts: opts}
}

func (p *Pipe) AddWorker(name string, workerConfig PipeWorkerConfig) {
	workerConfig.Name = name
	workerConfig.WorkerIDs = []string{}
	for idx, cfg := range p.WorkerConfigs {
		if cfg.Name == name {
			p.WorkerConfigs[idx] = workerConfig
			return
		}
	}
	p.WorkerConfigs = append(p.WorkerConfigs, workerConfig)
}

func (p *Pipe) RemoveWorker(name string) {
	nws := []PipeWorkerConfig{}
	for _, cfg := range p.WorkerConfigs {
		if cfg.Name != name {
			nws = append(nws, cfg)
		}
	}
	p.WorkerConfigs = nws
}
