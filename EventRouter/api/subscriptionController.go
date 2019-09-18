package api

import (
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

// @Summary subscribe an application
// @Description Subscribe an application to receive events.
// @Accept  json
// @Produce  json
// @Param subscription body models.Subscription true "subscription to be created"
// @Success 201 {object} models.Subscription "subscription created"
// @Failure 400 {object} models.APIError
// @Router /subscriptions [post]
func subscribe(writer http.ResponseWriter, request *http.Request) {
	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	// ignore user input info message
	subscription.SubscriptionInfo = ""
	if subscription.SubscriptionStatus == "" {
		subscription.SubscriptionStatus = models.Active
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

	respondCode(writer, http.StatusCreated, subscriptionContext.Subscription)
}

// @Summary update a subscription
// @Description Update a subscription for receiving events. Updating the subscription can be used to change the endpoint url, adjust the filtering, add of remove the subscription for event types. When the subscription failed to deliver and got SUSPENDED, the endpoint can used to re-ACTIVATE the subscription.
// @Accept  json
// @Produce  json
// @Param id path int true "subscription id"
// @Param subscription body models.Subscription true "subscription to be updated"
// @Success 200 {object} models.Subscription "subscription updated"
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Router /subscriptions/{id} [put]
func updateSubscription(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	id := vars["subscriptionId"]
	if subscription.ID != "" && subscription.ID != id {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed("subscription id doesn't match"))
		return
	}
	subscription.ID = id

	// ignore user input message
	subscription.SubscriptionInfo = ""
	if subscription.SubscriptionStatus == "" {
		subscription.SubscriptionStatus = models.Active
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
	if notFoundError, ok := err.(errors.SubscriptionNotFound); ok {
		responseError(writer, http.StatusNotFound, errors.NewInvalidRequestDetailed(notFoundError.Error()))
		return
	} else if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	respond(writer, subscriptionContext.Subscription)
}

// @Summary get a subscription
// @Description Return a subscription with the given id.
// @Accept  json
// @Produce  json
// @Param id path int true "subscription id"
// @Success 200 {object} models.Subscription "subscription"
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Router /subscriptions/{id} [get]
func getSubscription(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	id := vars["subscriptionId"]

	subscriptionContext, err := repository.SubscriptionRepository.ReadSubscription(id)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}

	respond(writer, subscriptionContext.Subscription)
}

// @Summary delete a subscription
// @Description Unsubscribe an application from receiving events.
// @Accept  json
// @Produce  json
// @Param id path int true "subscription id"
// @Success 200 "subscription deleted"
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Router /subscriptions/{id} [delete]
func unsubscribe(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	id := vars["subscriptionId"]
	err := repository.SubscriptionRepository.DeleteSubscription(id)
	if err != nil {
		responseError(writer, http.StatusBadRequest, errors.NewInvalidRequestDetailed(err.Error()))
		return
	}
}

func validateSubscription(subscription *models.Subscription) error {
	u, err := url.ParseRequestURI(subscription.CallbackURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid callback url: %v", err)
	}

	switch subscription.CallbackType {
	case models.HTTP:
		if subscription.Credentials.AccessToken != "" || subscription.Credentials.BasicAuthUsername != "" || subscription.Credentials.BasicAuthPassword != "" {
			return fmt.Errorf("credentials are set but will not be used")
		}
	case models.BearerToken:
		if subscription.Credentials.AccessToken == "" {
			return fmt.Errorf("access token required")
		}
	case models.BasicAuth:
		if subscription.Credentials.BasicAuthUsername == "" || subscription.Credentials.BasicAuthPassword == "" {
			return fmt.Errorf("username and password are required")
		}
	default:
		return fmt.Errorf("unknown callback type: should be one of [%s,%s,%s]", models.HTTP, models.BasicAuth, models.BearerToken)
	}

	for eventType := range subscription.Filters {
		switch eventType {
		case models.BlockCommit:
		case models.EntryRegistration:
		case models.ChainRegistration:
		case models.ProcessMessage:
		case models.NodeMessage:
		case models.EntryContentRegistration:
		default:
			return fmt.Errorf("invalid event type: %s", eventType)
		}
	}

	switch subscription.SubscriptionStatus {
	case models.Active:
	case models.Suspended:
	default:
		return fmt.Errorf("unknown subscription status: should be one of [%s, %s]", models.Active, models.Suspended)
	}

	return nil
}
