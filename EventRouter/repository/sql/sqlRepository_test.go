package sql

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

var repo *sqlRepository

func init() {
	log.SetLevel(log.D)
	repository, err := New()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	repo = repository
}

func TestCRUD(t *testing.T) {
	// TODO clean up, also if test fails
	subscription := &models.Subscription{
		Callback:     "url",
		CallbackType: models.HTTP,
	}
	createdSubscription, err := repo.CreateSubscription(subscription)

	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, createdSubscription.Id)
	assert.Equal(t, subscription.Callback, createdSubscription.Callback)
	assert.Equal(t, subscription.CallbackType, createdSubscription.CallbackType)

	readSubscription, err := repo.ReadSubscription(subscription.Id)
	assert.Nil(t, err)
	assert.Equal(t, subscription.Id, readSubscription.Id)
	assert.Equal(t, subscription.Callback, readSubscription.Callback)
	assert.Equal(t, subscription.CallbackType, readSubscription.CallbackType)

	substituteSubscription := &models.Subscription{
		Id:       readSubscription.Id,
		Callback: "updated-url",
	}
	updatedSubscription, err := repo.UpdateSubscription(substituteSubscription)
	assert.Nil(t, err)
	assert.Equal(t, substituteSubscription.Id, updatedSubscription.Id)
	assert.Equal(t, substituteSubscription.Callback, updatedSubscription.Callback)
	assert.Equal(t, substituteSubscription.CallbackType, updatedSubscription.CallbackType)

	err = repo.DeleteSubscription(subscription.Id)
	assert.Nil(t, err)

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
	updatedSubscription, err := repo.UpdateSubscription(subscription)

	t.Logf("test update subscription error: %v", err)
	assert.NotNil(t, err)
	assert.Nil(t, updatedSubscription)
}

func TestRepository_GetSubscriptions(t *testing.T) {
	for i := 0; i < 1; i++ {
		subscription := &models.Subscription{
			Callback:     fmt.Sprintf("url: %d", i),
			CallbackType: models.HTTP,
			Filters: map[models.EventType]models.Filter{
				models.ANCHOR_EVENT: {Filtering: "filtering 1"},
				models.COMMIT_ENTRY: {Filtering: "filtering 2"},
				models.COMMIT_EVENT: {Filtering: "filtering 3"},
			},
		}
		repo.CreateSubscription(subscription)
	}

	// TODO dirty
	subscriptions, err := repo.GetAllSubscriptions()

	assert.Nil(t, err)

	fmt.Println("=========================================================================================")
	fmt.Println("")
	fmt.Println("=========================================================================================")

	fmt.Printf("SUBS: %v\n", subscriptions)

	for _, subscription := range subscriptions {
		fmt.Printf("%v\n", subscription)
		// r.DeleteSubscription(subscription.Id)
	}
}

func TestRepository_CreateSubscriptionFailure(t *testing.T) {
	// TODO fail with an inmemory db
	/*subscription := &models.Subscription{
		// max url size is 2083
		Callback:     "",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: struct{ Filtering models.GraphQL }{Filtering: "filtering 1"},
			models.COMMIT_ENTRY: struct{ Filtering models.GraphQL }{Filtering: "filtering 2"},
			models.COMMIT_EVENT: struct{ Filtering models.GraphQL }{Filtering: "filtering 3"},
		},
	}
	*/
	// createdSubscription, err := repo.CreateSubscription(subscription)

	// assert.NotNil(t, err)
	// assert.Nil(t, createdSubscription)
}
