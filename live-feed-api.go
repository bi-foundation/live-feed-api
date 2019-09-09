//go:generate make -C EventRouter clean
//go:generate make -C EventRouter
//go:generate $GOPATH/bin/swagger generate spec --scan-models -w ./EventRouter -o ./swagger.json
//go:generate $GOPATH/bin/swagger validate swagger.json
package main

import (
	"github.com/FactomProject/live-feed-api/EventRouter/api"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/events"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"time"
)

func main() {
	configuration, err := config.LoadConfiguration()
	if err != nil {
		log.Fatal("failed to load configuration: %v", err)
	}
	log.SetLevel(log.Parse(configuration.Log.LogLevel))

	eventServer := events.NewReceiver(configuration.Receiver)
	eventRouter := events.NewEventRouter(eventServer.GetEventQueue())

	eventServer.Start()
	eventRouter.Start()

	api.NewSubscriptionApi(configuration.Subscription).Start()

	for eventServer.GetState() < models.Stopping {
		time.Sleep(time.Second)
	}

	eventServer.Stop()
}
