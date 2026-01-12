// Package response provides helper functions for sending JSend-formatted HTTP responses.
// JSend is a specification for a simple, consistent JSON response format.
// See: https://github.com/omniti-labs/jsend
package response

import (
	"encoding/json"
	"net/http"
)

// JSend status constants
const (
	StatusSuccess = "success"
	StatusFail    = "fail"
	StatusError   = "error"
)

// Response represents a JSend response structure
type Response struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Success sends a JSend success response with status 200 OK.
// Use this when the request was successful and you have data to return.
//
// Example output: {"status": "success", "data": {"user": {"id": "123"}}}
func Success(w http.ResponseWriter, data any) {
	SuccessWithStatus(w, http.StatusOK, data)
}

// SuccessWithStatus sends a JSend success response with a custom HTTP status code.
// Useful for 201 Created, 202 Accepted, etc.
func SuccessWithStatus(w http.ResponseWriter, statusCode int, data any) {
	resp := Response{
		Status: StatusSuccess,
		Data:   data,
	}
	writeJSON(w, statusCode, resp)
}

// Created sends a JSend success response with status 201 Created.
// Use this when a new resource has been successfully created.
func Created(w http.ResponseWriter, data any) {
	SuccessWithStatus(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response.
// Use this for successful DELETE operations or when no data needs to be returned.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Fail sends a JSend fail response for client errors (4xx).
// Use this when the request failed due to invalid data provided by the client.
// The data parameter should contain details about what went wrong.
//
// Example output: {"status": "fail", "data": {"email": "Email is required"}}
func Fail(w http.ResponseWriter, statusCode int, data any) {
	resp := Response{
		Status: StatusFail,
		Data:   data,
	}
	writeJSON(w, statusCode, resp)
}

// BadRequest sends a JSend fail response with status 400 Bad Request.
// Use this when the request body is malformed or validation fails.
func BadRequest(w http.ResponseWriter, data any) {
	Fail(w, http.StatusBadRequest, data)
}

// NotFound sends a JSend fail response with status 404 Not Found.
// Use this when the requested resource doesn't exist.
func NotFound(w http.ResponseWriter, data any) {
	Fail(w, http.StatusNotFound, data)
}

// Unauthorized sends a JSend fail response with status 401 Unauthorized.
// Use this when authentication is required but not provided or invalid.
func Unauthorized(w http.ResponseWriter, data any) {
	Fail(w, http.StatusUnauthorized, data)
}

// Forbidden sends a JSend fail response with status 403 Forbidden.
// Use this when the user is authenticated but doesn't have permission.
func Forbidden(w http.ResponseWriter, data any) {
	Fail(w, http.StatusForbidden, data)
}

// Conflict sends a JSend fail response with status 409 Conflict.
// Use this when there's a conflict with the current state (e.g., duplicate email).
func Conflict(w http.ResponseWriter, data any) {
	Fail(w, http.StatusConflict, data)
}

// UnprocessableEntity sends a JSend fail response with status 422 Unprocessable Entity.
// Use this when the request is well-formed but contains semantic errors.
func UnprocessableEntity(w http.ResponseWriter, data any) {
	Fail(w, http.StatusUnprocessableEntity, data)
}

// Error sends a JSend error response for server errors (5xx).
// Use this when something went wrong on the server side.
// The message should be a human-readable error message.
//
// Example output: {"status": "error", "message": "Database connection failed", "code": 500}
func Error(w http.ResponseWriter, statusCode int, message string) {
	resp := Response{
		Status:  StatusError,
		Message: message,
		Code:    statusCode,
	}
	writeJSON(w, statusCode, resp)
}

// InternalError sends a JSend error response with status 500 Internal Server Error.
// Use this for unexpected server errors.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}

// ServiceUnavailable sends a JSend error response with status 503 Service Unavailable.
// Use this when a dependent service is unavailable.
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, message)
}

// ValidationError is a helper to create validation error data.
// Returns a map with field names as keys and error messages as values.
//
// Example: ValidationError("email", "Email is required")
// Output: {"email": "Email is required"}
func ValidationError(field, message string) map[string]string {
	return map[string]string{field: message}
}

// ValidationErrors is a helper to send multiple validation errors.
// Takes a map of field names to error messages.
//
// Example: ValidationErrors(map[string]string{"email": "Required", "name": "Too short"})
func ValidationErrors(errors map[string]string) map[string]string {
	return errors
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, try to send a plain error response
		http.Error(w, `{"status":"error","message":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}
