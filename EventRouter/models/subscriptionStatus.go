package models

// SubscriptionStatus status of the subscription
type SubscriptionStatus string

// different subscriptions status
const (
	Active SubscriptionStatus = "ACTIVE"

	Suspended SubscriptionStatus = "SUSPENDED"
)
