package models

type SubscriptionContext struct {
	Subscription Subscription `json:"subscription"`

	Failures int `json:"failures"`
}
