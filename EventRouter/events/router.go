package events

import (
	"bytes"
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

		// TODO what about types?
		subscriptions, err := repository.SubscriptionRepository.GetSubscriptions(models.COMMIT_EVENT)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		err = send(subscriptions, factomEvent)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		switch factomEvent.Value.(type) {
		case *eventmessages.FactomEvent_AnchorEvent:
			log.Info("Received AnchoredEvent with event source %v: %v", factomEvent.GetEventSource(), factomEvent.GetAnchorEvent())
		case *eventmessages.FactomEvent_CommitChain:
			log.Info("Received CommitChain with event source %v: %v", factomEvent.GetEventSource(), factomEvent.GetCommitChain())
		case *eventmessages.FactomEvent_CommitEntry:
			log.Info("Received CommitEntry with event source %v: %v", factomEvent.GetEventSource(), factomEvent.GetCommitEntry())
		case *eventmessages.FactomEvent_RevealEntry:
			log.Info("Received FactomEvent_RevealEntry with event source %v: %v", factomEvent.GetEventSource(), factomEvent.GetRevealEntry())
		case *eventmessages.FactomEvent_NodeMessage:
			log.Info("Received FactomEvent_NodeMessage with event source %v: %v", factomEvent.GetEventSource(), factomEvent.GetNodeMessage())
		}
	}
}

func send(subscriptions []*models.Subscription, factomEvent *eventmessages.FactomEvent) error {
	event, err := json.Marshal(factomEvent)
	if err != nil {
		return fmt.Errorf("failed to create json from factom event")
	}
	for _, subscription := range subscriptions {
		url := subscription.Callback
		sendEvent(url, event)
	}
	return nil
}

func sendEvent(url string, event []byte) {
	response, err := http.Post(url, "application/json", bytes.NewBuffer(event))
	// TODO handle endpoint failure
	if err != nil {
		log.Error("failed to send event to '%s': %v", url, err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		log.Error("failed to receive correct response from '%s': %v", url, response)
	}
}
