package models

type CallbackType string

const (
	HTTP         CallbackType = "HTTP"
	BEARER_TOKEN CallbackType = "BEARER_TOKEN"
	BASIC_AUTH   CallbackType = "BASIC_AUTH"
)
