package api

import (
	"github.com/FactomProject/live-api/EventRouter/api/errors"
	"github.com/FactomProject/live-api/EventRouter/api/models"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"net/http"
)

func subscribe(writer http.ResponseWriter, request *http.Request) {
	subscription := &models.Subscription{}
	if decode(writer, request, subscription) {
		return
	}

	if subscription == nil {
		log.Error("invalid request: %v", request.Body)
		responseError(writer, errors.NewInvalidRequest())
	}

	subscription = repository.StoreSubscription(subscription)
	respond(writer, subscription)
}
