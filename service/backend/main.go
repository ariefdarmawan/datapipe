package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kext/mod/kmdb"
	"git.kanosolution.net/kano/kext/mod/kmui"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe/engine"
	"github.com/ariefdarmawan/datapipe/model"
	"github.com/ariefdarmawan/kconfigurator"
	"github.com/google/uuid"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"

	_ "github.com/ariefdarmawan/datapipe/library/kfslocal"
	_ "github.com/ariefdarmawan/datapipe/library/kfsmn"
	_ "github.com/ariefdarmawan/flexmgo"
)

var (
	config      = flag.String("config", "config/app.yml", "path to config file")
	appConfig   = new(kconfigurator.AppConfig)
	serviceName = "backend"
	version     = "v1"
	log         = appkit.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()

	// log and config
	e := appkit.ReadConfig(*config, appConfig)
	if e != nil {
		log.Error(e.Error())
		os.Exit(1)
	}

	evServer := appConfig.EventServer
	if evServer.Group == "" {
		evServer.Group = uuid.New().String()
	}
	appConfig.EventServer = evServer

	// service
	ev := knats.NewEventHub(evServer.Server, byter.NewByter("")).SetSignature(appConfig.EventServer.Group)
	defer ev.Close()

	s := kaos.NewService().SetLogger(log).
		SetBasePoint(version).
		RegisterEventHub(ev, "default", evServer.Group)

	// datahub
	h, err := kconfigurator.MakeHub(appConfig, "default")
	if err != nil {
		log.Errorf("InvalidHub: %s", e.Error())
		os.Exit(1)
	}
	s.RegisterDataHub(h, "default")

	// model registration
	mdb := kmdb.New()
	mly := kmui.New()

	s.RegisterModel(new(model.Storage), "storage").SetMod(mdb, mly)
	s.RegisterModel(new(model.Connection), "connection").SetMod(mdb, mly)
	s.RegisterModel(new(model.Variable), "variable").SetMod(mdb, mly)

	s.RegisterModel(new(engine.StorageEngine), "storage")

	// mux
	mux := http.NewServeMux()

	if e = hd.NewHttpDeployer().Deploy(s, mux); e != nil {
		log.Errorf("unable to deploy. %s", e.Error())
		os.Exit(1)
	}

	// run service
	csign := make(chan os.Signal)
	hostName := appConfig.Hosts[serviceName]
	if hostName == "" {
		log.Errorf("unable to start service %s, hostname is not defined", serviceName)
		os.Exit(1)
	}
	go func() {
		s.Log().Infof("starting %v service on %s", serviceName, hostName)
		err := http.ListenAndServe(hostName, mux)
		if err != nil {
			csign <- syscall.SIGINT
		}
	}()

	// grace shutdown
	signal.Notify(csign, os.Interrupt, os.Kill)
	<-csign
	log.Infof("stopping %v service", serviceName)
}
