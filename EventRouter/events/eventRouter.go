package events

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
	"net/http"
	"time"
)

const defaultMaxFailures = 3
const defaultRetryTimeout = 10 * time.Second

var maxFailures = defaultMaxFailures
var retryTimeout = defaultRetryTimeout

// EventRouter that route the events to subscriptions
type EventRouter interface {
	Start()
}

type eventRouter struct {
	eventsInQueue chan *eventmessages.FactomEvent
	emitQueue     map[string]SubscriptionStack
}

// NewEventRouter create a new event router that listens to a given queue
func NewEventRouter(queue chan *eventmessages.FactomEvent) EventRouter {
	return &eventRouter{
		eventsInQueue: queue,
		emitQueue:     make(map[string]SubscriptionStack),
	}
}

// Start the event router
func (eventRouter *eventRouter) Start() {
	go eventRouter.handleEvents()
}

func (eventRouter *eventRouter) handleEvents() {
	for factomEvent := range eventRouter.eventsInQueue {
		eventType, err := mapEventType(factomEvent)
		if err != nil {
			log.Error("invalid event type %v: '%v'", err, factomEvent.Value)
			continue
		}

		log.Info("handle %s event: %v", eventType, factomEvent)

		subscriptionContexts, err := repository.SubscriptionRepository.GetActiveSubscriptions(eventType)
		if err != nil {
			log.Error("%v", err)
			continue
		}

		err = eventRouter.send(subscriptionContexts, factomEvent)
		if err != nil {
			log.Error("%v", err)
			continue
		}
	}
}

func mapEventType(factomEvent *eventmessages.FactomEvent) (models.EventType, error) {
	switch factomEvent.Value.(type) {
	case *eventmessages.FactomEvent_DirectoryBlockCommit:
		return models.DirectoryBlockCommit, nil
	case *eventmessages.FactomEvent_ChainCommit:
		return models.ChainCommit, nil
	case *eventmessages.FactomEvent_EntryCommit:
		return models.EntryCommit, nil
	case *eventmessages.FactomEvent_EntryReveal:
		return models.EntryReveal, nil
	case *eventmessages.FactomEvent_ProcessMessage:
		return models.ProcessMessage, nil
	case *eventmessages.FactomEvent_NodeMessage:
		return models.NodeMessage, nil
	default:
		return "", fmt.Errorf("failed to map factom event to event type")
	}
}

func (eventRouter *eventRouter) send(subscriptions []*models.SubscriptionContext, factomEvent *eventmessages.FactomEvent) error {
	event, err := json.Marshal(factomEvent)
	if err != nil {
		return fmt.Errorf("failed to create json from factom event")
	}
	for _, subscription := range subscriptions {
		eventRouter.sendEvent(subscription, event)
	}
	return nil
}

// start a thread if the queue is empty and no thread is already sending events for the subscription
func (eventRouter *eventRouter) sendEvent(subscriptionContext *models.SubscriptionContext, event []byte) {
	if _, ok := eventRouter.emitQueue[subscriptionContext.Subscription.ID]; !ok {
		eventRouter.emitQueue[subscriptionContext.Subscription.ID] = NewSubscriptionStack(subscriptionContext)
	}
	eventRouter.emitQueue[subscriptionContext.Subscription.ID].Add(event)

	// start new thread to handle the process list if there isn't already a thread busy sending to the subscription
	if !eventRouter.emitQueue[subscriptionContext.Subscription.ID].IsProcessing() {
		go func() {
			eventRouter.emitEvent(subscriptionContext.Subscription.ID)
		}()
	}
}

func (eventRouter *eventRouter) emitEvent(subscriptionsId string) {
	// process all events that should be send to the subscription
	eventRouter.emitQueue[subscriptionsId].Processing(true)
	for emittingEvents := true; emittingEvents; {
		subscriptionContext, event := eventRouter.emitQueue[subscriptionsId].Pop()
		// check if there is nothing left to process
		if event == nil || subscriptionContext.Subscription.SubscriptionStatus != models.Active {
			emittingEvents = false
			continue
		}

		// update the subscription if there was a failure in the mean time
		if subscriptionContext.Failures > 0 {
			// is subscription context ready updated?
			var err error
			subscriptionContext, err = repository.SubscriptionRepository.ReadSubscription(subscriptionsId)
			if err != nil {
				handleSendFailure(subscriptionContext, err.Error())

				// put the event back on the stack and wait to resend event
				eventRouter.emitQueue[subscriptionsId].Push(event)
				time.Sleep(retryTimeout)
				continue
			}
			eventRouter.emitQueue[subscriptionsId].UpdateSubscription(subscriptionContext)
		}

		err := executeSend(&subscriptionContext.Subscription, event)

		// if there was a failure, update the context in case the subscription has been updated in the mean time
		if err != nil {
			handleSendFailure(subscriptionContext, err.Error())

			// put the event back on the stack and wait to resend event
			eventRouter.emitQueue[subscriptionsId].Push(event)
			time.Sleep(retryTimeout)
			continue
		}

		handleSendSuccessful(subscriptionContext)
	}
	eventRouter.emitQueue[subscriptionsId].Processing(false)
}

func executeSend(subscription *models.Subscription, event []byte) error {
	url := subscription.CallbackURL

	// Create a new request
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(event))
	if err != nil || request == nil {
		return fmt.Errorf("failed to create request to '%s': %v", url, err)
	}

	// setup authentication
	if subscription.CallbackType == models.BasicAuth {
		auth := subscription.Credentials.BasicAuthUsername + ":" + subscription.Credentials.BasicAuthPassword
		authentication := base64.StdEncoding.EncodeToString([]byte(auth))
		request.Header.Add("Authorization", "Basic "+authentication)
	} else if subscription.CallbackType == models.BearerToken {
		bearer := "Bearer " + subscription.Credentials.AccessToken
		request.Header.Add("Authorization", bearer)
	}

	log.Debug("send event to '%s' %v", subscription.CallbackURL, subscription.CallbackType)

	// send request using default http Client
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return fmt.Errorf("failed to send event to '%s': %v", url, err)
	}
	if response == nil {
		return fmt.Errorf("failed to receive correct response from '%s': no response", url)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to receive correct response from '%s': code=%d, body=%v", url, response.StatusCode, response)
	}

	return nil
}

// emit event fails, if the number of failures pass a threshold, suspend the subscription
// set the reason in the subscription info
func handleSendFailure(subscriptionContext *models.SubscriptionContext, reason string) {
	subscriptionContext.Failures++
	subscriptionContext.Subscription.SubscriptionInfo = fmt.Sprintf("%s%d: %s\n", subscriptionContext.Subscription.SubscriptionInfo, subscriptionContext.Failures, reason)
	if subscriptionContext.Failures >= maxFailures {
		subscriptionContext.Subscription.SubscriptionStatus = models.Suspended
	}
	// update the database
	_, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
	if err != nil {
		log.Error("failed update subscription after delivery failure: %v", err)
	}
}

func handleSendSuccessful(subscriptionContext *models.SubscriptionContext) {
	// update only the subscription if the failures and status needs to be reset
	if subscriptionContext.Failures > 0 {
		subscriptionContext.Failures = 0
		subscriptionContext.Subscription.SubscriptionStatus = models.Active
		subscriptionContext.Subscription.SubscriptionInfo = ""

		// update the database
		_, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
		if err != nil {
			log.Error("failed update subscription after delivery failure: %v", err)
		}
	}
}
