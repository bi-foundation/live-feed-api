package repository

import "github.com/FactomProject/live-api/EventRouter/models"

type Repository interface {
	CreateSubscription(subscription *models.Subscription) (*models.Subscription, error)
	ReadSubscription(id string) (*models.Subscription, error)
	UpdateSubscription(subscription *models.Subscription) (*models.Subscription, error)
	DeleteSubscription(id string) error
	GetSubscriptions(models.EventType) ([]*models.Subscription, error)
}
