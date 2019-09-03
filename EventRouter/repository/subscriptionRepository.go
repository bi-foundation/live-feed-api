package repository

import (
	"github.com/FactomProject/live-feed-api/EventRouter/repository/inmemory"
)

var SubscriptionRepository Repository = inmemory.New()
