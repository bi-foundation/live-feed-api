package repository

import "github.com/FactomProject/live-feed-api/EventRouter/models"

// Repository for storing and retrieving subscriptions
type Repository interface {
	CreateSubscription(subscription *models.SubscriptionContext) (*models.SubscriptionContext, error)
	ReadSubscription(id string) (*models.SubscriptionContext, error)
	UpdateSubscription(subscription *models.SubscriptionContext) (*models.SubscriptionContext, error)
	DeleteSubscription(id string) error
	GetActiveSubscriptions(models.EventType) ([]*models.SubscriptionContext, error)
}
