package api

import (
	"github.com/gorilla/mux"
	"live-api/EventRouter/api/errors"
	"live-api/EventRouter/api/models"
	"live-api/EventRouter/log"
	"live-api/EventRouter/repository"
	"net/http"
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

	// TODO validate callback url
	if len(subscription.Callback) < 1 {
		log.Error("invalid subscribe request: %v", subscription)
		responseError(writer, errors.NewInvalidRequestDetailed("wrong callback url format"))
		return
	}

	subscription = repository.CreateSubscription(subscription)
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
	subscription := repository.DeleteSubscription(id)
	respond(writer, subscription)
}
