package models

// status of subscription: [ACTIVE, SUSPENDED]
// Normally a subscription is active. When events fail to be delivered the subscription will be suspended. The subscription can become active again by updating the subscription.
// swagger:model SubscriptionStatuss
type SubscriptionStatus string

const (
	ACTIVE SubscriptionStatus = "ACTIVE"

	SUSPENDED SubscriptionStatus = "SUSPENDED"
)
