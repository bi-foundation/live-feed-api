package inmemory

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/FactomProject/live-api/EventRouter/log"
	"strconv"
)

type InMemoryRepository struct {
	id int
	db []models.Subscription
}

func (repository *InMemoryRepository) CreateSubscription(subscription *models.Subscription) (*models.Subscription, error) {
	subscription.Id = strconv.Itoa(repository.id)
	repository.db = append(repository.db, *subscription)
	repository.id++
	log.Debug("stored subscription: %v", subscription)
	return subscription, nil
}

func (repository *InMemoryRepository) ReadSubscription(id string) (*models.Subscription, error) {
	_, subscription, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}

	log.Info("read subscription: %v", subscription)
	return subscription, nil
}

func (repository *InMemoryRepository) UpdateSubscription(id string, substitute *models.Subscription) (*models.Subscription, error) {
	index, subscription, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}

	log.Debug("update subscription: %v with: %v", subscription, substitute)
	repository.db[index].Callback = substitute.Callback
	substitute.Id = id
	return substitute, err
}

func (repository *InMemoryRepository) DeleteSubscription(id string) (*models.Subscription, error) {
	index, _, err := repository.findSubscription(id)
	if err != nil {
		return nil, err
	}
	subscription := repository.db[index]
	repository.db = append(repository.db[:index], repository.db[index+1:]...)
	log.Debug("deleted subscription: %v", subscription)
	return &subscription, nil
}

func (repository *InMemoryRepository) findSubscription(id string) (int, *models.Subscription, error) {
	for i, subscription := range repository.db {
		if subscription.Id == id {
			return i, &subscription, nil
		}
	}
	log.Debug("subscription not found: %s", id)
	return -1, nil, fmt.Errorf("failed to find subscription '%s'", id)
}

func (repository *InMemoryRepository) ReadSubscriptions() []models.Subscription {
	// TODO filter on events
	return repository.db
}
