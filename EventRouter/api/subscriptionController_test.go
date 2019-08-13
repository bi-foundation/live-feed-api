package api

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/models"
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
				Callback:     "http://test/callback",
				CallbackType: models.HTTP,
			},
			Error: nil,
		},
		"valid filters": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.HTTP,
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
				Callback: "invalid-callback",
			},
			Error: fmt.Errorf("invalid callback url: parse invalid-callback: invalid URI for request"),
		},
		"no callback type": {
			Subscription: &models.Subscription{
				Callback: "http://test/callback",
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid callback type": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: "WRONG",
			},
			Error: fmt.Errorf("unknown callback type: should be one of [HTTP,BASIC_AUTH,BEARER_TOKEN]"),
		},
		"invalid filters": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.HTTP,
				Filters: map[models.EventType]models.Filter{
					"invalid": {Filtering: "as"},
				},
			},
			Error: fmt.Errorf("invalid event type: invalid"),
		},
		"invalid http": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.HTTP,
				Credentials:  models.Credentials{AccessToken: "token"},
			},
			Error: fmt.Errorf("credentials are set but will not be used"),
		},
		"valid basic auth": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.BASIC_AUTH,
				Credentials:  models.Credentials{BasicAuthUsername: "test", BasicAuthPassword: "test"},
			},
			Error: nil,
		},
		"invalid basic auth": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.BASIC_AUTH,
			},
			Error: fmt.Errorf("username and password are required"),
		},
		"valid bearer token": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.BEARER_TOKEN,
				Credentials:  models.Credentials{AccessToken: "test"},
			},
			Error: nil,
		},
		"invalid bearer token": {
			Subscription: &models.Subscription{
				Callback:     "http://test/callback",
				CallbackType: models.BEARER_TOKEN,
			},
			Error: fmt.Errorf("access token required"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateSubscription(testCase.Subscription)
			assert.EqualValues(t, testCase.Error, err)
		})
	}
}
