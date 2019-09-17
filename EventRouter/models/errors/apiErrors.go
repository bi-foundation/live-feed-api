package errors

import "github.com/FactomProject/live-feed-api/EventRouter/models"

// NewInternalError create a new internal error
func NewInternalError(reason string) *models.APIError {
	return &models.APIError{
		Code:    -410800,
		Message: "internal error",
		Details: reason,
	}
}

// NewMethodNotFoundError create a new method not found error
func NewMethodNotFoundError() *models.APIError {
	return &models.APIError{Code: -410801, Message: "method not found", Details: ""}
}

// NewInvalidRequest create a new invalid request error
func NewInvalidRequest() *models.APIError {
	return NewInvalidRequestDetailed("")
}

// NewInvalidRequestDetailed create a new invalid request error with given details
func NewInvalidRequestDetailed(reason string) *models.APIError {
	return &models.APIError{Code: -410810, Message: "invalid request", Details: reason}
}

// NewParseError create a new parse error
func NewParseError() *models.APIError {
	return &models.APIError{Code: -410800, Message: "parse error", Details: ""}
}
