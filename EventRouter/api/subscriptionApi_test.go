package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/api/errors"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func init() {
	log.SetLevel(log.D)

	server := NewSubscriptionApi(":8070")
	server.Start()
	time.Sleep(1 * time.Second)
}

var testSubscription = &models.Subscription{
	Id:           "id",
	Callback:     "http://url/callback",
	CallbackType: models.HTTP,
	Filters: map[models.EventType]models.Filter{
		models.COMMIT_CHAIN: {
			Filtering: "filtering 1",
		},
		models.COMMIT_ENTRY: {
			Filtering: "filtering 2",
		},
	},
}

var suspendedSubscription = &models.Subscription{
	Callback:           "http://url/callback",
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
			URL:          "/subscribe",
			Method:       http.MethodPost,
			content:      content(t, testSubscription),
			responseCode: http.StatusOK,
			assert:       assertTestSubscribe,
		},
		"subscribe-invalid": {
			URL:    "/subscribe",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				Callback: "invalid url",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-nothing": {
			URL:          "/subscribe",
			Method:       http.MethodPost,
			content:      nil,
			responseCode: http.StatusBadRequest,
			assert:       assertParseError,
		},
		"subscribe-something-else": {
			URL:          "/subscribe",
			Method:       http.MethodPost,
			content:      []byte(`{"message":"invalid object"}`),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-suspended": {
			URL:          "/subscribe",
			Method:       http.MethodPost,
			content:      content(t, suspendedSubscription),
			responseCode: http.StatusOK,
			assert:       assertSuspendedSubscribe,
		},
		"subscribe-invalid-status": {
			URL:    "/subscribe",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				Callback:           "http://url/callback/suspended",
				CallbackType:       models.HTTP,
				SubscriptionStatus: "invalid status",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-db-fail": {
			URL:    "/subscribe",
			Method: http.MethodPost,
			content: content(t, &models.Subscription{
				Callback:     "http://url/callback/internal/error",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusInternalServerError,
			assert:       assertInternalError,
		},
		"get-subscription": {
			URL:          "/subscribe/id",
			Method:       http.MethodGet,
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertGetSubscribe,
		},
		"get-subscription-unknown": {
			URL:          "/subscribe/unknown",
			Method:       http.MethodGet,
			content:      nil,
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-subscription": {
			URL:          "/subscribe/id",
			Method:       http.MethodPut,
			content:      content(t, testSubscription),
			responseCode: http.StatusOK,
			assert:       assertTestSubscribe,
		},
		"update-unknown-id ": {
			URL:          "/subscribe/unknown-id",
			Method:       http.MethodPut,
			content:      content(t, testSubscription),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-id-mismatch ": {
			URL:    "/subscribe/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Id:           "id-mismatch",
				Callback:     "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-subscription-invalid-url": {
			URL:    "/subscribe/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Id:           "id",
				Callback:     "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-invalid-subscription ": {
			URL:    "/subscribe/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				Callback:     "http://url/test",
				CallbackType: "invalid",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"unsubscribe": {
			URL:          "/subscribe/0",
			Method:       http.MethodDelete,
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertEmptyResponse,
		},
		"unsubscribe not found": {
			URL:          "/subscribe/",
			Method:       http.MethodDelete,
			content:      nil,
			responseCode: http.StatusNotFound,
			assert:       assertNotFound,
		},
		"unsubscribe-invalid": {
			URL:    "/subscribe/notfound",
			Method: http.MethodDelete,
			content: content(t, &models.Subscription{
				Callback: "invalid url",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-wrong-method": {
			URL:          "/subscribe",
			Method:       http.MethodDelete,
			content:      content(t, testSubscription),
			responseCode: http.StatusMethodNotAllowed,
			assert:       assertEmptyResponse,
		},
	}

	// init mock repository,
	mockStore := repository.InitMockRepository()
	mockStore.On("CreateSubscription", "http://url/callback").Return(nil, nil).Twice()
	mockStore.On("CreateSubscription", "http://url/callback/internal/error").Return(nil, fmt.Errorf("something failed")).Once()
	mockStore.On("ReadSubscription", "id").Return(suspendedSubscriptionContext, nil).Once()
	mockStore.On("ReadSubscription", "unknown").Return(&models.SubscriptionContext{}, fmt.Errorf("subscription not found")).Once()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Once()
	mockStore.On("DeleteSubscription", "0").Return(nil).Once()
	mockStore.On("DeleteSubscription", "notfound").Return(fmt.Errorf("subscription not found")).Once()

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:8070%s", testCase.URL)
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

	assert.Equal(t, expected.Callback, actual.Callback)
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

func parseApiBody(t *testing.T, body []byte) errors.ApiError {
	var result errors.ApiError
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
