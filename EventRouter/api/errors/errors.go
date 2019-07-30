package errors

// ApiError
// description: An ApiEror is an error that is used when the required input fails validation.
// swagger:response ApiError
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewInternalError() *ApiError {
	return &ApiError{-410800, "Internal error"}
}

func NewMethodNotFoundError() *ApiError {
	return &ApiError{-410801, "Method not found"}
}

func NewInvalidRequest() *ApiError {
	return &ApiError{-410810, "invalid request"}
}

func NewParseError() *ApiError {
	return &ApiError{-32700, "Parse error"}
}
