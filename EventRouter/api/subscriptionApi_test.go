package api

import (
	"bytes"
	"crypto/tls"
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
	"os"
	"testing"
	"time"
)

const (
	httpPort  = 8700
	httpsPort = 8701
	basePath  = "/live/feed"
)

func init() {
	log.SetLevel(log.D)
}

var testSubscription = &models.Subscription{
	CallbackURL:  "http://url/callback",
	CallbackType: models.HTTP,
	Filters: map[models.EventType]models.Filter{
		models.ChainRegistration: {
			Filtering: "filtering 1",
		},
		models.EntryRegistration: {
			Filtering: "filtering 2",
		},
	},
}

var suspendedSubscription = &models.Subscription{
	CallbackURL:        "http://url/callback",
	CallbackType:       models.HTTP,
	SubscriptionStatus: models.Suspended,
	SubscriptionInfo:   "read only information",
}

var suspendedSubscriptionContext = &models.SubscriptionContext{
	Subscription: *suspendedSubscription,
	Failures:     0,
}

func TestSubscriptionAPI(t *testing.T) {
	// start http subscription api
	configuration := &config.SubscriptionConfig{
		BindAddress: "",
		Port:        httpPort,
		BasePath:    basePath,
		Scheme:      "HTTP",
	}
	startAPI(configuration)

	testSubscriptionAPI(t, "http", httpPort)
}

func TestTLSSubscriptionAPI(t *testing.T) {
	// start https subscription api
	certFile, pkFile, cleanup := testSetupCertificateFiles(t)
	defer cleanup()

	configuration := &config.SubscriptionConfig{
		BindAddress:     "",
		Port:            httpsPort,
		BasePath:        basePath,
		Scheme:          "HTTPS",
		CertificateFile: certFile,
		PrivateKeyFile:  pkFile,
	}
	startAPI(configuration)

	testSubscriptionAPI(t, "https", httpsPort)
}

func testSubscriptionAPI(t *testing.T, scheme string, port int) {
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
				CallbackURL: "invalid url",
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
				CallbackURL:        "http://url/callback/suspended",
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
				CallbackURL:  "http://url/callback/internal/error",
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
				ID:           "id",
				CallbackURL:  "http://url/callback",
				CallbackType: models.HTTP,
				Filters: map[models.EventType]models.Filter{
					models.ChainRegistration: {
						Filtering: "filtering 1",
					},
					models.EntryRegistration: {
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
				ID:           "id-mismatch",
				CallbackURL:  "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-subscription-invalid-url": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				ID:           "id",
				CallbackURL:  "invalid-url",
				CallbackType: models.HTTP,
			}),
			responseCode: http.StatusBadRequest,
			assert:       assertInvalidRequestError,
		},
		"update-invalid-subscription ": {
			URL:    "/subscriptions/id",
			Method: http.MethodPut,
			content: content(t, &models.Subscription{
				CallbackURL:  "http://url/test",
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
			url := fmt.Sprintf("%s://localhost:%d%s%s", scheme, port, basePath, testCase.URL)

			request, err := http.NewRequest(testCase.Method, url, bytes.NewBuffer(testCase.content))
			assert.Nil(t, err, "failed to create request")

			transCfg := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
			}
			client := &http.Client{Transport: transCfg}

			response, err := client.Do(request)

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

func startAPI(configuration *config.SubscriptionConfig) {
	// use info from swagger to init will be called to register the swagger which is provided through an endpoint
	info := docs.SwaggerInfo
	log.Info("start %s api %s %s", configuration.Scheme, info.Title, info.Version)

	// Start the new server at random port
	server := NewSubscriptionAPI(configuration)
	server.Start()

	time.Sleep(1 * time.Second)
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

	assert.Equal(t, expected.CallbackURL, actual.CallbackURL)
	assert.Equal(t, expected.CallbackType, actual.CallbackType)
	assert.EqualValues(t, expected.Filters, actual.Filters)
	assert.Equal(t, expected.Credentials, actual.Credentials)
	if expected.SubscriptionStatus != "" {
		assert.Equal(t, expected.SubscriptionStatus, actual.SubscriptionStatus)
	} else {
		assert.Equal(t, models.Active, actual.SubscriptionStatus)
	}
	assert.Equal(t, expected.SubscriptionInfo, actual.SubscriptionInfo)
	assert.NotNil(t, actual.ID)
}

func assertEmptyResponse(t *testing.T, body []byte) {
	assert.Equal(t, "", string(body))
}

func assertNotEmptyResponse(t *testing.T, body []byte) {
	assert.NotEqual(t, "", string(body))
}

func assertParseError(t *testing.T, body []byte) {
	result := parseAPIBody(t, body)

	assert.Equal(t, "parse error", result.Message)
	assert.Equal(t, errors.NewParseError().Code, result.Code)
}

func assertInvalidRequestError(t *testing.T, body []byte) {
	result := parseAPIBody(t, body)

	assert.Equal(t, "invalid request", result.Message)
	assert.Equal(t, errors.NewInvalidRequest().Code, result.Code)
}

func assertInternalError(t *testing.T, body []byte) {
	result := parseAPIBody(t, body)

	assert.Equal(t, "internal error", result.Message)
	assert.Equal(t, errors.NewInternalError("").Code, result.Code)
}

func assertNotFound(t *testing.T, body []byte) {
	assert.Equal(t, "404 page not found\n", string(body))
}

func parseAPIBody(t *testing.T, body []byte) models.APIError {
	var result models.APIError
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
