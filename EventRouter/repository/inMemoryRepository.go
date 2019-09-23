package repository

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	"strconv"
	"sync"
)

type inMemoryRepository struct {
	sync.RWMutex
	id int
	db []*models.SubscriptionContext
}

// NewInMemoryRepository create a new in memory repository
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		id: 0,
	}
}

// CreateSubscription create a subscription
func (repository *inMemoryRepository) CreateSubscription(subscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	repository.Lock()
	defer repository.Unlock()

	subscriptionContext.Subscription.ID = strconv.Itoa(repository.id)
	repository.db = append(repository.db, subscriptionContext)
	repository.id++
	log.Debug("stored subscription: %v", subscriptionContext)
	return subscriptionContext, nil
}

// ReadSubscription read a subscription
func (repository *inMemoryRepository) ReadSubscription(id string) (*models.SubscriptionContext, error) {
	_, subscriptionContext, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}

	log.Info("read subscription: %v", subscriptionContext)
	return subscriptionContext, nil
}

// UpdateSubscription update a subscription
func (repository *inMemoryRepository) UpdateSubscription(substituteSubscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	index, subscriptionContext, err := repository.findSubscription(substituteSubscriptionContext.Subscription.ID)
	if err != nil {
		return nil, err
	}

	repository.Lock()
	defer repository.Unlock()
	log.Debug("update subscription: %v with: %v", subscriptionContext, substituteSubscriptionContext.Subscription)
	repository.db[index].Subscription.CallbackURL = substituteSubscriptionContext.Subscription.CallbackURL
	repository.db[index].Subscription.CallbackType = substituteSubscriptionContext.Subscription.CallbackType
	repository.db[index].Subscription.SubscriptionStatus = substituteSubscriptionContext.Subscription.SubscriptionStatus
	repository.db[index].Subscription.SubscriptionInfo = substituteSubscriptionContext.Subscription.SubscriptionInfo
	repository.db[index].Subscription.Credentials.AccessToken = substituteSubscriptionContext.Subscription.Credentials.AccessToken
	repository.db[index].Subscription.Credentials.BasicAuthUsername = substituteSubscriptionContext.Subscription.Credentials.BasicAuthUsername
	repository.db[index].Subscription.Credentials.BasicAuthPassword = substituteSubscriptionContext.Subscription.Credentials.BasicAuthPassword
	repository.db[index].Subscription.Filters = substituteSubscriptionContext.Subscription.Filters
	return substituteSubscriptionContext, err
}

func (repository *inMemoryRepository) findSubscription(id string) (int, *models.SubscriptionContext, error) {
	repository.RLock()
	defer repository.RUnlock()

	for i, subscriptionContext := range repository.db {
		if subscriptionContext.Subscription.ID == id {
			return i, subscriptionContext, nil
		}
	}
	log.Debug("subscription not found: %s", id)
	return -1, nil, errors.NewSubscriptionNotFound(id)
}

// DeleteSubscription delete a subscription
func (repository *inMemoryRepository) DeleteSubscription(id string) error {
	index, _, err := repository.findSubscription(id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	repository.Lock()
	defer repository.Unlock()
	repository.db = append(repository.db[:index], repository.db[index+1:]...)
	log.Debug("deleted subscription: %s", id)
	return nil
}

// GetActiveSubscriptions retrieve all active subscriptions
func (repository *inMemoryRepository) GetActiveSubscriptions(eventType models.EventType) ([]*models.SubscriptionContext, error) {
	repository.RLock()
	defer repository.RUnlock()

	subscriptionContexts := repository.db[:0]
	for _, subscriptionContext := range repository.db {
		if _, ok := subscriptionContext.Subscription.Filters[eventType]; ok && subscriptionContext.Subscription.SubscriptionStatus == models.Active {
			subscriptionContexts = append(subscriptionContexts, subscriptionContext)
		}
	}

	return subscriptionContexts, nil
}
