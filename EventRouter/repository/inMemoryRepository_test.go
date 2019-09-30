package repository

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

const initID = 0

var repo *inMemoryRepository

func init() {
	repo, _ = NewInMemoryRepository().(*inMemoryRepository)
}

func TestCRUD(t *testing.T) {
	subscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			ID:                 "ID",
			CallbackURL:        "url",
			SubscriptionStatus: models.Active,
			Filters: map[models.EventType]models.Filter{
				models.NodeMessage: {Filtering: fmt.Sprintf("filtering 1")},
			},
		},
	}
	createdSubscription, err := repo.CreateSubscription(subscriptionContext)

	id := strconv.Itoa(initID)

	assert.Nil(t, err)
	assert.Equal(t, id, createdSubscription.Subscription.ID)
	assert.Equal(t, subscriptionContext.Subscription.CallbackURL, createdSubscription.Subscription.CallbackURL)

	readSubscription, err := repo.ReadSubscription(subscriptionContext.Subscription.ID)
	assert.Nil(t, err)
	assert.Equal(t, id, readSubscription.Subscription.ID)
	assert.Equal(t, subscriptionContext.Subscription.CallbackURL, readSubscription.Subscription.CallbackURL)

	allSubscriptions, err := repo.GetActiveSubscriptions(models.NodeMessage)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allSubscriptions))
	assert.Equal(t, id, allSubscriptions[0].Subscription.ID)
	assert.Equal(t, subscriptionContext.Subscription.CallbackURL, allSubscriptions[0].Subscription.CallbackURL)
	assert.Equal(t, subscriptionContext.Subscription.Filters[models.NodeMessage].Filtering, allSubscriptions[0].Subscription.Filters[models.NodeMessage].Filtering)

	substituteSubscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			ID:          createdSubscription.Subscription.ID,
			CallbackURL: "updated-url",
		},
	}

	updatedSubscription, err := repo.UpdateSubscription(substituteSubscriptionContext)
	assert.Nil(t, err)
	assert.Equal(t, id, updatedSubscription.Subscription.ID)
	assert.Equal(t, substituteSubscriptionContext.Subscription.CallbackURL, updatedSubscription.Subscription.CallbackURL)

	err = repo.DeleteSubscription(subscriptionContext.Subscription.ID)
	assert.Nil(t, err)

	unknownSubscription, err := repo.ReadSubscription(subscriptionContext.Subscription.ID)
	assert.IsType(t, errors.SubscriptionNotFound{}, err)
	assert.Nil(t, unknownSubscription)
}

func TestConcurrency(t *testing.T) {
	n := 100
	wait := sync.WaitGroup{}
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(x int) {
			defer wait.Done()
			subscriptionContext := &models.SubscriptionContext{
				Subscription: models.Subscription{
					CallbackURL: fmt.Sprintf("url: %d", x),
				},
			}

			subscriptionContext, err := repo.CreateSubscription(subscriptionContext)
			assert.Nil(t, err)
			// t.Logf("%d: created %s", x, subscriptionContext.Subscription.ID)
		}(i)
	}
	wait.Wait()

	assert.Equal(t, n, len(repo.db))
}
