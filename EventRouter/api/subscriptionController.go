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
	// swagger:route POST /subscriptions subscription CreateSubscriptionRequest
	//
	// Subscribe an application to receive events.
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        201: CreateSubscriptionResponse
	//        400: ApiError
	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	// ignore user input info message
	subscription.SubscriptionInfo = ""
	if subscription.SubscriptionStatus == "" {
		subscription.SubscriptionStatus = models.ACTIVE
	}

	if err := validateSubscription(subscription); err != nil {
		log.Debug("invalid subscribe request %v: %v", subscription, err)
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: *subscription,
		Failures:     0,
	}

	subscriptionContext, err := repository.SubscriptionRepository.CreateSubscription(subscriptionContext)
	if err != nil {
		responseError(writer, http.StatusInternalServerError, errors.NewInternalError(fmt.Sprintf("failed to store subscription: %v", err)))
		return
	}
	respond(writer, subscriptionContext.Subscription)
}

func updateSubscription(writer http.ResponseWriter, request *http.Request) {
	// swagger:route PUT /subscriptions/{id} subscription UpdateSubscriptionRequest
	//
	// Update a subscription for receiving events.
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

	// ignore user input message
	subscription.SubscriptionInfo = ""
	if subscription.SubscriptionStatus == "" {
		subscription.SubscriptionStatus = models.ACTIVE
	}

	if err := validateSubscription(subscription); err != nil {
		log.Debug("invalid subscribe request %v: %v", subscription, err)
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	subscriptionContext := &models.SubscriptionContext{
		Subscription: *subscription,
		Failures:     0,
	}

	subscriptionContext, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	respond(writer, subscriptionContext.Subscription)
}

func getSubscription(writer http.ResponseWriter, request *http.Request) {
	// swagger:route GET /subscriptions/{id} subscription GetSubscriptionRequest
	//
	// Return a subscription with the given id.
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        200: GetSubscriptionResponse
	//        400: ApiError
	vars := mux.Vars(request)

	id := vars["subscriptionId"]

	subscriptionContext, err := repository.SubscriptionRepository.ReadSubscription(id)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	respond(writer, subscriptionContext.Subscription)
}

func unsubscribe(writer http.ResponseWriter, request *http.Request) {
	// swagger:route DELETE /subscriptions/{id} subscription DeleteSubscriptionRequest
	//
	// Unsubscribe a subscription from receiving events.
	//
	// Consumes:
	//   - application/json
	//
	// Produces:
	//   - application/json
	//
	// Responses:
	//        200: DeleteSubscriptionResponse
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
	u, err := url.ParseRequestURI(subscription.CallbackUrl)
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
		case models.PROCESS_MESSAGE:
		case models.NODE_MESSAGE:
		case models.REVEAL_ENTRY:
		default:
			return fmt.Errorf("invalid event type: %s", eventType)
		}
	}

	switch subscription.SubscriptionStatus {
	case models.ACTIVE:
	case models.SUSPENDED:
	default:
		return fmt.Errorf("unknown subscription status: should be one of [%s, %s]", models.ACTIVE, models.SUSPENDED)
	}

	return nil
}
