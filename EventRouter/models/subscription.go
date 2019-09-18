package models

// Subscription for subscribing a callback url to receive events
type Subscription struct {

	// The id of the subscription.
	ID string `json:"id" binding:"required"`

	// The callback endpoint to receive the events.
	CallbackURL string `json:"callbackUrl" binding:"required" example:"https://server.com/events"`

	// Type of callback.
	// - HTTP to deliver the events to a http/https endpoint.
	// - BEARER_TOKEN to deliver the events to a http/https endpoint with a bearer token for authentication.
	// - BASIC_AUTH to deliver the events to a http/https endpoint with a basic authentication.
	CallbackType CallbackType `json:"callbackType" binding:"required" example:"HTTP" enums:"HTTP,BEARER_TOKEN,BASIC_AUTH"`

	// Status of subscription. Normally a subscription is active. When events fail to be delivered the subscription will be suspended. The subscription can become active again by updating the subscription. When the subscription is suspended, the error information is set in the info field.
	SubscriptionStatus SubscriptionStatus `json:"status" example:"ACTIVE" enums:"ACTIVE,SUSPENDED"`

	// Information of the subscription. An information message can be for example about why the subscription is suspended.
	SubscriptionInfo string `json:"info"`

	// The emitted event can be filter to receive not all data from an event type. Subscribe on one or more event types. For every event type a filtering can be defined.
	Filters map[EventType]Filter `json:"filters"`

	// Credentials of the callback endpoint where events are delivered.
	Credentials Credentials `json:"credentials"`
}
