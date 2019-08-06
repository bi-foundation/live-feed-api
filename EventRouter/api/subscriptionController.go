package api

import (
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/api/errors"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

func subscribe(writer http.ResponseWriter, request *http.Request) {
	// swagger:route POST /subscribe subscription SubscriptionRequest
	//
	// Subscribe a new application to receive an event
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        200: SubscriptionResponse
	//        400: ApiError
	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	u, err := url.ParseRequestURI(subscription.Callback)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Debug("invalid subscribe request %v: %v", subscription, err)
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed("invalid callback url"))
		return
	}

	subscription, err = repository.SubscriptionRepository.CreateSubscription(subscription)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(fmt.Sprintf("failed to create subscription: %v", err)))
		return
	}
	respond(writer, subscription)
}

func unsubscribe(writer http.ResponseWriter, request *http.Request) {
	// swagger:route DELETE /unsubscribe/{id} subscription UnsubscribeRequest
	//
	// Unsubscribe an application from receiving events from the api
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        200: UnsubscriptionResponse
	//        400: ApiError

	vars := mux.Vars(request)

	id := vars["subscriptionId"]
	subscription, err := repository.SubscriptionRepository.DeleteSubscription(id)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}
	respond(writer, subscription)
}
