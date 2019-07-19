package main

import (
	"github.com/FactomProject/live-api/EventRouter/events"
	"github.com/FactomProject/live-api/EventRouter/network"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"time"
)

var (
	eventServer = network.NewEventServer()
	eventRouter = events.NewEventRouter(eventServer.GetEventsInQueue())
)

func main() {
	eventServer.Start()
	eventRouter.Start()

	for eventServer.GetState() < runstate.Stopping {
		time.Sleep(time.Second)
	}

	eventServer.Stop()
}
