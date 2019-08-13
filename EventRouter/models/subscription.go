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

	// the emitted event can be filter to receive not all data from an event type
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
// This is used for operations that want an Order as body of the request
// swagger:parameters SubscriptionRequest
type subscriptionRequest struct {
	// The subscription registration for receiving information from factomd through the live api.
	//
	// in: body
	Subscription *Subscription `json:"subscription"`
}

// An SubscriptionResponse is the stored subscription for factom events
//
// swagger:response SubscriptionResponse
type subscriptionResponse struct {
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
