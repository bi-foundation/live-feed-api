package repository

import (
	"github.com/FactomProject/live-api/EventRouter/repository/inmemory"
)

var SubscriptionRepository Repository = inmemory.New()
