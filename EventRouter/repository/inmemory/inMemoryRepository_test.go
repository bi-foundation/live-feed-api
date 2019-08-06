package inmemory

import (
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

var repo repository.Repository

func init() {
	repo = &InMemoryRepository{}
}

func TestCRUD(t *testing.T) {
	subscription := &models.Subscription{
		Id:       "ID",
		Callback: "url",
	}
	createdSubscription, err := repo.CreateSubscription(subscription)

	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, createdSubscription.Id)
	assert.Equal(t, subscription.Callback, createdSubscription.Callback)

	readSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, readSubscription.Id)
	assert.Equal(t, subscription.Callback, readSubscription.Callback)

	substituteSubscription := &models.Subscription{
		Id:       "ID",
		Callback: "updated-url",
	}
	updatedSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, substituteSubscription.Id, updatedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, updatedSubscription.Callback)

	deletedSubscription, err := repo.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, substituteSubscription.Id, deletedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, deletedSubscription.Callback)

	unknownSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}
