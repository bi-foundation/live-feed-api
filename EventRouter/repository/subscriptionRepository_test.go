package repository_test

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/FactomProject/live-api/EventRouter/repository/inmemory"
	"github.com/FactomProject/live-api/EventRouter/repository/sql"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var repositories map[string]repository.Repository

func init() {
	log.SetLevel(log.D)

	sqlRepository, err := sql.New()
	if err != nil {
		log.Error("setup test: %v", err)
	}

	repositories = map[string]repository.Repository{
		"inmemory": inmemory.New(),
		"sql":      sqlRepository,
	}
}

func TestCRUD(t *testing.T) {
	subscription := &models.Subscription{
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 1"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("filtering 2"))},
			models.COMMIT_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 3"))},
		},
	}

	substituteSubscription := &models.Subscription{
		Callback:     "updated-url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering update 1"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("filtering update 2"))},
		},
	}

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testCreate(t, repo, subscription)
			testRead(t, repo, subscription)

			substituteSubscription.Id = subscription.Id
			testUpdate(t, repo, substituteSubscription)
			testRead(t, repo, substituteSubscription)

			testDelete(t, repo, substituteSubscription)
			testNoExits(t, repo, subscription)
		})
	}
}

func testCreate(t *testing.T, repository repository.Repository, subscription *models.Subscription) {
	createdSubscription, err := repository.CreateSubscription(subscription)
	assert.Nil(t, err)
	assertSubscription(t, subscription, createdSubscription)
	subscription.Id = createdSubscription.Id
}

func testRead(t *testing.T, repository repository.Repository, subscription *models.Subscription) {
	readSubscription, err := repository.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assertSubscription(t, subscription, readSubscription)
}

func testUpdate(t *testing.T, repository repository.Repository, subscription *models.Subscription) {
	updatedSubscription, err := repository.UpdateSubscription(subscription)
	assert.Nil(t, err)
	assertSubscription(t, subscription, updatedSubscription)
}

func testDelete(t *testing.T, repository repository.Repository, subscription *models.Subscription) {
	err := repository.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)
}

func testNoExits(t *testing.T, repository repository.Repository, subscription *models.Subscription) {
	unknownSubscription, err := repository.ReadSubscription(subscription.Id)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscription)
}

func assertSubscription(t *testing.T, expected *models.Subscription, actual *models.Subscription) {
	if actual == nil {
		assert.FailNow(t, "subscription is nil")
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

func TestConcurrency(t *testing.T) {
	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testConcurrency(t, repo)
		})
	}
}

func testConcurrency(t *testing.T, repository repository.Repository) {
	eventType := models.COMMIT_ENTRY
	subscription := &models.Subscription{

		Callback: "url",
		Filters: map[models.EventType]models.Filter{
			eventType: {Filtering: ""},
		},
	}

	// calculate the offset if the database already has entries
	// Although the database should be clean and clean-up afterwards,
	previousSubscriptions, err := repository.GetSubscriptions(eventType)
	offset := len(previousSubscriptions)

	n := 100
	wait := sync.WaitGroup{}
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(x int) {
			defer wait.Done()

			subscription, err := repository.CreateSubscription(subscription)
			assert.Nil(t, err)
			t.Logf("%d: created %s", x, subscription.Id)
		}(i)
	}
	wait.Wait()

	subscriptions, err := repository.GetSubscriptions(eventType)
	assert.Nil(t, err)
	assert.Equal(t, n, len(subscriptions)-offset)
}
