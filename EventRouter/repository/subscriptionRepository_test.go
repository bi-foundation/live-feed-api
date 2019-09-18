package repository

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var repositories map[string]Repository

func init() {
	log.SetLevel(log.D)

	repositories = map[string]Repository{
		"inmemory": NewInMemoryRepository(),
	}
}

func testCRUD(t *testing.T) {

	subscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			CallbackURL:        "url",
			CallbackType:       models.BearerToken,
			SubscriptionStatus: models.Active,
			Filters: map[models.EventType]models.Filter{
				models.BlockCommit:       {Filtering: fmt.Sprintf("filtering 1")},
				models.EntryRegistration: {Filtering: fmt.Sprintf("filtering 2")},
				models.ChainRegistration: {Filtering: fmt.Sprintf("filtering 3")},
			},
			Credentials: models.Credentials{
				AccessToken: "token",
			},
		},
		Failures: 0,
	}

	substituteSubscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			CallbackURL:        "updated-url",
			CallbackType:       models.BasicAuth,
			SubscriptionStatus: models.Suspended,
			SubscriptionInfo:   "reason",
			Filters: map[models.EventType]models.Filter{
				models.BlockCommit:       {Filtering: fmt.Sprintf("filtering update 1")},
				models.EntryRegistration: {Filtering: fmt.Sprintf("filtering update 2")},
			},
			Credentials: models.Credentials{
				BasicAuthUsername: "username",
				BasicAuthPassword: "password",
			},
		},
		Failures: 0,
	}

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testCreate(t, repo, subscriptionContext)
			testRead(t, repo, subscriptionContext)

			substituteSubscriptionContext.Subscription.ID = subscriptionContext.Subscription.ID
			testUpdate(t, repo, substituteSubscriptionContext)
			testRead(t, repo, substituteSubscriptionContext)

			testDelete(t, repo, substituteSubscriptionContext)
			testNoExits(t, repo, subscriptionContext)
		})
	}
}

func testCreate(t *testing.T, repository Repository, subscriptionContext *models.SubscriptionContext) {
	createdSubscriptionContext, err := repository.CreateSubscription(subscriptionContext)
	assertNilError(t, err)
	assertSubscription(t, subscriptionContext, createdSubscriptionContext)
	subscriptionContext.Subscription.ID = createdSubscriptionContext.Subscription.ID
}

func testRead(t *testing.T, repository Repository, subscriptionContext *models.SubscriptionContext) {
	readSubscriptionContext, err := repository.ReadSubscription(subscriptionContext.Subscription.ID)
	assertNilError(t, err)
	assertSubscription(t, subscriptionContext, readSubscriptionContext)
}

func testUpdate(t *testing.T, repository Repository, subscriptionContext *models.SubscriptionContext) {
	updatedSubscriptionContext, err := repository.UpdateSubscription(subscriptionContext)
	assertNilError(t, err)
	assertSubscription(t, subscriptionContext, updatedSubscriptionContext)
}

func testDelete(t *testing.T, repository Repository, subscriptionContext *models.SubscriptionContext) {
	err := repository.DeleteSubscription(subscriptionContext.Subscription.ID)
	assertNilError(t, err)
}

func testNoExits(t *testing.T, repository Repository, subscriptionContext *models.SubscriptionContext) {
	unknownSubscriptionContext, err := repository.ReadSubscription(subscriptionContext.Subscription.ID)
	assert.NotNil(t, err)
	assert.Nil(t, unknownSubscriptionContext)
}

func assertSubscription(t *testing.T, expected *models.SubscriptionContext, actual *models.SubscriptionContext) {
	if actual == nil {
		assert.Fail(t, "subscription is nil")
		return
	}
	assert.NotNil(t, actual.Subscription.ID)
	assert.Equal(t, expected.Failures, actual.Failures)
	assert.Equal(t, expected.Subscription.CallbackURL, actual.Subscription.CallbackURL)
	assert.Equal(t, expected.Subscription.CallbackType, actual.Subscription.CallbackType)
	assert.Equal(t, expected.Subscription.SubscriptionStatus, actual.Subscription.SubscriptionStatus)
	assert.Equal(t, expected.Subscription.SubscriptionInfo, actual.Subscription.SubscriptionInfo)
	assert.Equal(t, expected.Subscription.Credentials.AccessToken, actual.Subscription.Credentials.AccessToken)
	assert.Equal(t, expected.Subscription.Credentials.BasicAuthUsername, actual.Subscription.Credentials.BasicAuthUsername)
	assert.Equal(t, expected.Subscription.Credentials.BasicAuthPassword, actual.Subscription.Credentials.BasicAuthPassword)
	assert.Equal(t, len(expected.Subscription.Filters), len(actual.Subscription.Filters))

	for eventType, filter := range expected.Subscription.Filters {
		assert.NotNil(t, actual.Subscription.Filters[eventType])
		assert.Equal(t, filter.Filtering, actual.Subscription.Filters[eventType].Filtering)
	}
}

func testConcurrencyAllRepos(t *testing.T) {
	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testConcurrency(t, repo)
		})
	}
}

func testConcurrency(t *testing.T, repository Repository) {
	eventType := models.EntryRegistration
	subscription := models.Subscription{
		CallbackURL:        "url",
		SubscriptionStatus: models.Active,
		Filters: map[models.EventType]models.Filter{
			eventType: {Filtering: ""},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	// calculate the offset if the database already has entries
	// Although the database should be clean and clean-up afterwards,
	previousSubscriptions, err := repository.GetActiveSubscriptions(eventType)
	offset := len(previousSubscriptions)

	n := 100
	wait := sync.WaitGroup{}
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(x int) {
			defer wait.Done()

			subscriptionContext, err := repository.CreateSubscription(subscriptionContext)
			assert.Nil(t, err)
			t.Logf("%d: created %s", x, subscriptionContext.Subscription.ID)
		}(i)
	}
	wait.Wait()

	subscriptions, err := repository.GetActiveSubscriptions(eventType)
	assert.Nil(t, err)
	assert.Equal(t, n, len(subscriptions)-offset)
}

func assertNilError(t *testing.T, err error) {
	if err != nil {
		assert.Nil(t, err)
		t.FailNow()
	}
}
