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

	createdSubscription := CreateSubscription(subscription)

	assert.Equal(t, subscription.Id, createdSubscription.Id)
	assert.Equal(t, subscription.Callback, createdSubscription.Callback)

	readedSubscription := ReadSubscription(subscription.Id)
	assert.Equal(t, subscription.Id, readedSubscription.Id)
	assert.Equal(t, subscription.Callback, readedSubscription.Callback)

	deletedSubscription := DeleteSubscription(subscription.Id)
	assert.Equal(t, subscription.Id, deletedSubscription.Id)
	assert.Equal(t, subscription.Callback, deletedSubscription.Callback)

	unknownSubscription := ReadSubscription(subscription.Id)
	assert.Nil(t, unknownSubscription)
}
