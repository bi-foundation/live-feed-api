package errors

// ApiError
// An error occurred. This can be an invalid input of other unexpected error occurred.
// swagger:model ApiError
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// ApiErrorResponse
//
// swagger:response ApiError
type apiErrorResponse struct {
	// API Error
	//
	// in: body
	apiError *ApiError `json:"error"`
}

func NewInternalError() *ApiError {
	return &ApiError{-410800, "Internal error", ""}
}

func NewMethodNotFoundError() *ApiError {
	return &ApiError{-410801, "Method not found", ""}
}

func NewInvalidRequest() *ApiError {
	return NewInvalidRequestDetailed("")
}

func NewInvalidRequestDetailed(reason string) *ApiError {
	return &ApiError{-410810, "invalid request", reason}
}

func NewParseError() *ApiError {
	return &ApiError{-32700, "Parse error", ""}
}
