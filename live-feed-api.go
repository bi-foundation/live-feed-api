//go:generate make -C EventRouter clean
//go:generate make -C EventRouter
//go:generate $GOPATH/bin/swagger generate spec --scan-models -w ./EventRouter -o ./swagger.json
//go:generate $GOPATH/bin/swagger validate swagger.json
package main

import (
	"github.com/FactomProject/live-feed-api/EventRouter/api"
	"github.com/FactomProject/live-feed-api/EventRouter/events"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"time"
)

var (
	appConfig   = config.LoadEventRouterConfig()
	eventServer = events.NewDefaultReceiver(appConfig.EventListenerConfig)
	eventRouter = events.NewEventRouter(eventServer.GetEventQueue())
)

func main() {
	log.SetLevel(log.D)

	eventServer.Start()
	eventRouter.Start()

	api.NewSubscriptionApi(appConfig.SubscriptionApiConfig).Start()

	for eventServer.GetState() < models.Stopping {
		time.Sleep(time.Second)
	}

	eventServer.Stop()
}