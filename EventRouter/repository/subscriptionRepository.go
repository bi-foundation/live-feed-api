package repository

import (
	"live-api/EventRouter/api/models"
	"live-api/EventRouter/log"
	"strconv"
)

type Repository interface {
	StoreSubscription(subscription models.Subscription)
	ReadSubscription(id string) *models.Subscription
}

var id = 0
var tmpSubscriptionDB []models.Subscription

func StoreSubscription(subscription *models.Subscription) *models.Subscription {
	subscription.Id = strconv.Itoa(id)
	tmpSubscriptionDB = append(tmpSubscriptionDB, *subscription)
	id++
	log.Info("stored subscription: %v", subscription)
	return subscription
}

func ReadSubscription(id string) *models.Subscription {
	for _, subscription := range tmpSubscriptionDB {
		if subscription.Id == id {
			return &subscription
		}
	}
	return nil
}