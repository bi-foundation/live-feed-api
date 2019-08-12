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

func TestConcurrency(t *testing.T) {
	n := 100
	wait := sync.WaitGroup{}
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(x int) {
			defer wait.Done()
			subscription := &models.Subscription{
				Callback: fmt.Sprintf("url: %d", x),
			}

			subscription, err := repo.CreateSubscription(subscription)
			assert.Nil(t, err)
			t.Logf("%d: created %s", x, subscription.Id)
		}(i)
	}
	wait.Wait()

	assert.Equal(t, n, len(repo.db))
}
