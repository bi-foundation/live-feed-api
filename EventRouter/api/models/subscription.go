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
