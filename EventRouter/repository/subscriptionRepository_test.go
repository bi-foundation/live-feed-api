package repository_test

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/FactomProject/live-api/EventRouter/repository/inmemory"
	"github.com/FactomProject/live-api/EventRouter/repository/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCRUD(t *testing.T) {
	sqlRepository, _ := sql.New()

	repositories := map[string]repository.Repository{
		"inmemory": inmemory.New(),
		"sql":      sqlRepository,
	}

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testCRUD(t, repo)
		})
	}
}

func testCRUD(t *testing.T, repository repository.Repository) {
	subscription := &models.Subscription{
		Id:           "ID",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 1"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("filtering 2"))},
			models.COMMIT_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 3"))},
		},
	}

	createdSubscription, err := repository.CreateSubscription(subscription)
	assert.Nil(t, err)
	assertSubscription(t, subscription, createdSubscription)

	readSubscription, err := repository.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assertSubscription(t, subscription, readSubscription)

	substituteSubscription := &models.Subscription{
		Id:           "ID",
		Callback:     "updated-url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering update 1"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("filtering update 2"))},
		},
	}

	updatedSubscription, err := repository.UpdateSubscription(subscription.Id, substituteSubscription)
	assert.Nil(t, err)
	assertSubscription(t, substituteSubscription, updatedSubscription)

	deletedSubscription, err := repository.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
	assertSubscription(t, substituteSubscription, deletedSubscription)

	unknownSubscription, err := repository.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}

func assertSubscription(t *testing.T, expected *models.Subscription, actual *models.Subscription) {
	if actual == nil {
		assert.Fail(t, "subscription is nil")
		return
	}
	assert.NotNil(t, actual.Id)
	assert.Equal(t, expected.Callback, actual.Callback)
	assert.Equal(t, expected.CallbackType, actual.CallbackType)
	assert.Equal(t, len(expected.Filters), len(actual.Filters))

	for eventType, filter := range expected.Filters {
		assert.NotNil(t, actual.Filters[eventType])
		assert.Equal(t, filter.Filtering, actual.Filters[eventType].Filtering)
	}
}
