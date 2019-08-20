package models

// Subscription
//
// An application is able to have a subscription for events.
// Events will be send to the callback url endpoint of the subscription.
// swagger:model Subscription
type Subscription struct {

	// The id of the subscription.
	//
	// read only: true
	Id string `json:"id"`

	// The endpoint to receive the events.
	//
	// example: https://servser.com/events
	// required: true
	CallbackUrl string `json:"callbackUrl"`

	// Type of callback [HTTP, BEARER_TOKEN, BASIC_AUTH].
	// - HTTP to deliver the events to a http/https endpoint.
	// - BEARER_TOKEN to deliver the events to a http/https endpoint with a bearer token for authentication.
	// - BASIC_AUTH to deliver the events to a http/https endpoint with a basic authentication.
	//
	// example: HTTP
	// required: true
	CallbackType CallbackType `json:"callbackType"`

	// Status of subscription. Normally a subscription is active. When events fail to be delivered the subscription will be suspended. The subscription can become active again by updating the subscription. When the subscription is suspended, the error information is set in the info field
	//
	// example: ACTIVE
	// required: false
	SubscriptionStatus SubscriptionStatus `json:"status"`

	// Information of the subscription. An information message can be for example about why the subscription is suspended.
	//
	// example:
	// read only: true
	SubscriptionInfo string `json:"info"`

	// The emitted event can be filter to receive not all data from an event type. Subscribe on one or more event types. For every event type a filtering can be defined.
	//
	// example: { "COMMIT_CHAIN": { "filtering": "string" }, "NODE_MESSAGE": { "filtering": "string" } }
	Filters map[EventType]Filter `json:"filters"`

	// the emitted event can be filter to receive not all data from an event type
	//
	Credentials Credentials `json:"credentials"`
}

// SubscriptionRequest
// summary:
// An SubscriptionRequest model.
//
// This is used to subscribe for factom events
// swagger:parameters CreateSubscriptionRequest
type createSubscriptionRequest struct {
	// The subscription registration for receiving information from factomd through the live api.
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// An SubscriptionResponse is the stored subscription
//
// swagger:response CreateSubscriptionResponse
type createSubscriptionResponse struct {
	// The subscription
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// Get a subscription request
// swagger:parameters GetSubscriptionRequest
type getSubscriptionRequest struct {
	// subscription id
	//
	// In: path
	ID string `json:"id"`
}

// An SubscriptionResponse is the stored subscription
//
// swagger:response GetSubscriptionResponse
type getSubscriptionResponse struct {
	// The subscription
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// UpdateSubscriptionRequest
// summary:
// An UpdateSubscriptionRequest model.
//
// This is used to update a subscription as body of the request
// swagger:parameters UpdateSubscriptionRequest
type updateSubscriptionRequest struct {
	// subscription id
	//
	// In: path
	ID string `json:"id"`

	// The subscription registration for receiving information from factomd through the live api.
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// An SubscriptionResponse is the stored subscription
//
// swagger:response UpdateSubscriptionResponse
type updateSubscriptionResponse struct {
	// The subscription
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// unsubscription request
// swagger:parameters DeleteSubscriptionRequest
type deleteSubscriptionRequest struct {
	// subscription id
	//
	// In: path
	ID string `json:"id"`
}

// A Subscription is returned that is successfully unsubscribed
//
// swagger:response DeleteSubscriptionResponse
type deleteSubscriptionResponse struct {
	// The subscription
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}
