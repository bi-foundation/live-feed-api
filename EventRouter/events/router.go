package events

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"net/http"
)

type EventRouter struct {
	eventsInQueue chan *eventmessages.FactomEvent
}

func NewEventRouter(queue chan *eventmessages.FactomEvent) EventRouter {
	return EventRouter{eventsInQueue: queue}
}

func (evr *EventRouter) Start() {
	go evr.handleEvents()
}

func (evr *EventRouter) handleEvents() {
	for factomEvent := range evr.eventsInQueue {
		log.Debug("handle event: %v", factomEvent)

		eventType, err := mapEventType(factomEvent)
		if err != nil {
			log.Error("invalid event type %v: '%v'", err, factomEvent.Value)
			continue
		}

		log.Info("Received %s with event source %v: %v", eventType, factomEvent.GetEventSource(), factomEvent.GetAnchorEvent())

		subscriptions, err := repository.SubscriptionRepository.GetSubscriptions(eventType)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		err = send(subscriptions, factomEvent)
		if err != nil {
			log.Error("%v", err)
			continue
		}
	}
}

func mapEventType(factomEvent *eventmessages.FactomEvent) (models.EventType, error) {
	switch factomEvent.Value.(type) {
	case *eventmessages.FactomEvent_AnchorEvent:
		return models.ANCHOR_EVENT, nil
	case *eventmessages.FactomEvent_CommitChain:
		return models.COMMIT_CHAIN, nil
	case *eventmessages.FactomEvent_CommitEntry:
		return models.COMMIT_ENTRY, nil
	case *eventmessages.FactomEvent_RevealEntry:
		return models.REVEAL_ENTRY, nil
	case *eventmessages.FactomEvent_NodeMessage:
		return models.NODE_MESSAGE, nil
	default:
		return "", fmt.Errorf("failed to map correct node")
	}
}

func send(subscriptions []*models.Subscription, factomEvent *eventmessages.FactomEvent) error {
	event, err := json.Marshal(factomEvent)
	if err != nil {
		return fmt.Errorf("failed to create json from factom event")
	}
	for _, subscription := range subscriptions {
		sendEvent(subscription, event)
	}
	return nil
}

func sendEvent(subscription *models.Subscription, event []byte) {
	url := subscription.Callback
	// Create a new request
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(event))
	if err != nil || request == nil {
		log.Error("failed to create request to '%s': %v", url, err)
		return
	}

	// setup authentication
	if subscription.CallbackType == models.BASIC_AUTH {
		auth := subscription.Credentials.BasicAuthUsername + ":" + subscription.Credentials.BasicAuthPassword
		authentication := base64.StdEncoding.EncodeToString([]byte(auth))
		request.Header.Add("Authorization", "Basic "+authentication)
	} else if subscription.CallbackType == models.BEARER_TOKEN {
		bearer := "Bearer " + subscription.Credentials.AccessToken
		request.Header.Add("Authorization", bearer)
	}

	log.Debug("send event to '%s' %v", subscription.Callback, subscription.CallbackType)
	// Send request using default http Client
	response, err := http.DefaultClient.Do(request)

	// TODO handle endpoint failure
	if err != nil {
		log.Error("failed to send event to '%s': %v", url, err)
		return
	}
	if response == nil {
		log.Error("failed to receive correct response from '%s': no response", url)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Error("failed to receive correct response from '%s': code=%d, body=%v", url, response.StatusCode, response)
		return
	}
}
