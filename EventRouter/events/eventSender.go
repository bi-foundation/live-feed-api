package events

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/repository"
	"net/http"
	"time"
)

const (
	defaultOutputChannelSize = 65535
	retrySleepDuration       = 10 * time.Second
)

type eventBufferContainer struct {
	eventBuffer []byte
}

type eventSender struct {
	senderConfig         *config.SenderConfig
	subscriptionContext  *models.SubscriptionContext
	outputQueue          chan *eventBufferContainer
	postponeSendingUntil time.Time
}

func NewEventSender(senderConfig *config.SenderConfig, subscriptionContext *models.SubscriptionContext) *eventSender {
	sender := &eventSender{
		senderConfig:        senderConfig,
		subscriptionContext: subscriptionContext,
		outputQueue:         make(chan *eventBufferContainer, defaultOutputChannelSize),
	}
	go sender.processQueueLoop()
	return sender
}

func (sender eventSender) QueueEvent(eventBufferContainer *eventBufferContainer) {
	select {
	case sender.outputQueue < eventBufferContainer:
	default:
		// TODO counter?
	}

}

func (sender eventSender) processQueueLoop() {
	for event := range sender.outputQueue {
		if sender.postponeSendingUntil.IsZero() || sender.postponeSendingUntil.Before(time.Now()) {
			sender.sendEvent(event)
		}
	}
}

func (sender eventSender) sendEvent(event *[]byte) {
	subscriptionContext := sender.subscriptionContext
	subscription := subscriptionContext.Subscription
	url := subscription.CallbackURL
	sendSuccessful := false
	var lastError *string
	for retry := uint16(0); retry < sender.senderConfig.MaxEventRetries && !sendSuccessful; retry++ {

		// Create a new request
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(*event))
		if err != nil || request == nil {
			reason := fmt.Sprintf("failed to create request to '%s': %v", url, err)
			log.Error(reason)
			lastError = &reason
			time.Sleep(retrySleepDuration)
			continue
		}
		setAuthentication(subscription, request)

		// send request using default http Client
		log.Debug("sending event to '%s' %v", subscription.CallbackURL, subscription.CallbackType)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			reason := fmt.Sprintf("failed to send event to '%s': %v", url, err)
			log.Error(reason)
			lastError = &reason
			time.Sleep(retrySleepDuration)
			continue
		}
		if response == nil {
			reason := fmt.Sprintf("failed to receive correct response from '%s': no response", url)
			log.Error(reason)
			lastError = &reason
			time.Sleep(retrySleepDuration)
			continue
		}
		if response.StatusCode != http.StatusOK {
			reason := fmt.Sprintf("failed to receive correct response from '%s': code=%d, body=%v", url, response.StatusCode, response)
			log.Error(reason)
			lastError = &reason
			time.Sleep(retrySleepDuration)
			continue
		}
	}
	if lastError != nil {
		sender.sendEventFailure(subscriptionContext, *lastError)
	} else {
		sendSuccessful = true
		sender.sendEventSuccessful(subscriptionContext)
	}
}

func setAuthentication(subscription models.Subscription, request *http.Request) {
	// build authentication header param
	if subscription.CallbackType == models.BasicAuth {
		auth := subscription.Credentials.BasicAuthUsername + ":" + subscription.Credentials.BasicAuthPassword
		authentication := base64.StdEncoding.EncodeToString([]byte(auth))
		request.Header.Add("Authorization", "Basic "+authentication)
	} else if subscription.CallbackType == models.BearerToken {
		bearer := "Bearer " + subscription.Credentials.AccessToken
		request.Header.Add("Authorization", bearer)
	}
}

// emit event fails, if the number of failures pass a threshold, suspend the subscription
// set the reason in the subscription info
func (sender eventSender) sendEventFailure(subscriptionContext *models.SubscriptionContext, reason string) {
	subscriptionContext.Failures++
	if subscriptionContext.Failures > sender.senderConfig.MaxReconnectRetries {
		subscriptionContext.Subscription.SubscriptionStatus = models.Suspended
		subscriptionContext.Subscription.SubscriptionInfo = reason
	} else {
		sender.postponeSendingUntil = time.Now().Add(sender.senderConfig.ReconnectHoldOffDuration)
	}

	// update the database
	_, err := repository.SubscriptionRepository.UpdateSubscription(subscriptionContext)
	if err != nil {
		log.Error("failed update subscription after delivery failure: %v", err)
	}
}

func (sender eventSender) sendEventSuccessful(subscriptionContext *models.SubscriptionContext) {
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
