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
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
	docs "github.com/FactomProject/live-feed-api/EventRouter/swagger"
)

func main() {
	// use info from swagger to init will be called to register the swagger which is provided through an endpoint
	info := docs.SwaggerInfo
	log.Info("start %s %s", info.Title, info.Version)

	configuration := loadConfiguration()
	log.SetLevel(log.Parse(configuration.Log.LogLevel))
	setupDatabase(configuration.Database)

	eventServer := events.NewReceiver(configuration.Receiver)
	eventRouter := events.NewEventRouter(eventServer.GetEventQueue())

	eventServer.Start()
	eventRouter.Start()

	api.NewSubscriptionAPI(configuration.Subscription).Start()

	select {}
}

func loadConfiguration() (configuration *config.Config) {
	explicitConfigFile := flag.String("config-file", "", "Override the configuration file")
	flag.Parse()

	var err error
	if explicitConfigFile != nil && len(*explicitConfigFile) > 0 {
		log.Info("load configuration file: %s", *explicitConfigFile)
		configuration, err = config.LoadConfigurationFrom(*explicitConfigFile)
	} else {
		configuration, err = config.LoadConfiguration()
	}
	if err != nil {
		log.Fatal("%v", err)
	}

	log.Info("loaded configuration: { \n\treceiver: %v, \n\tsubscription: %v, \n\tdatabase: %v, \n\tlog: %v\n }", configuration.Receiver, configuration.Subscription, configuration.Database, configuration.Log)
	return configuration
}

func setupDatabase(configuration *config.DatabaseConfig) {
	switch configuration.Database {
	case "inmemory":
		repository.SubscriptionRepository = repository.NewInMemoryRepository()
	case "mysql":
		repo, err := repository.NewSQLRepository(configuration)
		if err != nil {
			log.Fatal("failed to configure database: %v", err)
		}
		repository.SubscriptionRepository = repo
	default:
		log.Fatal("failed to configure database: %v", configuration.Database)
	}
}
