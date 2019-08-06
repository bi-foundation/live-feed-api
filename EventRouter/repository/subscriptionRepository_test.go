package repository

import (
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCRUD(t *testing.T) {
	subscription := &models.Subscription{
		Id:       "ID",
		Callback: "url",
	}

	createdSubscription, err := SubscriptionRepository.CreateSubscription(subscription)

	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, createdSubscription.Id)
	assert.Equal(t, subscription.Callback, createdSubscription.Callback)

	readedSubscription, err := SubscriptionRepository.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, readedSubscription.Id)
	assert.Equal(t, subscription.Callback, readedSubscription.Callback)

	deletedSubscription, err := SubscriptionRepository.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, deletedSubscription.Id)
	assert.Equal(t, subscription.Callback, deletedSubscription.Callback)

	unknownSubscription, err := SubscriptionRepository.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}
