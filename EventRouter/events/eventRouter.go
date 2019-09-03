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
	"github.com/graphql-go/graphql"
	"net/http"
)

const MAX_FAILURES_DEFAULT = 3

var maxFailures = MAX_FAILURES_DEFAULT

type EventRouter struct {
	eventsInQueue chan *eventmessages.FactomEvent
	graphQlSchema graphql.Schema
}

func NewEventRouter(queue chan *eventmessages.FactomEvent) EventRouter {
	return EventRouter{eventsInQueue: queue}
}

func (evr *EventRouter) Start() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{})
	if err != nil {
		panic(fmt.Sprintf("could initialize graphql: %v", err))
	}
	evr.graphQlSchema = schema

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

		log.Info("received %s with event source %v", eventType, factomEvent /*.GetEventSource(), factomEvent.GetAnchorEvent()*/)

		subscriptionContexts, err := repository.SubscriptionRepository.GetActiveSubscriptions(eventType)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		var event *[]byte
		for _, subscription := range subscriptionContexts {
			if !evr.filterPass(subscription, factomEvent) {
				continue
			}

			if event == nil {
				eventData, err := json.Marshal(factomEvent)
				if err != nil {
					log.Error("failed to create json from factom event: %v", err)
					continue
				}
				event = &eventData
			}
			sendEvent(subscription, *event)
		}
	}
}

func (evr *EventRouter) filterPass(context *models.SubscriptionContext, event *eventmessages.FactomEvent) bool {
	filters := context.Subscription.Filters
	result := true
	for _, filter := range filters {
		if len(filter.Filtering) > 0 {
			result = result && evr.evalFilter(filter, event)
		}
	}
	return result
}

func (evr *EventRouter) evalFilter(filter models.Filter, event *eventmessages.FactomEvent) bool {
	graphql.Do(graphql.Params{})
}

func mapEventType(factomEvent *eventmessages.FactomEvent) (models.EventType, error) {
	// TODO fix models, and proto
	return models.ANCHOR_EVENT, nil
	/*	switch factomEvent.Value.(type) {
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
	*/
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
