package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datapipe/library/kdp"
	"github.com/ariefdarmawan/datapipe/library/mgowr"
	"github.com/ariefdarmawan/kconfigurator"
	"github.com/eaciit/toolkit"
	"github.com/kanoteknologi/knats"
	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "github.com/ariefdarmawan/flexmgo"
)

var (
	e error
	s *kaos.Service

	appConfig      = new(kconfigurator.AppConfig)
	nats           = flag.String("n", "nats://localhost:4222", "address of NATS server")
	secret         = flag.String("key", "", "key-secret for msvc")
	getConfigTopic = flag.String("topic", "/v1/config/get", "name of topic to get configuration")
	serviceName    = mgowr.WorkerName
	version        = "v1"
	nodeID         string

	log *toolkit.LogEngine
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeID = primitive.NewObjectID().Hex()
	log = appkit.LogWithPrefix(fmt.Sprintf("%s-%s", serviceName, nodeID[len(nodeID)-4:]))

	toolkit.DefaultCase = toolkit.CaseAsIs
	toolkit.SetTagName(toolkit.CaseAsIs)
	flag.Parse()

	// services
	ev := knats.NewEventHub(*nats, byter.NewByter("")).SetSignature(*secret).SetSecret(*secret)
	defer ev.Close()

	appConfig, err := kconfigurator.GetConfigFromEventHub(ev, *getConfigTopic)
	if err != nil {
		log.Errorf("config error. %s", err.Error())
		os.Exit(1)
	}
	log.Infof("successfully reading config from NATS")

	h, err := kconfigurator.MakeHub(appConfig, "default")
	if err != nil {
		log.Errorf("fail to connect to db. %s", err.Error())
		os.Exit(1)
	}
	defer h.Close()

	s := kaos.NewService().SetLogger(log).SetContext(ctx).SetBasePoint(version).
		RegisterDataHub(h, "default").
		RegisterEventHub(ev, "default", appConfig.EventServer.Group)

	ts := kdp.NewKxWorker(s, mgowr.NewEngine(nodeID))
	defer ts.StopEngine()

	// deployy
	if err = s.ActivateEvent(); err != nil {
		log.Errorf("fail to activate event: %s", err.Error())
		os.Exit(1)
	}

	if err = ts.PingCoordinator(); err != nil {
		log.Errorf("fail to connect to coordinator: %s", err.Error())
		os.Exit(1)
	}
	defer ts.UnpingCoordinator()

	log.Infof("Starting service %s", serviceName)
	csign := make(chan os.Signal)
	signal.Notify(csign, os.Interrupt, os.Kill)
	<-csign
	log.Infof("stopping %v service", serviceName)
}
