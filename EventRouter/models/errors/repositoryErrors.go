package errors

import (
	"fmt"
)

type SubscriptionNotFound struct {
	error
}

func NewSubscriptionNotFound(id string) SubscriptionNotFound {
	return SubscriptionNotFound{fmt.Errorf("subscription '%s' not found", id)}
}
