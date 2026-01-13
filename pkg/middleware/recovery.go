// Package middleware provides HTTP middleware functions for the API.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"go-api-template/pkg/response"
)

// Recovery returns a middleware that recovers from panics.
// It logs the panic with stack trace and returns a 500 error response.
// This prevents the server from crashing on unhandled panics.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := debug.Stack()

					// Get request ID if available
					requestID := GetRequestID(r.Context())

					// Log the panic with full context
					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("request_id", requestID),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("remote_addr", r.RemoteAddr),
						slog.String("stack", string(stack)),
					)

					// Return 500 error to client
					// Don't expose internal error details to client
					response.InternalError(w, "An unexpected error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
