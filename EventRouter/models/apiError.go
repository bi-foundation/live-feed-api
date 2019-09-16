package models

type ApiError struct {

	// Error code.
	Code int `json:"code"`

	// Error message.
	Message string `json:"message"`

	// Error details
	Details string `json:"details"`
}
