package models

// swagger:enum CallbackType
type CallbackType string

const (
	// emit event with regular http call
	// swagger:model HTTP
	HTTP CallbackType = "HTTP"

	// emit event with http call including OAUTH2 token
	// swagger:model OAUTH2
	OAUTH2 CallbackType = "OAUTH2"
)
