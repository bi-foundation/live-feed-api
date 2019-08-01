package repository

import (
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/FactomProject/live-api/EventRouter/log"
	"strconv"
)

type Repository interface {
	StoreSubscription(subscription models.Subscription)
	ReadSubscription(id string) *models.Subscription
}

var id = 0
var tmpSubscriptionDB []models.Subscription

func CreateSubscription(subscription *models.Subscription) *models.Subscription {
	subscription.Id = strconv.Itoa(id)
	tmpSubscriptionDB = append(tmpSubscriptionDB, *subscription)
	id++
	log.Debug("stored subscription: %v", subscription)
	return subscription
}

func ReadSubscription(id string) *models.Subscription {
	for _, subscription := range tmpSubscriptionDB {
		if subscription.Id == id {
			log.Info("read subscription: %v", subscription)
			return &subscription
		}
	}
	log.Debug("subscription not found: %s", id)
	return nil
}

func DeleteSubscription(id string) *models.Subscription {
	var index = -1
	for i, subscription := range tmpSubscriptionDB {
		if subscription.Id == id {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}

	subscription := tmpSubscriptionDB[index]
	tmpSubscriptionDB = append(tmpSubscriptionDB[:index], tmpSubscriptionDB[index+1:]...)
	log.Debug("deleted subscription: %v", subscription)
	return &subscription
}
