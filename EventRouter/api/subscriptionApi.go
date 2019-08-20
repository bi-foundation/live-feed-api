// Live Feed API
//
// The live feed API is a service for receiving events from the factom blockchain. The API is connected to a factomd
// node. The received events will be emitted to the subscriptions in the API. Users can subscribe a callback url where
// able to receive different types of events.
//
//     Schemes: http
//     Host: localhost:8700
//     TODO change port
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//
// swagger:meta
package api

import (
	"encoding/json"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models/errors"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type SubscriptionApi interface {
	Start()
}

type api struct {
	address string
}

func NewSubscriptionApi(address string) SubscriptionApi {
	return &api{
		address: address,
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
	router.HandleFunc("/subscriptions", subscribe).Methods(http.MethodPost)
	router.HandleFunc("/subscriptions/{subscriptionId}", unsubscribe).Methods(http.MethodDelete)
	router.HandleFunc("/subscriptions/{subscriptionId}", getSubscription).Methods(http.MethodGet)
	router.HandleFunc("/subscriptions/{subscriptionId}", updateSubscription).Methods(http.MethodPut)
	router.HandleFunc("/swagger.json", swagger).Methods(http.MethodGet)
	router.Schemes("HTTP")

	go func() {
		log.Info("start subscription api at: %s", api.address)
		err := http.ListenAndServe(api.address, router)
		if err != nil {
			log.Error("%v", err)
		}
	}()
}

func swagger(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	http.ServeFile(writer, request, "swagger.json")
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
		responseError(writer, http.StatusInternalServerError, errors.NewInternalError("failed to write error"))
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
