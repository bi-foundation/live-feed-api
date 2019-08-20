package events

import (
	"bytes"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/gogo/protobuf/types"
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

func TestSendEvent(t *testing.T) {
	accessToken := "access-token"

	// init test cases
	testCases := map[string]struct {
		EndpointPostfix          string
		CallbackType             models.CallbackType
		Credentials              models.Credentials
		AuthenticationValidation func(r *http.Request) bool
	}{
		"basic http": {
			CallbackType: models.HTTP,
		},
		"bearer token": {
			CallbackType: models.BEARER_TOKEN,
			Credentials: models.Credentials{
				AccessToken: accessToken,
			},
			AuthenticationValidation: validateToken(accessToken),
		},
		"invalid bearer token": {
			CallbackType: models.BEARER_TOKEN,
			Credentials: models.Credentials{
				AccessToken: accessToken,
			},
			AuthenticationValidation: validateToken("invalid"),
		},
		"basic auth": {
			CallbackType: models.BASIC_AUTH,
			Credentials: models.Credentials{
				BasicAuthUsername: "username",
				BasicAuthPassword: "password",
			},
			AuthenticationValidation: validateUsernamePassword("username", "password"),
		},
		"basic invalid auth": {
			CallbackType: models.BASIC_AUTH,
			Credentials: models.Credentials{
				BasicAuthUsername: "username",
				BasicAuthPassword: "password",
			},
			AuthenticationValidation: validateUsernamePassword("incorrect", "password"),
		},
		"url not found": {
			EndpointPostfix: "/not/here",
			CallbackType:    models.HTTP,
		},
	}

	mockStore := repository.InitMockRepository()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Times(3)

	// init data
	factomEvent := mockAnchorEvent()
	event, err := json.Marshal(factomEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	index := 0
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			index++
			port := 23232 + index
			var eventsReceived int32 = 0
			subscriptionContext := &models.SubscriptionContext{
				Subscription: models.Subscription{
					Id:           "id",
					CallbackUrl:  fmt.Sprintf("http://localhost:%[1]d/callback%[1]d%s", port, testCase.EndpointPostfix),
					CallbackType: testCase.CallbackType,
					Filters: map[models.EventType]models.Filter{
						models.COMMIT_CHAIN: {Filtering: ""},
					},
					Credentials: testCase.Credentials,
				},
			}

			// start server to receive events
			startMockServer(t, port, &eventsReceived, testCase.AuthenticationValidation, event)

			// test send to http oauth2 endpoint
			sendEvent(subscriptionContext, event)
		})
	}
	mockStore.AssertExpectations(t)
}

func TestHTTPSEndpoint(t *testing.T) {
	accessToken := "access-token"
	subscriptionContexts := []*models.SubscriptionContext{
		{
			Subscription: models.Subscription{
				CallbackUrl:  "https://localhost:23232/callback23232",
				CallbackType: models.BEARER_TOKEN,
				Filters: map[models.EventType]models.Filter{
					models.COMMIT_CHAIN: {Filtering: ""},
				},
				Credentials: models.Credentials{
					AccessToken: accessToken,
				},
			},
		},
	}
	factomEvent := mockAnchorEvent()
	expectedEvent, err := json.Marshal(factomEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	certFile, pkFile, cleanup := testSetupCertificateFiles(t)
	defer cleanup()

	var eventsReceived int32 = 0

	// as the code makes use of the default client, we allow to ignore expired certificate
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	http.DefaultClient = &http.Client{Transport: transCfg}

	// start https server to receive event
	startMockTLSServer(t, 23232, certFile, pkFile, &eventsReceived, nil, expectedEvent)

	// test send to http oauth2 endpoint
	err = send(subscriptionContexts, factomEvent)
	assert.Nil(t, err)

	waitOnEventReceived(&eventsReceived, len(subscriptionContexts), 1*time.Minute)
}

// test sending 5 subscriptions to two different endpoints
func TestHandleEvents(t *testing.T) {
	subscription1 := models.Subscription{
		CallbackUrl: "http://localhost:23222/callback23222",
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: ""},
		},
	}
	subscription2 := models.Subscription{
		CallbackUrl: "http://localhost:23223/callback23223",
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: ""},
		},
	}
	subscriptionContexts := []*models.SubscriptionContext{
		{Subscription: subscription1},
		{Subscription: subscription2},
		{Subscription: subscription2},
		{Subscription: subscription1},
		{Subscription: subscription1},
	}

	// init mock repository
	mockStore := repository.InitMockRepository()
	mockStore.On("GetSubscriptions", models.ANCHOR_EVENT).Return(subscriptionContexts, nil).Once()

	var eventsReceived int32 = 0
	factomEvent := mockAnchorEvent()
	expectedEvent, err := json.Marshal(factomEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(queue)
	router.Start()

	// start two mock server on different porst
	startMockServer(t, 23222, &eventsReceived, nil, expectedEvent)
	startMockServer(t, 23223, &eventsReceived, nil, expectedEvent)

	// test send event if an event is send through the channel
	queue <- factomEvent

	waitOnEventReceived(&eventsReceived, len(subscriptionContexts), 1*time.Minute)
}

func TestSendEventFailure(t *testing.T) {
	mockStore := repository.InitMockRepository()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Twice()
	mockStore.On("UpdateSubscription", "error").Return(nil, fmt.Errorf("db failure")).Once()

	testCases := map[string]*models.SubscriptionContext{
		"update": {
			Subscription: models.Subscription{Id: "id"},
			Failures:     0,
		},
		"update-max-failures": {
			Subscription: models.Subscription{Id: "id"},
			Failures:     2,
		},
		"failure": {
			Subscription: models.Subscription{Id: "error"},
			Failures:     0,
		},
	}

	for name, subscriptionContext := range testCases {
		t.Run(name, func(t *testing.T) {
			sendEventFailure(subscriptionContext, "failed to deliver event")
		})

		mockStore.AssertCalled(t, "UpdateSubscription", subscriptionContext.Subscription.Id)
	}
	mockStore.AssertExpectations(t)
}

func TestSendEventSuccessful(t *testing.T) {
	mockStore := repository.InitMockRepository()
	mockStore.On("UpdateSubscription", "id").Return(nil, nil).Once()
	mockStore.On("UpdateSubscription", "error").Return(nil, fmt.Errorf("db failure")).Once()

	testCases := map[string]*models.SubscriptionContext{
		"update": {
			Subscription: models.Subscription{Id: "id"},
			Failures:     1,
		},
		"update-nothing": {
			Subscription: models.Subscription{Id: "id"},
			Failures:     0,
		},
		"failure": {
			Subscription: models.Subscription{Id: "error"},
			Failures:     1,
		},
	}

	for name, subscriptionContext := range testCases {
		t.Run(name, func(t *testing.T) {
			sendEventSuccessful(subscriptionContext)
		})
	}
	mockStore.AssertExpectations(t)
}

func startMockServer(t *testing.T, port int, eventsReceived *int32, authenticationValidation func(r *http.Request) bool, expectedEvent []byte) {
	startMockTLSServer(t, port, "", "", eventsReceived, authenticationValidation, expectedEvent)
}

func startMockTLSServer(t *testing.T, port int, certFile string, pkFile string, eventsReceived *int32, validAuthentication func(r *http.Request) bool, expectedEvent []byte) {
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

func mockAnchorEvent() *eventmessages.FactomEvent {
	now := time.Now()
	testHash := []byte("12345678901234567890123456789012")
	return &eventmessages.FactomEvent{
		EventSource: 0,
		Value: &eventmessages.FactomEvent_AnchorEvent{
			AnchorEvent: &eventmessages.AnchoredEvent{
				DirectoryBlock: &eventmessages.DirectoryBlock{
					Header: &eventmessages.DirectoryBlockHeader{
						BodyMerkleRoot: &eventmessages.Hash{
							HashValue: testHash,
						},
						PreviousKeyMerkleRoot: &eventmessages.Hash{
							HashValue: testHash,
						},
						PreviousFullHash: &eventmessages.Hash{
							HashValue: testHash,
						},
						Timestamp:  &types.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())},
						BlockCount: 456,
					},
					Entries: []*eventmessages.Entry{
						{
							ChainID: &eventmessages.Hash{
								HashValue: testHash,
							},
							KeyMerkleRoot: &eventmessages.Hash{
								HashValue: testHash,
							},
						}, {
							ChainID: &eventmessages.Hash{
								HashValue: testHash,
							},
							KeyMerkleRoot: &eventmessages.Hash{
								HashValue: testHash,
							},
						},
					},
				},
			},
		},
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
