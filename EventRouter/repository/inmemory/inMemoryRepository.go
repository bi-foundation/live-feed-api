package inmemory

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"strconv"
	"sync"
)

type inMemoryRepository struct {
	sync.RWMutex
	id int
	db []*models.SubscriptionContext
}

func New() *inMemoryRepository {
	return &inMemoryRepository{
		id: 0,
	}
}

func (repository *inMemoryRepository) CreateSubscription(subscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	repository.Lock()
	defer repository.Unlock()

	subscriptionContext.Subscription.Id = strconv.Itoa(repository.id)
	repository.db = append(repository.db, subscriptionContext)
	repository.id++
	log.Debug("stored subscription: %v", subscriptionContext)
	return subscriptionContext, nil
}

func (repository *inMemoryRepository) ReadSubscription(id string) (*models.SubscriptionContext, error) {
	_, subscriptionContext, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}

	log.Info("read subscription: %v", subscriptionContext)
	return subscriptionContext, nil
}

func (repository *inMemoryRepository) UpdateSubscription(substituteSubscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	index, subscriptionContext, err := repository.findSubscription(substituteSubscriptionContext.Subscription.Id)
	if err != nil {
		return nil, err
	}

	repository.Lock()
	defer repository.Unlock()
	log.Debug("update subscription: %v with: %v", subscriptionContext, substituteSubscriptionContext.Subscription)
	repository.db[index].Subscription.Callback = substituteSubscriptionContext.Subscription.Callback
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
		if subscriptionContext.Subscription.Id == id {
			return i, subscriptionContext, nil
		}
	}
	log.Debug("subscription not found: %s", id)
	return -1, nil, fmt.Errorf("failed to find subscription '%s'", id)
}

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

func (repository *inMemoryRepository) GetSubscriptions(eventType models.EventType) ([]*models.SubscriptionContext, error) {
	repository.RLock()
	defer repository.RUnlock()

	subscriptionContexts := repository.db[:0]
	for _, subscriptionContext := range repository.db {
		if _, ok := subscriptionContext.Subscription.Filters[eventType]; ok {
			subscriptionContexts = append(subscriptionContexts, subscriptionContext)
		}
	}

	return subscriptionContexts, nil
}
