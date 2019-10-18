package events

import (
	"bytes"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"

	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	log.SetLevel(log.D)
}

func TestHandleEvent(t *testing.T) {
	port := 23221
	subscriptionContexts := models.SubscriptionContexts{initSubscription("id", port, 0)}

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("GetActiveSubscriptions", models.EntryCommit).Return(subscriptionContexts, nil).Once()

	var eventsReceived int32 = 0
	factomEvent, expectedEvent := mockFactomEvent(t)

	startMockServer(t, port, &eventsReceived, nil, expectedEvent)

	configuration := &config.RouterConfig{MaxRetries: 3, RetryTimeout: 10}

	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(configuration, queue)
	router.Start()

	// test send event if an event is send through the channel
	queue <- factomEvent

	waitOnEventReceived(&eventsReceived, len(subscriptionContexts), 1*time.Minute)

	assert.Equal(t, int32(1), eventsReceived)
	mockStore.AssertExpectations(t)
}

func TestHandleEvents(t *testing.T) {
	// test sending 5 subscriptions to two different endpoints
	port1 := 23222
	port2 := 23223

	subscription1 := initSubscription("id1", port1, 0)
	subscription2 := initSubscription("id2", port2, 0)
	subscriptionContexts := models.SubscriptionContexts{
		subscription1, subscription1, subscription2, subscription2, subscription1,
	}

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("GetActiveSubscriptions", models.EntryCommit).Return(subscriptionContexts, nil).Once()

	var eventsReceived int32 = 0
	factomEvent, expectedEvent := mockFactomEvent(t)

	configuration := &config.RouterConfig{MaxRetries: 3, RetryTimeout: 1}
	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(configuration, queue)
	router.Start()

	// start two mock server on different port
	startMockServer(t, port1, &eventsReceived, nil, expectedEvent)
	startMockServer(t, port2, &eventsReceived, nil, expectedEvent)

	// test send event if an event is send through the channel
	queue <- factomEvent

	waitOnEventReceived(&eventsReceived, len(subscriptionContexts), 1*time.Minute)

	assert.Equal(t, int32(5), eventsReceived, "failed to deliver correct number of events: %d expected != %d received", 5, eventsReceived)
	mockStore.AssertExpectations(t)
}

func TestHandleFactomEvents(t *testing.T) {
	port := 23224
	subscriptionContexts := models.SubscriptionContexts{initSubscription("id", port, 0)}

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("GetActiveSubscriptions", models.EntryCommit).Return(subscriptionContexts, nil).Twice()

	var eventsReceived int32 = 0
	factomEvent, expectedEvent := mockFactomEvent(t)

	startMockServer(t, port, &eventsReceived, nil, expectedEvent)

	configuration := &config.RouterConfig{MaxRetries: 3, RetryTimeout: 1}
	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(configuration, queue)
	router.Start()

	// test send event if an event is send through the channel
	queue <- factomEvent
	queue <- factomEvent

	waitOnEventReceived(&eventsReceived, len(subscriptionContexts)*2, 1*time.Minute)

	assert.Equal(t, int32(2), eventsReceived)
}

func TestSend(t *testing.T) {
	port := 26231
	subscriptionID := "id"
	subscriptionContexts := models.SubscriptionContexts{initSubscription(subscriptionID, port, 0)}

	var eventsReceived int32 = 0
	factomEvent, event := mockFactomEvent(t)
	startMockServer(t, port, &eventsReceived, nil, event)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack)}
	eventRouter.send(subscriptionContexts, factomEvent)

	waitOnEventReceived(&eventsReceived, 1, 1*time.Minute)

	assert.Equal(t, int32(1), eventsReceived)
}

func TestSendEvents(t *testing.T) {
	port := 26232
	subscriptionID := "id"
	subscriptionContext := initSubscription(subscriptionID, port, 0)

	n := 3
	var eventsReceived int32 = 0
	_, event := mockFactomEvent(t)
	startMockServer(t, port, &eventsReceived, nil, event)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack)}

	// test send events
	for i := 0; i < n; i++ {
		eventRouter.sendEvent(subscriptionContext, event)
	}

	waitOnEventReceived(&eventsReceived, n, 1*time.Minute)

	assert.Equal(t, int32(n), eventsReceived)
}

func TestMapEventType(t *testing.T) {
	testCases := []models.EventType{models.ChainCommit, models.EntryCommit, models.EntryReveal, models.DirectoryBlockCommit, models.StateChange, models.ProcessMessage, models.NodeMessage}

	for _, testCase := range testCases {
		t.Run(string(testCase), func(t *testing.T) {
			eventType, err := mapEventType(createNewEvent(testCase))

			assert.Nil(t, err)
			assert.Equal(t, testCase, eventType)
		})
	}
}

func TestMapEventTypeUnknown(t *testing.T) {
	event := eventmessages.NewPopulatedFactomEvent(randomizer, true)
	event.Value = nil
	_, err := mapEventType(event)

	assert.Error(t, err)
}

func TestEmitEvent(t *testing.T) {
	port := 25231
	subscriptionID := "id"
	subscriptionContext := initSubscription(subscriptionID, port, 0)

	var eventsReceived int32 = 0
	_, event := mockFactomEvent(t)
	startMockServer(t, port, &eventsReceived, nil, event)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack)}
	eventRouter.emitQueue[subscriptionContext.Subscription.ID] = NewSubscriptionStack(subscriptionContext)
	eventRouter.emitQueue[subscriptionContext.Subscription.ID].Add(event)

	// test emit event
	eventRouter.emitEvent(subscriptionID)

	assert.Equal(t, int32(1), eventsReceived)
	assert.Equal(t, uint16(0), subscriptionContext.Failures)
	assert.Equal(t, models.Active, subscriptionContext.Subscription.SubscriptionStatus)
	assert.Equal(t, "", subscriptionContext.Subscription.SubscriptionInfo)
}

func TestEmitEventFailureRecover(t *testing.T) {
	// send the event to the wrong endpoint and simulate an external update during the timeout
	maxRetries := uint16(3)

	port := 25232
	subscriptionID := "fail-recover-id"
	subscriptionContext := initSubscription(subscriptionID, 999, 0)

	tmp := *subscriptionContext
	updatedSubscriptionContext := &tmp
	updatedSubscriptionContext.Failures = 1
	updatedSubscriptionContext.Subscription.CallbackURL = fmt.Sprintf("http://localhost:%[1]d/callback%[1]d", port)

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("ReadSubscription", subscriptionID).Return(updatedSubscriptionContext, nil).Once()
	mockStore.On("UpdateSubscription", subscriptionID).Return(nil, nil).Times(2)

	eventsReceived := int32(0)
	_, event := mockFactomEvent(t)
	startMockServer(t, port, &eventsReceived, nil, event)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack), maxRetries: maxRetries, retryTimeout: 1 * time.Millisecond}
	eventRouter.emitQueue[subscriptionContext.Subscription.ID] = NewSubscriptionStack(subscriptionContext)
	eventRouter.emitQueue[subscriptionContext.Subscription.ID].Add(event)

	// test emit event retry
	eventRouter.emitEvent(subscriptionID)

	assert.Equal(t, int32(1), eventsReceived)
	assert.Equal(t, uint16(1), subscriptionContext.Failures)
	assert.Equal(t, uint16(0), updatedSubscriptionContext.Failures)
	assert.Equal(t, models.Active, updatedSubscriptionContext.Subscription.SubscriptionStatus)
	assert.Equal(t, "", updatedSubscriptionContext.Subscription.SubscriptionInfo)

	mockStore.AssertExpectations(t)
}

func TestEmitEventFailureRetry(t *testing.T) {
	maxRetries := uint16(3)

	port := 25233
	subscriptionID := "id"
	subscriptionContext := initSubscription(subscriptionID, port, 0)

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("ReadSubscription", subscriptionID).Return(subscriptionContext, nil).Twice()
	mockStore.On("UpdateSubscription", subscriptionID).Return(nil, nil).Times(3)

	eventsReceived := int32(0)
	_, event := mockFactomEvent(t)
	authFailure := func(r *http.Request) bool { return false }
	startMockServer(t, port, &eventsReceived, authFailure, event)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack), maxRetries: maxRetries, retryTimeout: 1 * time.Millisecond}
	eventRouter.emitQueue[subscriptionContext.Subscription.ID] = NewSubscriptionStack(subscriptionContext)
	eventRouter.emitQueue[subscriptionContext.Subscription.ID].Add(event)

	// test emit event retry
	eventRouter.emitEvent(subscriptionID)

	assert.Equal(t, int32(3), eventsReceived)
	assert.Equal(t, maxRetries, subscriptionContext.Failures)
	assert.Equal(t, models.Suspended, subscriptionContext.Subscription.SubscriptionStatus)
	assert.NotEqual(t, "", subscriptionContext.Subscription.SubscriptionInfo)

	mockStore.AssertExpectations(t)
}

func TestEmitEventDBTimeout(t *testing.T) {
	maxRetries := uint16(3)

	port := 25234
	subscriptionID := "id"
	subscriptionContext := initSubscription(subscriptionID, port, 2)

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("ReadSubscription", subscriptionID).Return(subscriptionContext, fmt.Errorf("db timeout")).Once()
	mockStore.On("UpdateSubscription", subscriptionID).Return(nil, nil).Times(1)

	_, event := mockFactomEvent(t)

	eventRouter := &eventRouter{emitQueue: make(map[string]SubscriptionStack), maxRetries: maxRetries, retryTimeout: 1 * time.Millisecond}
	eventRouter.emitQueue[subscriptionContext.Subscription.ID] = NewSubscriptionStack(subscriptionContext)
	eventRouter.emitQueue[subscriptionContext.Subscription.ID].Add(event)

	// test emit event retry
	eventRouter.emitEvent(subscriptionID)

	assert.Equal(t, maxRetries, subscriptionContext.Failures)
	assert.Equal(t, models.Suspended, subscriptionContext.Subscription.SubscriptionStatus)
	assert.Contains(t, subscriptionContext.Subscription.SubscriptionInfo, "db timeout")

	mockStore.AssertExpectations(t)
}

func TestExecuteSendHTTP(t *testing.T) {
	port := 24231
	subscription := initSubscription("id", port, 0)

	_, event := mockFactomEvent(t)
	eventsReceived := int32(0)

	// start https server to receive event
	startMockServer(t, port, &eventsReceived, nil, event)

	// test send to the http endpoint
	err := executeSend(&subscription.Subscription, event)
	if err != nil {
		t.Fatalf("%v", err)
	}

	waitOnEventReceived(&eventsReceived, 1, 1*time.Minute)

	assert.Equal(t, int32(1), eventsReceived, "failed to deliver correct number of events: %d expected != %d received", 1, eventsReceived)
}

func TestExecuteSendHTTPSBearerToken(t *testing.T) {
	port := 24232
	accessToken := "access-token"
	subscription := &models.Subscription{
		CallbackURL:  fmt.Sprintf("https://localhost:%[1]d/callback%[1]d", port),
		CallbackType: models.BearerToken,
		Filters: map[models.EventType]models.Filter{
			models.ChainCommit: {Filtering: ""},
		},
		Credentials: models.Credentials{
			AccessToken: accessToken,
		},
	}

	_, event := mockFactomEvent(t)

	certFile, pkFile, cleanup := testSetupCertificateFiles(t)
	defer cleanup()

	eventsReceived := int32(0)

	// as the code makes use of the default client, we allow to ignore expired certificate
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	http.DefaultClient = &http.Client{Transport: transCfg}

	// start https server to receive event
	startMockTLSServer(t, port, certFile, pkFile, &eventsReceived, validateToken(accessToken), event)

	// test send to the http endpoint with oauth2
	err := executeSend(subscription, event)
	if err != nil {
		t.Fatalf("%v", err)
	}

	waitOnEventReceived(&eventsReceived, 1, 1*time.Minute)

	assert.Equal(t, int32(1), eventsReceived, "failed to deliver correct number of events: %d expected != %d received", 1, eventsReceived)
}

func TestExecuteSendHTTPSBasicAuth(t *testing.T) {
	port := 24233
	username := "usern@me"
	password := "passw0rd"
	subscription := &models.Subscription{
		CallbackURL:  fmt.Sprintf("https://localhost:%[1]d/callback%[1]d", port),
		CallbackType: models.BasicAuth,
		Filters: map[models.EventType]models.Filter{
			models.ChainCommit: {Filtering: ""},
		},
		Credentials: models.Credentials{
			BasicAuthUsername: username,
			BasicAuthPassword: password,
		},
	}

	_, event := mockFactomEvent(t)

	certFile, pkFile, cleanup := testSetupCertificateFiles(t)
	defer cleanup()

	eventsReceived := int32(0)

	// as the code makes use of the default client, we allow to ignore expired certificate
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	http.DefaultClient = &http.Client{Transport: transCfg}

	// start https server to receive event
	startMockTLSServer(t, port, certFile, pkFile, &eventsReceived, validateUsernamePassword(username, password), event)

	// test send to the http endpoint with oauth2
	err := executeSend(subscription, event)
	if err != nil {
		t.Fatalf("%v", err)
	}

	waitOnEventReceived(&eventsReceived, 1, 1*time.Minute)

	assert.Equal(t, int32(1), eventsReceived, "failed to deliver correct number of events: %d expected != %d received", 1, eventsReceived)
}

func TestExecuteSendNoEndpoint(t *testing.T) {
	subscription := &initSubscription("id", 999, 0).Subscription

	_, event := mockFactomEvent(t)

	// test send to http oauth2 endpoint
	err := executeSend(subscription, event)

	assert.Contains(t, err.Error(), "connect: connection refused")
}

func TestHandleSendFailure(t *testing.T) {
	mockStore := repository.InitMockRepository()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Times(3)
	mockStore.On("UpdateSubscription", "error").Return(nil, fmt.Errorf("db failure")).Once()

	maxRetries := uint16(3)
	testCases := map[string]*models.SubscriptionContext{
		"update": {
			Subscription: models.Subscription{ID: "id"},
			Failures:     0,
		},
		"update-max-failures": {
			Subscription: models.Subscription{ID: "id"},
			Failures:     2,
		},
		"db-failure": {
			Subscription: models.Subscription{ID: "error"},
			Failures:     0,
		},
		"max-failure": {
			Subscription: models.Subscription{ID: "id"},
			Failures:     maxRetries,
		},
	}
	eventRouter := &eventRouter{maxRetries: maxRetries}
	for name, subscriptionContext := range testCases {
		t.Run(name, func(t *testing.T) {
			eventRouter.handleSendFailure(subscriptionContext, "failed to deliver event")
		})

		mockStore.AssertCalled(t, "UpdateSubscription", subscriptionContext.Subscription.ID)
	}
	mockStore.AssertExpectations(t)
}

func TestHandleSendSuccessful(t *testing.T) {
	mockStore := repository.InitMockRepository()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Once()
	mockStore.On("UpdateSubscription", "error").Return(nil, fmt.Errorf("db failure")).Once()

	testCases := map[string]*models.SubscriptionContext{
		"update": {
			Subscription: models.Subscription{ID: "id"},
			Failures:     1,
		},
		"update-nothing": {
			Subscription: models.Subscription{ID: "id"},
			Failures:     0,
		},
		"failure": {
			Subscription: models.Subscription{ID: "error"},
			Failures:     1,
		},
	}
	eventRouter := &eventRouter{}
	for name, subscriptionContext := range testCases {
		t.Run(name, func(t *testing.T) {
			eventRouter.handleSendSuccessful(subscriptionContext)
		})
	}
	mockStore.AssertExpectations(t)
}

func initSubscription(subscriptionID string, port int, failures uint16) *models.SubscriptionContext {
	return &models.SubscriptionContext{
		Subscription: models.Subscription{
			ID:                 subscriptionID,
			CallbackURL:        fmt.Sprintf("http://localhost:%[1]d/callback%[1]d", port),
			CallbackType:       models.HTTP,
			SubscriptionStatus: models.Active,
			Filters: map[models.EventType]models.Filter{
				models.ChainCommit: {Filtering: ""},
			},
		},
		Failures: failures,
	}
}

func mockFactomEvent(t testing.TB) (*eventmessages.FactomEvent, []byte) {
	factomEvent := eventmessages.NewPopulatedFactomEvent(randomizer, true)
	expectedEvent, err := json.Marshal(factomEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	return factomEvent, expectedEvent
}

func startMockServer(t testing.TB, port int, eventsReceived *int32, authenticationValidation func(r *http.Request) bool, expectedEvent []byte) {
	startMockTLSServer(t, port, "", "", eventsReceived, authenticationValidation, expectedEvent)
}

func startMockTLSServer(t testing.TB, port int, certFile string, pkFile string, eventsReceived *int32, validAuthentication func(r *http.Request) bool, expectedEvent []byte) {
	// start http server to receive event
	http.HandleFunc(fmt.Sprintf("/callback%d", port), func(w http.ResponseWriter, r *http.Request) {
		defer atomic.AddInt32(eventsReceived, 1)

		// validate auth
		if validAuthentication != nil && !validAuthentication(r) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		// verify body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error("failed to read body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// assert body
		assert.EqualValues(t, expectedEvent, body)

		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "\t")
		if err != nil {
			t.Error("failed to parse json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// fmt.Printf("> %s\n", string(prettyJSON.Bytes()))
		w.WriteHeader(http.StatusOK)
	})

	if certFile != "" && pkFile != "" {
		go http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, pkFile, nil)
	} else {
		go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}

}

func waitOnEventReceived(eventsReceived *int32, n int, timeLimit time.Duration) {
	deadline := time.Now().Add(timeLimit)
	for int(atomic.LoadInt32(eventsReceived)) != n && time.Now().Before(deadline) {
		fmt.Print(".")
		time.Sleep(100 * time.Millisecond)
	}
}

func validateToken(accessToken string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		authorization := r.Header.Get("authorization")
		token := strings.TrimPrefix(strings.ToLower(authorization), "bearer ")
		return strings.EqualFold(token, accessToken)
	}
}

func validateUsernamePassword(username string, password string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			fmt.Printf("basic auth validation failed: %s, %s, %t\n", user, pass, ok)
			return false
		}
		return true
	}
}

// an arbitrary self-signed certificate, generated with
// `openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout cert.pem -out cert.pem`
var pkey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDVBUw40q0zpF/zWzwBf0GFkXmnkw+YCNTiV8l7mso1DCv/VTYM
cqtvy0g2KNBV7SFLC+NHuxJkNOAtJ8Fxx1EpeIw5A3KeCRNb4lo6ecAkuDLiPYGO
qgAqjj8QmhmZA68qTIuWGYM1FTtUK3wO4wrHnqHEjs3cWNghmby6AgLHVQIDAQAB
AoGAcy5GJINlu4KpjwBJ1dVlLD+YtA9EY0SDN0+YVglARKasM4dzjg+CuxQDm6U9
4PgzBE0NO3/fVedxP3k7k7XeH73PosaxjWpfMawXR3wSLFKJBwxux/8gNdzeGRHN
X1sYsJ70WiZLFOAPQ9jctF1ejUP6fpLHsti6ZHQj/R1xqBECQQDrHxmpMoviQL6n
4CBR4HvlIRtd4Qr21IGEXtbjIcC5sgbkfne6qhqdv9/zxsoiPTi0859cr704Mf3y
cA8LZ8c3AkEA5+/KjSoqgzPaUnvPZ0p9TNx6odxMsd5h1AMIVIbZPT6t2vffCaZ7
R0ffim/KeWfoav8u9Cyz8eJpBG6OHROT0wJBAML54GLCCuROAozePI8JVFS3NqWM
OHZl1R27NAHYfKTBMBwNkCYYZ8gHVKUoZXktQbg1CyNmjMhsFIYWTTONFNMCQFsL
eBld2f5S1nrWex3y0ajgS4tKLRkNUJ2m6xgzLwepmRmBf54MKgxbHFb9dx+dOFD4
Bvh2q9RhqhPBSiwDyV0CQBxN3GPbaa8V7eeXBpBYO5Evy4VxSWJTpgmMDtMH+RUp
9eAJ8rUyhZ2OaElg1opGCRemX98s/o2R5JtzZvOx7so=
-----END RSA PRIVATE KEY-----
`

var cert = `-----BEGIN CERTIFICATE-----
MIIDXDCCAsWgAwIBAgIJAJqbbWPZgt0sMA0GCSqGSIb3DQEBBQUAMH0xCzAJBgNV
BAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEPMA0G
A1UEChMGV2ViLmdvMRcwFQYDVQQDEw5NaWNoYWVsIEhvaXNpZTEfMB0GCSqGSIb3
DQEJARYQaG9pc2llQGdtYWlsLmNvbTAeFw0xMzA0MDgxNjIzMDVaFw0xNDA0MDgx
NjIzMDVaMH0xCzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2Fu
IEZyYW5jaXNjbzEPMA0GA1UEChMGV2ViLmdvMRcwFQYDVQQDEw5NaWNoYWVsIEhv
aXNpZTEfMB0GCSqGSIb3DQEJARYQaG9pc2llQGdtYWlsLmNvbTCBnzANBgkqhkiG
9w0BAQEFAAOBjQAwgYkCgYEA1QVMONKtM6Rf81s8AX9BhZF5p5MPmAjU4lfJe5rK
NQwr/1U2DHKrb8tINijQVe0hSwvjR7sSZDTgLSfBccdRKXiMOQNyngkTW+JaOnnA
JLgy4j2BjqoAKo4/EJoZmQOvKkyLlhmDNRU7VCt8DuMKx56hxI7N3FjYIZm8ugIC
x1UCAwEAAaOB4zCB4DAdBgNVHQ4EFgQURizcvrgUl8yhIEQvJT/1b5CzV8MwgbAG
A1UdIwSBqDCBpYAURizcvrgUl8yhIEQvJT/1b5CzV8OhgYGkfzB9MQswCQYDVQQG
EwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xDzANBgNV
BAoTBldlYi5nbzEXMBUGA1UEAxMOTWljaGFlbCBIb2lzaWUxHzAdBgkqhkiG9w0B
CQEWEGhvaXNpZUBnbWFpbC5jb22CCQCam21j2YLdLDAMBgNVHRMEBTADAQH/MA0G
CSqGSIb3DQEBBQUAA4GBAGBPoVCReGMO1FrsIeVrPV/N6pSK7H3PLdxm7gmmvnO9
K/LK0OKIT7UL3eus+eh0gt0/Tv/ksq4nSIzXBLPKyPggLmpC6Agf3ydNTpdLQ23J
gWrxykqyLToIiAuL+pvC3Jv8IOPIiVFsY032rOqcwSGdVUyhTsG28+7KnR6744tM
-----END CERTIFICATE-----
`

func testSetupCertificateFiles(t *testing.T) (string, string, func()) {
	certificates := make([]tls.Certificate, 1)
	var err error
	certificates[0], err = tls.X509KeyPair([]byte(cert), []byte(pkey))

	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certFile, cleanCertFile := testTempFile(t, "cert", cert)
	pkFile, cleanPKFile := testTempFile(t, "pk", pkey)

	cleanup := func() {
		cleanCertFile()
		cleanPKFile()
	}
	return certFile, pkFile, cleanup
}

func testTempFile(t *testing.T, prefix string, content string) (string, func()) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatalf("error creating temp file %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("error write to temp file %v", err)
	}
	cleanFile := func() { os.Remove(tmpFile.Name()) }

	return tmpFile.Name(), cleanFile
}

func BenchmarkHandleEvents(b *testing.B) {
	// setup test
	port := 21221
	subscriptionContext := initSubscription("id", port, 0)
	subscriptionContexts := models.SubscriptionContexts{subscriptionContext}

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("GetActiveSubscriptions", models.EntryCommit).Return(subscriptionContexts, nil)
	mockStore.On("ReadSubscription", "id").Return(subscriptionContext, nil)
	mockStore.On("UpdateSubscription", "id").Return(subscriptionContext, nil)

	eventsReceived := int32(0)
	factomEvent, expectedEvent := mockFactomEvent(b)

	startMockServer(b, port, &eventsReceived, nil, expectedEvent)

	configuration := &config.RouterConfig{MaxRetries: 3, RetryTimeout: 1}
	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(configuration, queue)
	router.Start()

	// benchmark table
	benchmarks := map[string]struct{ SubscriptionContexts models.SubscriptionContexts }{
		"send event": {models.SubscriptionContexts{subscriptionContext}},
	}

	// run benchmark
	for name := range benchmarks {
		b.Run(name, func(b *testing.B) {
			eventsReceived = 0
			n := 0
			for ; n < b.N; n++ {
				queue <- factomEvent
			}
			waitOnEventReceived(&eventsReceived, n, time.Duration(b.N)*time.Second)
			assert.Equal(b, int32(n), eventsReceived)
		})
	}
}
