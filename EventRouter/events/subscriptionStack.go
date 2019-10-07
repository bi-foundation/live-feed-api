package events

import (
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"sync"
)

type subscriptionStack struct {
	sync.Mutex
	subscription *models.SubscriptionContext
	events       [][]byte
	processing   bool
}

// SubscriptionStack is stack to track which subscription should be processed.
type SubscriptionStack interface {
	UpdateSubscription(subscription *models.SubscriptionContext)
	Add([]byte)
	Push([]byte)
	Pop() (*models.SubscriptionContext, []byte)
	Processing(bool)
	IsProcessing() bool
}

// NewSubscriptionStack creates a subscription stack
func NewSubscriptionStack(subscription *models.SubscriptionContext) SubscriptionStack {
	return &subscriptionStack{
		subscription: subscription,
		events:       [][]byte{},
		processing:   false,
	}
}

// add the event to the back of the list
func (q *subscriptionStack) Add(item []byte) {
	q.Lock()
	defer q.Unlock()
	q.events = append(q.events, item)
}

// add the event to the front of the list
func (q *subscriptionStack) Push(item []byte) {
	q.Lock()
	defer q.Unlock()
	q.events = append([][]byte{item}, q.events...)
}

// get and remove the first item of the list
func (q *subscriptionStack) Pop() (*models.SubscriptionContext, []byte) {
	q.Lock()
	defer q.Unlock()
	if len(q.events) == 0 {
		return q.subscription, nil
	}
	item := q.events[0]
	q.events = q.events[1:]
	return q.subscription, item
}

func (q *subscriptionStack) UpdateSubscription(subscription *models.SubscriptionContext) {
	q.subscription = subscription
}

func (q *subscriptionStack) Processing(processing bool) {
	q.Lock()
	defer q.Unlock()
	q.processing = processing
}

func (q *subscriptionStack) IsProcessing() bool {
	q.Lock()
	defer q.Unlock()
	return q.processing
}
