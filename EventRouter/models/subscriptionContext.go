package models

// SubscriptionContext information about the subscription without exporting the information
type SubscriptionContext struct {
	Subscription Subscription `json:"subscription"`

	Failures int `json:"failures"`
}
