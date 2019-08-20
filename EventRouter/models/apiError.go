package models

// ApiError
//
// An error occurred. This can be an invalid input of other unexpected error occurred.
// swagger:model ApiError
type ApiError struct {

	// Error code.
	Code int `json:"code"`

	// Error message.
	Message string `json:"message"`

	// Error details
	Details string `json:"details"`
}

// An error has occurred
//
// swagger:response ApiError
type apiErrorResponse struct {
	// API Error
	//
	// in: body
	apiError *ApiError `json:"error"`
}

// Subscription not found
//
// swagger:response SubscriptionNotFoundError
type subscriptionNotFoundError struct {
	// API Error
	//
	// in: body
	apiError *ApiError `json:"error"`
}
