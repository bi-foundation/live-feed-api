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
				CallbackURL:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.Active,
			},
			Error: nil,
		},
		"valid filters": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.Active,
				Filters: map[models.EventType]models.Filter{
					models.BlockCommit:       {Filtering: "filtering"},
					models.EntryRegistration: {Filtering: "filtering"},
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
				CallbackURL: "invalid-callback",
			},
			Error: fmt.Errorf("invalid callback url: parse invalid-callback: invalid URI for request"),
		},
		"no callback type": {
			Subscription: &models.Subscription{
				CallbackURL: "http://test/callback",
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid callback type": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       "WRONG",
				SubscriptionStatus: models.Active,
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid filters": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.Active,
				Filters: map[models.EventType]models.Filter{
					"invalid": {Filtering: "as"},
				},
			},
			Error: fmt.Errorf("invalid event type: invalid"),
		},
		"invalid http": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.HTTP,
				SubscriptionStatus: models.Active,
				Credentials:        models.Credentials{AccessToken: "token"},
			},
			Error: fmt.Errorf("credentials are set but will not be used"),
		},
		"valid basic auth": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.BasicAuth,
				SubscriptionStatus: models.Active,
				Credentials:        models.Credentials{BasicAuthUsername: "test", BasicAuthPassword: "test"},
			},
			Error: nil,
		},
		"invalid basic auth": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.BasicAuth,
				SubscriptionStatus: models.Active,
			},
			Error: fmt.Errorf("username and password are required"),
		},
		"valid bearer token": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.BearerToken,
				SubscriptionStatus: models.Active,
				Credentials:        models.Credentials{AccessToken: "test"},
			},
			Error: nil,
		},
		"invalid bearer token": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
				CallbackType:       models.BearerToken,
				SubscriptionStatus: models.Active,
			},
			Error: fmt.Errorf("access token required"),
		},
		"invalid status": {
			Subscription: &models.Subscription{
				CallbackURL:        "http://test/callback",
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
