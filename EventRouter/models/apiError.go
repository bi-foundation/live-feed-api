package models

// APIError for api errors.
type APIError struct {

	// Error code.
	Code int `json:"code"`

	// Error message.
	Message string `json:"message"`

	// Error details.
	Details string `json:"details"`
}
