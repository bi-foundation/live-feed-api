package api

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateSubscription(t *testing.T) {
	testCases := map[string]struct {
		Subscription *models.Subscription
		Error        error
	}{
		"valid": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.ACTIVE,
			},
			Error: nil,
		},
		"valid filters": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.ACTIVE,
				Filters: map[models.EventType]models.Filter{
					models.ANCHOR_EVENT: {Filtering: "filtering"},
					models.COMMIT_ENTRY: {Filtering: "filtering"},
				},
			},
			Error: nil,
		},
		"empty url": {
			Subscription: &models.Subscription{},
			Error:        fmt.Errorf("invalid callback url: parse : empty url"),
		},
		"invalid url": {
			Subscription: &models.Subscription{
				CallbackUrl: "invalid-callback",
			},
			Error: fmt.Errorf("invalid callback url: parse invalid-callback: invalid URI for request"),
		},
		"no callback type": {
			Subscription: &models.Subscription{
				CallbackUrl: "http://test/callback",
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid callback type": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       "WRONG",
				SubscriptionStatus: models.ACTIVE,
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid filters": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.ACTIVE,
				Filters: map[models.EventType]models.Filter{
					"invalid": {Filtering: "as"},
				},
			},
			Error: fmt.Errorf("invalid event type: invalid"),
		},
		"invalid http": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.ACTIVE,
				Credentials:        models.Credentials{AccessToken: "token"},
			},
			Error: fmt.Errorf("credentials are set but will not be used"),
		},
		"valid basic auth": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.BASIC_AUTH,
				SubscriptionStatus: models.ACTIVE,
				Credentials:        models.Credentials{BasicAuthUsername: "test", BasicAuthPassword: "test"},
			},
			Error: nil,
		},
		"invalid basic auth": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.BASIC_AUTH,
				SubscriptionStatus: models.ACTIVE,
			},
			Error: fmt.Errorf("username and password are required"),
		},
		"valid bearer token": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.BEARER_TOKEN,
				SubscriptionStatus: models.ACTIVE,
				Credentials:        models.Credentials{AccessToken: "test"},
			},
			Error: nil,
		},
		"invalid bearer token": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.BEARER_TOKEN,
				SubscriptionStatus: models.ACTIVE,
			},
			Error: fmt.Errorf("access token required"),
		},
		"invalid status": {
			Subscription: &models.Subscription{
				CallbackUrl:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: "Something different",
			},
			Error: fmt.Errorf("unknown subscription status: should be one of [ACTIVE, SUSPENDED]"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateSubscription(testCase.Subscription)
			assert.EqualValues(t, testCase.Error, err)
		})
	}
}
