package inmemory

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

const initId = 0

var repo *inMemoryRepository

func init() {
	repo = New()
}

func TestCRUD(t *testing.T) {
	subscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			Id:       "ID",
			Callback: "url",
		},
	}
	createdSubscription, err := repo.CreateSubscription(subscriptionContext)

	id := strconv.Itoa(initId)

	assert.Nil(t, err)
	assert.Equal(t, id, createdSubscription.Subscription.Id)
	assert.Equal(t, subscriptionContext.Subscription.Callback, createdSubscription.Subscription.Callback)

	readSubscription, err := repo.ReadSubscription(subscriptionContext.Subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, id, readSubscription.Subscription.Id)
	assert.Equal(t, subscriptionContext.Subscription.Callback, readSubscription.Subscription.Callback)

	substituteSubscriptionContext := &models.SubscriptionContext{
		Subscription: models.Subscription{
			Id:       createdSubscription.Subscription.Id,
			Callback: "updated-url",
		},
	}
	updatedSubscription, err := repo.UpdateSubscription(substituteSubscriptionContext)
	assert.Nil(t, err)
	assert.Equal(t, id, updatedSubscription.Subscription.Id)
	assert.Equal(t, substituteSubscriptionContext.Subscription.Callback, updatedSubscription.Subscription.Callback)

	err = repo.DeleteSubscription(subscriptionContext.Subscription.Id)
	assert.Nil(t, err)

	unknownSubscription, err := repo.ReadSubscription(subscriptionContext.Subscription.Id)
	assert.NotNil(t, err)
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
					Callback: fmt.Sprintf("url: %d", x),
				},
			}

			subscriptionContext, err := repo.CreateSubscription(subscriptionContext)
			assert.Nil(t, err)
			t.Logf("%d: created %s", x, subscriptionContext.Subscription.Id)
		}(i)
	}
	wait.Wait()

	assert.Equal(t, n, len(repo.db))
}
