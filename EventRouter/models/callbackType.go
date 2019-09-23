package models

// CallbackType the type of callback url is used in the subscription
type CallbackType string

// Different callback types
const (
	HTTP        CallbackType = "HTTP"
	BearerToken CallbackType = "BEARER_TOKEN"
	BasicAuth   CallbackType = "BASIC_AUTH"
)
