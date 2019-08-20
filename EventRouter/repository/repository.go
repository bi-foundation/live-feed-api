package repository

import "github.com/FactomProject/live-api/EventRouter/models"

type Repository interface {
	CreateSubscription(subscription *models.SubscriptionContext) (*models.SubscriptionContext, error)
	ReadSubscription(id string) (*models.SubscriptionContext, error)
	UpdateSubscription(subscription *models.SubscriptionContext) (*models.SubscriptionContext, error)
	DeleteSubscription(id string) error
	GetActiveSubscriptions(models.EventType) ([]*models.SubscriptionContext, error)
}
