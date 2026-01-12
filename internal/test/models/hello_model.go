package models

// HelloData contains the hello world response data
type HelloData struct {
	// Message is the greeting message
	Message string `json:"message" example:"Hello World! This is an example endpoint"`
	// Example indicates if this is an example endpoint
	Example bool `json:"example" example:"true"`
}

// HelloResponse represents the success response for hello endpoint
// @Description Success response with hello data
type HelloResponse struct {
	// Status is always "success" for successful responses
	Status string `json:"status" example:"success"`
	// Data contains the response payload
	Data HelloData `json:"data"`
}

// ErrorData contains error details for failed requests
type ErrorData struct {
	// Message describes what went wrong
	Message string `json:"message" example:"Resource not found"`
}

// FailResponse represents a client error response (4xx)
// @Description Client error response
type FailResponse struct {
	// Status is always "fail" for client errors
	Status string `json:"status" example:"fail"`
	// Data contains error details
	Data ErrorData `json:"data"`
}

// ErrorResponse represents a server error response (5xx)
// @Description Server error response
type ErrorResponse struct {
	// Status is always "error" for server errors
	Status string `json:"status" example:"error"`
	// Message describes the server error
	Message string `json:"message" example:"Internal server error"`
	// Code is the HTTP status code
	Code int `json:"code" example:"500"`
}
