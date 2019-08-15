package models

// type of callback: [HTTP, BEARER_TOKEN, BASIC_AUTH]
// swagger:model CallbackType
type CallbackType string

const (
	// emit event with regular http call
	// swagger:model HTTP
	HTTP CallbackType = "HTTP"

	// emit event over http with OAUTH2 authorization
	// swagger:model OAUTH2
	// OAUTH2 CallbackType = "OAUTH2"

	// emit event over http call including a bearer token
	// swagger:model BEARER TOKEN
	BEARER_TOKEN CallbackType = "BEARER_TOKEN"

	// emit event over http call with basic authentication
	// swagger:model BEARER TOKEN
	BASIC_AUTH CallbackType = "BASIC_AUTH"
)
