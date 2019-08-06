//go:generate $GOPATH/bin/swagger generate spec --scan-models -w ./EventRouter -o ./swagger.json
//go:generate $GOPATH/bin/swagger validate swagger.json
package main

import (
	"github.com/FactomProject/live-api/EventRouter/api"
	"github.com/FactomProject/live-api/EventRouter/events"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/network"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"time"
)

var (
	eventServer = network.NewDefaultServer()
	eventRouter = events.NewEventRouter(eventServer.GetEventQueue())
)

func main() {
	log.SetLevel(log.D)

	go eventServer.Start()
	eventRouter.Start()

	api.NewSubscriptionApi(":8700").Start()

	for eventServer.GetState() < runstate.Stopping {
		time.Sleep(time.Second)
	}

	eventServer.Stop()
}
