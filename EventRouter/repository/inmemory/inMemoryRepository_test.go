package inmemory

import (
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const initId = 0

var repo *inMemoryRepository

func init() {
	repo = New()
}

func TestCRUD(t *testing.T) {
	subscription := &models.Subscription{
		Id:       "ID",
		Callback: "url",
	}
	createdSubscription, err := repo.CreateSubscription(subscription)

	id := strconv.Itoa(initId)

	assert.Nil(t, err)
	assert.Equal(t, id, createdSubscription.Id)
	assert.Equal(t, subscription.Callback, createdSubscription.Callback)

	readSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, id, readSubscription.Id)
	assert.Equal(t, subscription.Callback, readSubscription.Callback)

	substituteSubscription := &models.Subscription{
		Id:       createdSubscription.Id,
		Callback: "updated-url",
	}
	updatedSubscription, err := repo.UpdateSubscription(substituteSubscription)
	assert.Nil(t, err)
	assert.Equal(t, id, updatedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, updatedSubscription.Callback)

	err = repo.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)

	unknownSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}
