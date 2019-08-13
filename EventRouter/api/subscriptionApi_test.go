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
	Callback:     "http://url.nl/callback",
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
			Method:       "POST",
			content:      content(t, testSubscription),
			responseCode: http.StatusOK,
			assert:       assertSubscribe,
		},
		"unsubscribe": {
			URL:          "/unsubscribe/0",
			Method:       "DELETE",
			content:      nil,
			responseCode: http.StatusOK,
			assert:       assertEmptyResponse,
		},
		"subscribe-invalid": {
			URL:    "/subscribe",
			Method: "POST",
			content: content(t, &models.Subscription{
				Callback: "invalid url",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"unsubscribe-invalid": {
			URL:    "/unsubscribe/notfound",
			Method: "DELETE",
			content: content(t, &models.Subscription{
				Callback: "invalid url",
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-nothing": {
			URL:          "/subscribe",
			Method:       "POST",
			content:      nil,
			responseCode: http.StatusBadRequest,
			assert:       assertParseError,
		},
		"subscribe-something-else": {
			URL:          "/subscribe",
			Method:       "POST",
			content:      []byte(`{"message":"invalid object"}`),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"subscribe-wrong-method": {
			URL:          "/subscribe",
			Method:       "DELETE",
			content:      content(t, testSubscription),
			responseCode: http.StatusMethodNotAllowed,
			assert:       assertEmptyResponse,
		},
	}

	// init mock repository,
	mockStore := repository.InitMockRepository()
	mockStore.On("CreateSubscription").Return(testSubscription, nil).Once()
	mockStore.On("DeleteSubscription", "0").Return(nil).Once()
	mockStore.On("DeleteSubscription", "notfound").Return(fmt.Errorf("subscription not found")).Once()

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:8070%s", testCase.URL)
			request, err := http.NewRequest(testCase.Method, url, bytes.NewBuffer(testCase.content))

			assert.Nil(t, err, "failed to create request")

			response, err := http.DefaultClient.Do(request)

			assert.Nil(t, err, "failed to get response: %v", err)
			assert.Equal(t, testCase.responseCode, response.StatusCode)

			if response == nil {
				t.Fatalf("response incorrect")
			}

			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)

			t.Logf("%s response: %s", name, body)

			testCase.assert(t, body)
		})
	}
}

func assertSubscribe(t *testing.T, body []byte) {
	var result models.Subscription
	err := json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("unmarshalling failed: %v", err)
	}

	assert.Equal(t, testSubscription.Callback, result.Callback)
	assert.Equal(t, testSubscription.CallbackType, result.CallbackType)
	assert.EqualValues(t, testSubscription.Filters, result.Filters)
	assert.Equal(t, testSubscription.Credentials, result.Credentials)
	assert.NotNil(t, result.Id)
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
