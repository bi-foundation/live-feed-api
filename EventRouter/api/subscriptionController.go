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

	if err := validateSubscription(subscription); err != nil {
		log.Debug("invalid subscribe request %v: %v", subscription, err)
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	subscription, err := repository.SubscriptionRepository.CreateSubscription(subscription)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(fmt.Sprintf("failed to create subscription: %v", err)))
		return
	}
	respond(writer, subscription)
}

func updateSubscription(writer http.ResponseWriter, request *http.Request) {
	// swagger:route PUT /subscribe/{id} subscription UpdateSubscriptionRequest
	//
	// Update a subscription for receiving events from the api
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        200: UpdateSubscriptionResponse
	//        400: ApiError
	vars := mux.Vars(request)

	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	id := vars["subscriptionId"]
	if subscription.Id != id {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed("subscription id doesn't match"))
		return
	}

	subscription, err := repository.SubscriptionRepository.UpdateSubscription(subscription)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	respond(writer, subscription)
}

func unsubscribe(writer http.ResponseWriter, request *http.Request) {
	// swagger:route DELETE /subscribe/{id} subscription UnsubscribeRequest
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
	err := repository.SubscriptionRepository.DeleteSubscription(id)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}
}

func validateSubscription(subscription *models.Subscription) error {
	u, err := url.ParseRequestURI(subscription.Callback)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid callback url: %v", err)
	}

	switch subscription.CallbackType {
	case models.HTTP:
		if subscription.Credentials.AccessToken != "" || subscription.Credentials.BasicAuthUsername != "" || subscription.Credentials.BasicAuthPassword != "" {
			return fmt.Errorf("credentials are set but will not be used")
		}
	case models.BEARER_TOKEN:
		if subscription.Credentials.AccessToken == "" {
			return fmt.Errorf("access token required")
		}
	case models.BASIC_AUTH:
		if subscription.Credentials.BasicAuthUsername == "" || subscription.Credentials.BasicAuthPassword == "" {
			return fmt.Errorf("username and password are required")
		}
	default:
		return fmt.Errorf("unknown callback type: should be one of [%s,%s,%s]", models.HTTP, models.BASIC_AUTH, models.BEARER_TOKEN)
	}

	for eventType := range subscription.Filters {
		switch eventType {
		case models.ANCHOR_EVENT:
		case models.COMMIT_ENTRY:
		case models.COMMIT_CHAIN:
		case models.NODE_MESSAGE:
		case models.REVEAL_ENTRY:
		default:
			return fmt.Errorf("invalid event type: %s", eventType)
		}
	}

	return nil
}
