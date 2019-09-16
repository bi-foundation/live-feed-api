//go:generate make -C EventRouter clean
//go:generate make -C EventRouter
//go:generate $GOPATH/bin/swag init -g ./EventRouter/api/subscriptionApi.go -o ./EventRouter/swagger
package main

import (
	"github.com/FactomProject/live-feed-api/EventRouter/api"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/events"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	docs "github.com/FactomProject/live-feed-api/EventRouter/swagger"
	"time"
)

func main() {
	// use info from swagger to init will be called to register the swagger which is provided through an endpoint
	info := docs.SwaggerInfo
	log.Info("start %s %s", info.Title, info.Version)

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
