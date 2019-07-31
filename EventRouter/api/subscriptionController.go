package api

import (
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

	if len(subscription.Callback) > 0 {
		log.Error("invalid request: %v", request.Body)
		responseError(writer, errors.NewInvalidRequest())
	}

	subscription = repository.StoreSubscription(subscription)
	respond(writer, subscription)
}
