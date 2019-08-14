package models

//   An application is able to have a subscription for events
//   Events will be send to the callback endpoint of the subscription
// swagger:model Subscription
type Subscription struct {
	// the id of the subscription
	//
	// read only: true
	Id string `json:"id"`

	// the endpoint to receive the callback
	// required: true
	Callback string `json:"callback"`

	// the type of callback: HTTP(S), HTTP with basic authentication, HTTP with OAUTH2 token
	// swagger:enum CallbackType
	// required: true
	CallbackType CallbackType `json:"callbackType"`

	// Status of subscription. Normally a subscription is active. When events fail to be delivered the subscription
	// will be suspended. The subscription can become active again by updating the subscription. When the subscription
	// is suspended, the error information is set in the info field
	// swagger:enum SubscriptionStatus
	// required: false
	SubscriptionStatus SubscriptionStatus `json:"status"`

	// Information of the subscription. An information message can be for example about why the subscription is suspended.
	//
	// read only: true
	SubscriptionInfo string `json:"info"`

	// the emitted event can be filter to receive not all data from an event type
	// swagger:enum EventType
	//
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
// swagger:parameters SubscriptionRequest
type subscriptionRequest struct {
	// The subscription registration for receiving information from factomd through the live api.
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// An SubscriptionResponse is the stored subscription
//
// swagger:response SubscriptionResponse
type subscriptionResponse struct {
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
// swagger:response GetSubscriptionRequest
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
// swagger:parameters UnsubscribeRequest
type unsubscribeRequest struct {
	// subscription id
	//
	// In: path
	ID string `json:"id"`
}

// A Subscription is returned that is successfully unsubscribed
//
// swagger:response UnsubscriptionResponse
type unsubscriptionResponse struct {
	// The subscription
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}
