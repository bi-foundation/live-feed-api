package sql

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

var repo repository.Repository

func init() {
	log.SetLevel(log.D)
	repository, err := NewSQLRepository()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	repo = repository
}

func TestCRUD(t *testing.T) {
	// TODO clean up, also if test fails
	subscription := &models.Subscription{
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
	updatedSubscription, err := repo.UpdateSubscription(subscription.Id, substituteSubscription)
	assert.Nil(t, err)
	assert.Equal(t, substituteSubscription.Id, updatedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, updatedSubscription.Callback)

	deletedSubscription, err := repo.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, deletedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, deletedSubscription.Callback)

	unknownSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}

func TestReadUnknownId(t *testing.T) {
	updatedSubscription, err := repo.ReadSubscription("unknownId")

	t.Logf("test read subscription error: %v", err)
	assert.NotNil(t, err)
	assert.Nil(t, updatedSubscription)
}

func TestUpdateUnknownId(t *testing.T) {
	subscription := &models.Subscription{
		Callback: "url",
	}
	updatedSubscription, err := repo.UpdateSubscription(subscription.Id, subscription)

	t.Logf("test update subscription error: %v", err)
	assert.NotNil(t, err)
	assert.Nil(t, updatedSubscription)
}
