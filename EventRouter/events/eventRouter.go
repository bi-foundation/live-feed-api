package events

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
)

const defaultMaxFailures = 3

var maxFailures = defaultMaxFailures

// EventRouter that route the events to subscriptions
type EventRouter interface {
	Start()
}

type eventRouter struct {
	eventsInQueue chan *eventmessages.FactomEvent
	senders       map[string]*eventSender
	senderConfig  *config.SenderConfig
}

// NewEventRouter create a new event router that listens to a given queue
func NewEventRouter(senderConfig *config.SenderConfig, queue chan *eventmessages.FactomEvent) EventRouter {
	return &eventRouter{
		eventsInQueue: queue,
		senderConfig:  senderConfig,
	}
}

// Start the event router
func (evr *eventRouter) Start() {
	go evr.handleEvents()
}

func (evr *eventRouter) handleEvents() {
	for factomEvent := range evr.eventsInQueue {
		log.Debug("handle event: %v", factomEvent)

		eventType, err := mapEventType(factomEvent)
		if err != nil {
			log.Error("invalid event type %v: '%v'", err, factomEvent.Value)
			continue
		}

		log.Info("received %s event: %v", eventType, factomEvent)

		subscriptionContexts, err := repository.SubscriptionRepository.GetActiveSubscriptions(eventType)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		evr.cleanupSuspendedSubscriptions(subscriptionContexts);
		err = evr.send(subscriptionContexts, factomEvent)
		if err != nil {
			log.Error("%v", err)
			continue
		}
	}
}

func mapEventType(factomEvent *eventmessages.FactomEvent) (models.EventType, error) {
	switch factomEvent.Value.(type) {
	case *eventmessages.FactomEvent_BlockCommit:
		return models.BlockCommit, nil
	case *eventmessages.FactomEvent_ChainRegistration:
		return models.ChainRegistration, nil
	case *eventmessages.FactomEvent_EntryRegistration:
		return models.EntryRegistration, nil
	case *eventmessages.FactomEvent_EntryContentRegistration:
		return models.EntryContentRegistration, nil
	case *eventmessages.FactomEvent_ProcessMessage:
		return models.ProcessMessage, nil
	case *eventmessages.FactomEvent_NodeMessage:
		return models.NodeMessage, nil
	default:
		return "", fmt.Errorf("failed to map correct node")
	}
}

func (evr *eventRouter) send(subscriptions []*models.SubscriptionContext, factomEvent *eventmessages.FactomEvent) error {

	event, err := json.Marshal(factomEvent)
	if err != nil {
		return fmt.Errorf("failed to create json from factom event")
	}
	for _, subscription := range subscriptions {
		var sender *eventSender
		if sender, ok := evr.senders[subscription.Subscription.ID]; !ok {
			sender = NewEventSender(evr.senderConfig, subscription)
			evr.senders[subscription.Subscription.ID] = sender
		}
		sender.QueueEvent(&event)
	}
	return nil
}

func (evr *eventRouter) cleanupSuspendedSubscriptions(subscriptions []*models.SubscriptionContext) {
	for id, _ := range evr.senders {
		found := false
		for _, activeSubscriptionContext := range subscriptions {
			if activeSubscriptionContext.Subscription.ID == id {
				found = true
				break
			}
		}
		if !found {
			delete(evr.senders, id)
		}
	}
}
