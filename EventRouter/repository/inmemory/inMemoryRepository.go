package inmemory

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"strconv"
)

type inMemoryRepository struct {
	id int
	db []*models.Subscription
}

func New() *inMemoryRepository {
	return &inMemoryRepository{
		id: 0,
	}
}

func (repository *inMemoryRepository) CreateSubscription(subscription *models.Subscription) (*models.Subscription, error) {
	subscription.Id = strconv.Itoa(repository.id)
	repository.db = append(repository.db, subscription)
	repository.id++
	log.Debug("stored subscription: %v", subscription)
	return subscription, nil
}

func (repository *inMemoryRepository) ReadSubscription(id string) (*models.Subscription, error) {
	_, subscription, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}

	log.Info("read subscription: %v", subscription)
	return subscription, nil
}

func (repository *inMemoryRepository) UpdateSubscription(substitute *models.Subscription) (*models.Subscription, error) {
	index, subscription, err := repository.findSubscription(substitute.Id)
	if err != nil {
		return nil, err
	}

	log.Debug("update subscription: %v with: %v", subscription, substitute)
	repository.db[index].Callback = substitute.Callback
	repository.db[index].CallbackType = substitute.CallbackType
	repository.db[index].Filters = substitute.Filters
	return substitute, err
}

func (repository *inMemoryRepository) DeleteSubscription(id string) error {
	index, _, err := repository.findSubscription(id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}
	repository.db = append(repository.db[:index], repository.db[index+1:]...)
	log.Debug("deleted subscription: %s", id)
	return nil
}

func (repository *inMemoryRepository) findSubscription(id string) (int, *models.Subscription, error) {
	for i, subscription := range repository.db {
		if subscription.Id == id {
			return i, subscription, nil
		}
	}
	log.Debug("subscription not found: %s", id)
	return -1, nil, fmt.Errorf("failed to find subscription '%s'", id)
}

func (repository *inMemoryRepository) GetSubscriptions(eventType models.EventType) ([]*models.Subscription, error) {
	subscriptions := repository.db[:0]
	for _, subscription := range repository.db {
		if _, ok := subscription.Filters[eventType]; ok {
			subscriptions = append(subscriptions, subscription)
		}
	}

	return subscriptions, nil
}
