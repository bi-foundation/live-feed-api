package errors

import (
	"fmt"
)

// SubscriptionNotFound to handle subscription not found error on the type level
type SubscriptionNotFound struct {
	error
}

// NewSubscriptionNotFound create a new subscription not found error
func NewSubscriptionNotFound(id string) SubscriptionNotFound {
	return SubscriptionNotFound{fmt.Errorf("subscription '%s' not found", id)}
}
