package errors

import "github.com/FactomProject/live-api/EventRouter/models"

func NewInternalError(reason string) *models.ApiError {
	return &models.ApiError{-410800, "internal error", reason}
}

func NewMethodNotFoundError() *models.ApiError {
	return &models.ApiError{-410801, "method not found", ""}
}

func NewInvalidRequest() *models.ApiError {
	return NewInvalidRequestDetailed("")
}

func NewInvalidRequestDetailed(reason string) *models.ApiError {
	return &models.ApiError{-410810, "invalid request", reason}
}

func NewParseError() *models.ApiError {
	return &models.ApiError{-410800, "parse error", ""}
}
