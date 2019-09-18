//go:generate make -C EventRouter clean
//go:generate make -C EventRouter
//go:generate $GOPATH/bin/swag init -g ./EventRouter/api/subscriptionApi.go -o ./EventRouter/swagger
package main

import (
	"flag"
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

	configuration := loadConfiguration()
	log.SetLevel(log.Parse(configuration.Log.LogLevel))

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

func loadConfiguration() *config.Config {
	explicitConfigFile := flag.String("config-file", "", "Override the configuration file")
	flag.Parse()
	var configuration *config.Config
	var err error
	if explicitConfigFile != nil && len(*explicitConfigFile) > 0 {
		configuration, err = config.LoadConfigurationFrom(*explicitConfigFile)
	} else {
		configuration, err = config.LoadConfiguration()
	}
	if err != nil {
		log.Fatal("failed to load configuration: %v", err)
	}
	return configuration
}
