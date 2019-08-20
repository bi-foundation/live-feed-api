package events

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/gen/eventmessages"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"net/http"
)

const MAX_FAILURES_DEFAULT = 3

var maxFailures = MAX_FAILURES_DEFAULT

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

		log.Info("received %s with event source %v: %v", eventType, factomEvent.GetEventSource(), factomEvent.GetAnchorEvent())

		subscriptionContexts, err := repository.SubscriptionRepository.GetActiveSubscriptions(eventType)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		err = send(subscriptionContexts, factomEvent)
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
	case *eventmessages.FactomEvent_ProcessEvent:
		return models.PROCESS_MESSAGE, nil
	case *eventmessages.FactomEvent_NodeEvent:
		return models.NODE_MESSAGE, nil
	default:
		return "", fmt.Errorf("failed to map correct node")
	}
}

func send(subscriptions []*models.SubscriptionContext, factomEvent *eventmessages.FactomEvent) error {
	event, err := json.Marshal(factomEvent)
	if err != nil {
		return fmt.Errorf("failed to create json from factom event")
	}
	for _, subscription := range subscriptions {
		sendEvent(subscription, event)
	}
	return nil
}

func sendEvent(subscriptionContext *models.SubscriptionContext, event []byte) {
	subscription := subscriptionContext.Subscription
	url := subscription.CallbackUrl

	// Create a new request
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(event))
	if err != nil || request == nil {
		log.Error("failed to create request to '%s': %v", url, err)
		sendEventFailure(subscriptionContext, fmt.Sprintf("create request failed to '%s': %v", url, err))
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

	log.Debug("send event to '%s' %v", subscription.CallbackUrl, subscription.CallbackType)
	// send request using default http Client
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Error("failed to send event to '%s': %v", url, err)
		sendEventFailure(subscriptionContext, fmt.Sprintf("send event failed to '%s': %v", url, err))
		return
	}
	if response == nil {
		log.Error("failed to receive correct response from '%s': no response", url)
		sendEventFailure(subscriptionContext, fmt.Sprintf("incorrect response from '%s': no response", url))
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Error("failed to receive correct response from '%s': code=%d, body=%v", url, response.StatusCode, response)
		sendEventFailure(subscriptionContext, fmt.Sprintf("incorrect response from '%s': code=%d, body=%v", url, response.StatusCode, response))
		return
	}

	sendEventSuccessful(subscriptionContext)
}

// emit event fails, if the number of failures pass a threshold, suspend the subscription
// set the reason in the subscription info
func sendEventFailure(subscriptionContext *models.SubscriptionContext, reason string) {
	subscriptionContext.Failures++
	if subscriptionContext.Failures > maxFailures {
		subscriptionContext.Subscription.SubscriptionStatus = models.SUSPENDED
		subscriptionContext.Subscription.SubscriptionInfo = reason
	}
	// update the database
	_, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
	if err != nil {
		log.Error("failed update subscription after delivery failure: %v", err)
	}
}

func sendEventSuccessful(subscriptionContext *models.SubscriptionContext) {
	if subscriptionContext.Failures > 0 {
		subscriptionContext.Failures = 0
		subscriptionContext.Subscription.SubscriptionStatus = models.ACTIVE
		subscriptionContext.Subscription.SubscriptionInfo = ""

		// update the database
		_, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
		if err != nil {
			log.Error("failed update subscription after delivery failure: %v", err)
		}
	}
}
