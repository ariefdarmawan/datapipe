package datapipe

import (
	"fmt"
	"path"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"github.com/logrusorgru/aurora"
)

const (
	ServiceTypeScan   string = "Scanner"
	ServiceTypeWorker        = "Worker"

	CmdSetupScanner string = "SetupScanner"
	CmdSetupWorker         = "SetupWorker"

	TopicNewService string = "NewService"
	TopicNewScanner        = "NewScanner"
	TopicNewWorker         = "NewWorker"
)

type PipeService struct {
	ID          string
	ServiceType string
	Providers   []string
}

type PipeEngine struct {
	services []*PipeService
	objects  []PipeObject
	cluster  string

	ks    *kaos.Service
	ev    kaos.EventHub
	dh    *datahub.Hub
	pipes map[string]*Pipe
}

func NewPipeEngine(name string, ks *kaos.Service) *PipeEngine {
	eng := new(PipeEngine)
	eng.ks = ks
	eng.ev, _ = ks.EventHub("default")
	eng.dh, _ = ks.GetDataHub("default")
	eng.cluster = name
	ks.RegisterModel(eng, name).SetDeploy(false).SetEvent(true)
	return eng
}

func (p *PipeEngine) NewService(ctx *kaos.Context, sv *PipeService) (string, error) {
	p.services = append(p.services, sv)
	ctx.Log().Infof("Registering service %s to Pipeline Engine", sv.ID)
	return "OK", nil
}

type PipeServiceProvider struct {
	ServiceID string
	Provider  string
	Add       bool
}

func (p *PipeEngine) NewServiceProvider(ctx *kaos.Context, req PipeServiceProvider) (string, error) {
	for _, s := range p.services {
		if s.ID == req.ServiceID {
			if req.Add {
				found := false
				for _, prov := range s.Providers {
					if prov == req.Provider {
						found = true
						break
					}
				}
				if !found {
					s.Providers = append(s.Providers, req.Provider)
					ctx.Log().Infof("Add provider %s to service %s", req.Provider, s.ID)
				}
			} else {
				found := false
				newps := []string{}
				for _, prov := range s.Providers {
					if prov != req.Provider {
						newps = append(newps, prov)
					} else {
						found = true
					}
				}
				if found {
					s.Providers = newps
				}
				ctx.Log().Infof("Remove provider %s from service %s", req.Provider, s.ID)
			}
			break
		}
	}
	return "OK", nil
}

type NewScannerRequest struct {
	ID       string
	Provider string
	Config   interface{}
}

type NewWorkerRequest struct {
	ID           string
	Provider     string
	TriggerTopic string
	Config       toolkit.M
}

func (p *PipeEngine) NewScanner(ctx *kaos.Context, parm *NewScannerRequest) (string, error) {
	ev, _ := ctx.DefaultEvent()
	services := []*PipeService{}

	if len(services) == 0 {
		return "", fmt.Errorf("No scan service with provider %s", parm.Provider)
	}
	rnd := toolkit.RandInt(len(services)) - 1
	if rnd < 0 {
		rnd = 0
	}
	sv := services[rnd]
	if err := ev.Publish(sv.ID+"|"+CmdSetupScanner, parm, nil); err != nil {
		return "NOK", err
	}
	return "OK", nil
}

/*
func (p *PipeEngine) NewWorker(ctx *kaos.Context, parm *NewWorkerRequest) (string, error) {
	ev, _ := ctx.DefaultEvent()
	services := []*PipeService{}
	if len(services) == 0 {
		return "", fmt.Errorf("No worker service with provider %s", parm.Provider)
	}
	rnd := toolkit.RandInt(len(services)) - 1
	if rnd < 0 {
		rnd = 0
	}
	sv := services[rnd]
	if err := ev.Publish(sv.ID+"|"+CmdSetupWorker, parm, nil); err != nil {
		return "NOK", err
	}
	return "OK", nil
}
*/

func (p *PipeEngine) Services() []*PipeService {
	return p.services
}

func (p *PipeEngine) Pipes() []Pipe {
	pipes := []Pipe{}
	for _, p := range p.pipes {
		pipes = append(pipes, *p)
	}
	return pipes
}

func (p *PipeEngine) PipeObjects() []PipeObject {
	return p.objects
}

type PipeDeployRequest struct {
	PreTopic string
}

func (eng *PipeEngine) Attach(p *Pipe) error {
	err := eng.provisionScanner(p)
	if err != nil {
		return err
	}

	if err = eng.provisionWorker(p); err != nil {
		return err
	}
	return nil
}

func (eng *PipeEngine) provisionScanner(p *Pipe) error {
	preTopic := path.Join(eng.ks.BasePoint(), eng.cluster)
	if eng.ev == nil {
		return fmt.Errorf("no valid eventhub")
	}

	mScanner := toolkit.M{}
	scanDeployTopic := strings.ToLower(path.Join(preTopic, "create", p.ScannerConfig.Provider))
	if err := eng.ev.Publish(scanDeployTopic, p.ScannerConfig.Opts, &mScanner); err != nil {
		return fmt.Errorf("fail to deploy scanner: %s | Data: %s",
			err.Error(),
			toolkit.JsonString(toolkit.M{}.Set("Topic", scanDeployTopic).Set("Opts", p.ScannerConfig.Opts)))
	}
	p.ScannerConfig.ScannerID = mScanner.GetString("_id")
	eng.objects = append(eng.objects,
		PipeObject{ID: p.ScannerConfig.ScannerID, Kind: "Scanner", PipeID: p.ID, ContainerID: mScanner.GetString("ContainerID"),
			Provider: p.ScannerConfig.Provider})

	if eng.pipes == nil {
		eng.pipes = make(map[string]*Pipe)
	}
	eng.pipes[p.ID] = p

	// subcribe to scanner result
	scanResultTopic := path.Join(eng.ks.BasePoint(), p.ScannerConfig.ScannerID, "result")
	eng.ev.SubscribeEx(scanResultTopic, eng.ks, nil, func(ctx *kaos.Context, req *WorkerMessage) (toolkit.M, error) {
		mres := toolkit.M{}
		sr := new(ScanResult)
		sr.ID = req.ID
		if err := eng.dh.Get(sr); err != nil {
			return mres, fmt.Errorf("fail get scan-result %s. %s", req.ID, err.Error())
		}
		eng.ks.Log().Infof("Initiating data flow %s, Process ID: %s", aurora.Yellow(p.ID), aurora.Yellow(req.ID))
		return mres, nil
	})
	return nil
}

func (eng *PipeEngine) provisionWorker(p *Pipe) error {
	if eng.ev == nil {
		return fmt.Errorf("no valid eventhub")
	}

	for _, wrkConfig := range p.WorkerConfigs {
		deployTopic := strings.ToLower(path.Join(eng.ks.BasePoint(), eng.cluster, "create", wrkConfig.Provider))
		err := eng.ev.Publish(deployTopic,
			toolkit.M{}.Set("Pipe", p.ID).Set("WorkerName", wrkConfig.Name),
			nil)
		if err != nil {
			return fmt.Errorf("fail to deploy worker %s with error: %s, parm: %s", wrkConfig.Name, err.Error(), toolkit.JsonString(wrkConfig))
		}

		// listen for /bp/cluster/pipeid/workerName/notify to notify PipeEngine for new worker
		notifyTopic := strings.ToLower(path.Join(eng.ks.BasePoint(), eng.cluster, p.ID, wrkConfig.Name, "notify"))
		eng.ev.SubscribeEx(notifyTopic, eng.ks, nil, func(ctx *kaos.Context, req toolkit.M) (toolkit.M, error) {
			mres := toolkit.M{}

			p.mtx.Lock()
			defer p.mtx.Unlock()

			workerID := req.GetString("_id")
			containerID := req.GetString("ContainerID")
			workerName := req.GetString("WorkerName")
			providerName := req.GetString("Provider")
			for workerIndex, cfg := range p.WorkerConfigs {
				if cfg.Name == workerName && cfg.Provider == providerName {
					wIDs := cfg.WorkerIDs
					wIDs = append(wIDs, workerID)
					cfg.WorkerIDs = wIDs
					p.WorkerConfigs[workerIndex] = cfg

					eng.objects = append(eng.objects, PipeObject{ID: workerID, Kind: "Worker", Provider: providerName, PipeID: p.ID, ContainerID: containerID})
					break
				}
			}

			return mres, nil
		})
	}

	return nil
}
