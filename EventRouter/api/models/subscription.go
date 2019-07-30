package models

// Subscription
//
// summary:
//   An application is able to have a subscription for events
//   Events will be send to the callback endpoint of the subscription
// swagger:response Subscription
type Subscription struct {
	// the id of the subscription
	//
	// read only: true
	Id string `json:"id"`

	// the endpoint to receive the callback
	// required: true
	Callback string `json:"callback"`
}
