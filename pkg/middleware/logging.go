// Package middleware provides HTTP middleware functions for the API.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// RequestIDKey is the context key for request ID
const RequestIDKey contextKey = "request_id"

// RequestIDHeader is the HTTP header name for request ID
const RequestIDHeader = "X-Request-ID"

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// shouldSkipLogging returns true if the path should not be logged
func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/health/live",
		"/health/ready",
		"/docs",
		"/docs/swagger.json",
		"/favicon.ico",
	}

	for _, skip := range skipPaths {
		if path == skip {
			return true
		}
	}
	return false
}

// getStatusColor returns ANSI color code based on HTTP status
func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return "\033[31m" // Red
	case status >= 400:
		return "\033[33m" // Yellow
	case status >= 300:
		return "\033[36m" // Cyan
	case status >= 200:
		return "\033[32m" // Green
	default:
		return "\033[0m" // Reset
	}
}

// Logging returns a middleware that logs HTTP requests with structured logging.
// It adds a unique request ID to each request and includes it in logs.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get or generate request ID
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()[:8] // Use short ID for cleaner logs
			}

			// Add request ID to response header
			w.Header().Set(RequestIDHeader, requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := newResponseWriter(w)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Skip logging for health checks and static files
			if shouldSkipLogging(r.URL.Path) {
				return
			}

			// Calculate duration
			duration := time.Since(start)

			// Build query string info
			queryInfo := ""
			if r.URL.RawQuery != "" {
				queryInfo = "?" + r.URL.RawQuery
			}

			// Format path with query
			fullPath := r.URL.Path + queryInfo

			// Log based on status code severity
			statusColor := getStatusColor(wrapped.statusCode)
			resetColor := "\033[0m"

			// Create log attributes
			attrs := []slog.Attr{
				slog.String("id", requestID),
				slog.String("method", r.Method),
				slog.String("path", fullPath),
				slog.Int("status", wrapped.statusCode),
				slog.String("duration", duration.Round(time.Millisecond).String()),
			}

			// Add bytes only if significant
			if wrapped.bytesWritten > 0 {
				attrs = append(attrs, slog.Int("bytes", wrapped.bytesWritten))
			}

			// Log at appropriate level based on status
			logMessage := "→ " + r.Method + " " + fullPath + " " +
				statusColor + http.StatusText(wrapped.statusCode) + resetColor

			switch {
			case wrapped.statusCode >= 500:
				logger.Error(logMessage, slog.Any("attrs", attrs))
			case wrapped.statusCode >= 400:
				logger.Warn(logMessage, slog.Any("attrs", attrs))
			default:
				// Clean, simple format for successful requests
				logger.Info(logMessage,
					slog.String("id", requestID),
					slog.Int("status", wrapped.statusCode),
					slog.String("duration", duration.Round(time.Millisecond).String()),
				)
			}
		})
	}
}

// GetRequestID retrieves the request ID from the context.
// Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
