package models

import (
	"fmt"
	"strings"
)

// SubscriptionContext information about the subscription without exporting the information
type SubscriptionContext struct {
	Subscription Subscription `json:"subscription"`

	Failures uint16 `json:"failures"`
}

// SubscriptionContexts are a list of subscription contexts
type SubscriptionContexts []*SubscriptionContext

func (subscriptionContexts SubscriptionContexts) String() string {
	var builder strings.Builder
	fmt.Fprint(&builder, "[")
	for i, subscriptionContext := range subscriptionContexts {
		if i > 0 {
			fmt.Fprint(&builder, ", ")
		}
		fmt.Fprintf(&builder, "%v", subscriptionContext)
	}
	fmt.Fprint(&builder, "]")
	return builder.String()
}
