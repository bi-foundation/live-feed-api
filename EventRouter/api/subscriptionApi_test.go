package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
	docs "github.com/FactomProject/live-feed-api/EventRouter/swagger"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	port = 8700
	basePath = "/live/feed"
)

func init() {
	log.SetLevel(log.D)

	configuration := &config.SubscriptionConfig{
		BindAddress: "",
		Port:        port,
		BasePath:    basePath,
		Schemes:     []string{"HTTP"},
	}

	// use info from swagger to init will be called to register the swagger which is provided through an endpoint
	info := docs.SwaggerInfo
	log.Info("start %s %s", info.Title, info.Version)

	// Start the new server at random port
	server := NewSubscriptionApi(configuration)
	server.Start()
	time.Sleep(1 * time.Second)
}

var testSubscription = &models.Subscription{
	CallbackUrl:  "http://url/callback",
	CallbackType: models.HTTP,
	Filters: map[models.EventType]models.Filter{
		models.CHAIN_REGISTRATION: {
			Filtering: "filtering 1",
		},
		models.ENTRY_REGISTRATION: {
			Filtering: "filtering 2",
		},
	},
}

var suspendedSubscription = &models.Subscription{
	CallbackUrl:        "http://url/callback",
	CallbackType:       models.HTTP,
	SubscriptionStatus: models.SUSPENDED,
	SubscriptionInfo:   "read only information",
}

var suspendedSubscriptionContext = &models.SubscriptionContext{
	Subscription: *suspendedSubscription,
	Failures:     0,
}

func TestSubscriptionApi(t *testing.T) {
	testCases := map[string]struct {
		URL          string
		Method       string
		content      []byte
		responseCode int
		assert       func(*testing.T, []byte)
	}{
		"subscribe": {
			URL:          "/subscriptions",
			Method:       http.MethodPost,
			content:      content(t, testSubscription),
			responseCode: http.StatusCreated,
			assert:       assertTestSubscribe,
		},
		"subscribe-invalid": {
			URL:    "/subscriptions",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				CallbackUrl: "invalid url",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-nothing": {
			URL:          "/subscriptions",
			Method:       http.MethodPost,
			content:      nil,
			responseCode: http.StatusBadRequest,
			assert:       assertParseError,
		},
		"subscribe-something-else": {
			URL:          "/subscriptions",
			Method:       http.MethodPost,
			content:      []byte(`{"message":"invalid object"}`),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-suspended": {
			URL:          "/subscriptions",
			Method:       http.MethodPost,
			content:      content(t, suspendedSubscription),
			responseCode: http.StatusCreated,
			assert:       assertSuspendedSubscribe,
		},
		"subscribe-invalid-status": {
			URL:    "/subscriptions",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				CallbackUrl:        "http://url/callback/suspended",
				CallbackType:       models.HTTP,
				SubscriptionStatus: "invalid status",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-db-fail": {
			URL:    "/subscriptions",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				CallbackUrl:  "http://url/callback/internal/error",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusInternalServerError,
			assert:       assertInternalError,
		},
		"get-subscription": {
			URL:          "/subscriptions/id",
			Method:       http.MethodGet,
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertGetSubscribe,
		},
		"get-subscription-unknown": {
			URL:          "/subscriptions/unknown",
			Method:       http.MethodGet,
			content:      nil,
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-subscription": {
			URL:          "/subscriptions/id",
			Method:       http.MethodPut,
			content:      content(t, testSubscription),
			responseCode: http.StatusOK,
			assert:       assertTestSubscribe,
		},
		"update-subscription-with-id-in-body": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Id:           "id",
				CallbackUrl:  "http://url/callback",
				CallbackType: models.HTTP,
				Filters: map[models.EventType]models.Filter{
					models.CHAIN_REGISTRATION: {
						Filtering: "filtering 1",
					},
					models.ENTRY_REGISTRATION: {
						Filtering: "filtering 2",
					},
				},
			}),
			responseCode: http.StatusOK,
			assert:       assertTestSubscribe,
		},
		"update-unknown-id ": {
			URL:          "/subscriptions/unknown-id",
			Method:       http.MethodPut,
			content:      content(t, testSubscription),
			responseCode: http.StatusNotFound,
			assert:       assertInvalidRequestError,
		},
		"update-id-mismatch": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Id:           "id-mismatch",
				CallbackUrl:  "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-subscription-invalid-url": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Id:           "id",
				CallbackUrl:  "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-invalid-subscription ": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				CallbackUrl:  "http://url/test",
				CallbackType: "invalid",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"unsubscribe": {
			URL:          "/subscriptions/0",
			Method:       http.MethodDelete,
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertEmptyResponse,
		},
		"unsubscribe no id": {
			URL:          "/subscriptions/",
			Method:       http.MethodDelete,
			content:      nil,
			responseCode: http.StatusNotFound,
			assert:       assertNotFound,
		},
		"subscribe-wrong-method": {
			URL:          "/subscriptions",
			Method:       http.MethodDelete,
			content:      content(t, testSubscription),
			responseCode: http.StatusMethodNotAllowed,
			assert:       assertEmptyResponse,
		},
		"swagger": {
			URL:          "/swagger.json",
			Method:       http.MethodGet,
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertNotEmptyResponse,
		},
	}

	// init mock repository,
	mockStore := repository.InitMockRepository()
	mockStore.On("CreateSubscription", "http://url/callback").Return(nil, nil).Twice()
	mockStore.On("CreateSubscription", "http://url/callback/internal/error").Return(nil, fmt.Errorf("something failed")).Once()
	mockStore.On("ReadSubscription", "id").Return(suspendedSubscriptionContext, nil).Once()
	mockStore.On("ReadSubscription", "unknown").Return(&models.SubscriptionContext{}, errors.NewSubscriptionNotFound("unknown")).Once()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Twice()
	mockStore.On("UpdateSubscription", "unknown-id").Return(nil, errors.NewSubscriptionNotFound("unknown")).Once()
	mockStore.On("DeleteSubscription", "0").Return(nil).Once()

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:%d%s%s", port, basePath, testCase.URL)
			request, err := http.NewRequest(testCase.Method, url, bytes.NewBuffer(testCase.content))

			assert.Nil(t, err, "failed to create request")

			response, err := http.DefaultClient.Do(request)

			assert.Nil(t, err, "failed to get response: %v", err)
			if response == nil {
				t.Fatalf("response incorrect")
			}
			assert.Equal(t, testCase.responseCode, response.StatusCode)

			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)

			t.Logf("%s response: %s", name, body)

			testCase.assert(t, body)
		})
	}

	mockStore.AssertExpectations(t)
}

func assertTestSubscribe(t *testing.T, body []byte) {
	testSubscription.SubscriptionInfo = ""
	assertSubscribe(t, testSubscription, body)
}

func assertGetSubscribe(t *testing.T, body []byte) {
	assertSubscribe(t, suspendedSubscription, body)
}

func assertSuspendedSubscribe(t *testing.T, body []byte) {
	expectedSubscription := models.Subscription(*suspendedSubscription)
	expectedSubscription.SubscriptionInfo = ""
	assertSubscribe(t, &expectedSubscription, body)
}

func assertSubscribe(t *testing.T, expected *models.Subscription, body []byte) {
	var actual models.Subscription
	err := json.Unmarshal(body, &actual)
	if err != nil {
		t.Fatalf("unmarshalling failed: %v", err)
	}

	assert.Equal(t, expected.CallbackUrl, actual.CallbackUrl)
	assert.Equal(t, expected.CallbackType, actual.CallbackType)
	assert.EqualValues(t, expected.Filters, actual.Filters)
	assert.Equal(t, expected.Credentials, actual.Credentials)
	if expected.SubscriptionStatus != "" {
		assert.Equal(t, expected.SubscriptionStatus, actual.SubscriptionStatus)
	} else {
		assert.Equal(t, models.ACTIVE, actual.SubscriptionStatus)
	}
	assert.Equal(t, expected.SubscriptionInfo, actual.SubscriptionInfo)
	assert.NotNil(t, actual.Id)
}

func assertEmptyResponse(t *testing.T, body []byte) {
	assert.Equal(t, "", string(body))
}

func assertNotEmptyResponse(t *testing.T, body []byte) {
	assert.NotEqual(t, "", string(body))
}

func assertParseError(t *testing.T, body []byte) {
	result := parseApiBody(t, body)

	assert.Equal(t, "parse error", result.Message)
	assert.Equal(t, errors.NewParseError().Code, result.Code)
}

func assertInvalidRequestError(t *testing.T, body []byte) {
	result := parseApiBody(t, body)

	assert.Equal(t, "invalid request", result.Message)
	assert.Equal(t, errors.NewInvalidRequest().Code, result.Code)
}

func assertInternalError(t *testing.T, body []byte) {
	result := parseApiBody(t, body)

	assert.Equal(t, "internal error", result.Message)
	assert.Equal(t, errors.NewInternalError("").Code, result.Code)
}

func assertNotFound(t *testing.T, body []byte) {
	assert.Equal(t, "404 page not found\n", string(body))
}

func parseApiBody(t *testing.T, body []byte) models.ApiError {
	var result models.ApiError
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("unmarshalling failed: %v", err)
	}
	return result
}

func content(t *testing.T, v interface{}) []byte {
	content, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshaling failed: %v", err)
	}
	return content
}
