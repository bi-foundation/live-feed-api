package repository

import (
	"github.com/FactomProject/live-api/EventRouter/models"
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

	readSubscription, err := SubscriptionRepository.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, readSubscription.Id)
	assert.Equal(t, subscription.Callback, readSubscription.Callback)

	substituteSubscription := &models.Subscription{
		Id:       "ID",
		Callback: "updated-url",
	}
	updatedSubscription, err := SubscriptionRepository.UpdateSubscription(subscription.Id, substituteSubscription)
	assert.Nil(t, err)
	assert.Equal(t, substituteSubscription.Id, updatedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, updatedSubscription.Callback)

	deletedSubscription, err := SubscriptionRepository.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, deletedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, deletedSubscription.Callback)

	unknownSubscription, err := SubscriptionRepository.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}
