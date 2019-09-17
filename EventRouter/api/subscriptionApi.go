package api

// @title Live Feed API
// @version 0.1
// @description The live feed API is a service for receiving events from the factom blockchain. The API is connected to a factomd node. The received events will be emitted to the subscriptions in the API. Users can subscribe a callback url where able to receive different types of events.

// @license.name MIT
// @license.url http://opensource.org/licenses/MIT
//
// @host localhost:8700
// @BasePath /live/feed/v0.1
// @schemes http https

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	"github.com/gorilla/mux"
	"github.com/swaggo/swag"
	"net/http"
	"strings"
	"time"
)

// SubscriptionAPI is API endpoint that users allow to register an callback to receive events
type SubscriptionAPI interface {
	Start()
}

type api struct {
	apiConfig *config.SubscriptionConfig
}

// NewSubscriptionAPI create a new SubscriptionAPI with a configuration
func NewSubscriptionAPI(apiConfig *config.SubscriptionConfig) SubscriptionAPI {
	return &api{
		apiConfig: apiConfig,
	}
}

func logInterceptor(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		f.ServeHTTP(w, r)
		log.Debug("%s\t%s\t%s\n", r.Method, r.RequestURI, time.Since(start))
	})
}

func (api *api) Start() {
	router := mux.NewRouter()
	router.Use(logInterceptor)
	router.Schemes(api.apiConfig.Schemes...)

	subscriptionRouter := router.PathPrefix(api.apiConfig.BasePath).Subrouter()
	subscriptionRouter.HandleFunc("/subscriptions", subscribe).Methods(http.MethodPost)
	subscriptionRouter.HandleFunc("/subscriptions/{subscriptionId}", unsubscribe).Methods(http.MethodDelete)
	subscriptionRouter.HandleFunc("/subscriptions/{subscriptionId}", getSubscription).Methods(http.MethodGet)
	subscriptionRouter.HandleFunc("/subscriptions/{subscriptionId}", updateSubscription).Methods(http.MethodPut)
	subscriptionRouter.HandleFunc("/swagger.json", swagger).Methods(http.MethodGet)

	go func() {
		address := fmt.Sprintf("%s:%d", api.apiConfig.BindAddress, api.apiConfig.Port)
		log.Info("start subscription api at: [%s]://%s%s", strings.Join(api.apiConfig.Schemes, ", "), address, api.apiConfig.BasePath)
		err := http.ListenAndServe(address, router)
		if err != nil {
			log.Error("failed to start subscription api: %v", err)
		}
	}()
}

func swagger(writer http.ResponseWriter, _ *http.Request) {
	swagger, err := swag.ReadDoc()
	if err != nil {
		log.Error("failed to read swagger: %v", err)
		responseError(writer, http.StatusInternalServerError, errors.NewInternalError("failed to read swagger"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_, err = fmt.Fprint(writer, swagger)
	if err != nil {
		log.Error("failed to write swagger: %v", err)
		responseError(writer, http.StatusInternalServerError, errors.NewInternalError("failed to write swagger"))
		return
	}
}

func decode(writer http.ResponseWriter, request *http.Request, v interface{}) bool {
	err := json.NewDecoder(request.Body).Decode(v)
	if err != nil {
		log.Error("failed to parse request: %v", request.Body)
		responseError(writer, http.StatusBadRequest, errors.NewParseError())
		return true
	}
	return false
}

func responseError(writer http.ResponseWriter, statusCode int, error interface{}) {
	writer.WriteHeader(statusCode)
	err := json.NewEncoder(writer).Encode(error)
	if err != nil {
		log.Error("failed to write error '%v': %v", error, err)
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(writer, errors.NewInternalError("failed to write error"))
	}
}

func respond(writer http.ResponseWriter, data interface{}) {
	respondCode(writer, http.StatusOK, data)
}

func respondCode(writer http.ResponseWriter, statusCode int, data interface{}) {
	writer.WriteHeader(statusCode)
	writer.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		log.Error("failed to write response '%v': %v", data, err)
		responseError(writer, http.StatusInternalServerError, errors.NewInternalError("failed to write response"))
	}
}
