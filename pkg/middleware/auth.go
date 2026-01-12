// Package middleware provides HTTP middleware for authentication, logging, and CORS
package middleware

import (
	"context"
	"net/http"
	"strings"

	"tacoshare-delivery-api/pkg/authx"
	"tacoshare-delivery-api/pkg/httpx"
)

type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey contextKey = "user_email"
	// UserRoleKey is the context key for user role
	UserRoleKey contextKey = "user_role"

	// Role constants
	RoleDriver = "driver"
)

// RequireAuth validates JWT token and adds user info to context
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Try Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract token: support both "Bearer <token>" and raw token
			parts := strings.Split(authHeader, " ")
			switch {
			case len(parts) == 2 && parts[0] == "Bearer":
				// Standard format: "Bearer <token>"
				token = parts[1]
			case len(parts) == 1 && !strings.Contains(authHeader, " "):
				// Raw token without "Bearer" prefix (for convenience in Scalar)
				token = authHeader
			default:
				httpx.RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
		} else {
			// Check for token in query parameters (for WebSocket connections)
			token = r.URL.Query().Get("token")
			if token == "" {
				httpx.RespondError(w, http.StatusUnauthorized, "Authorization header or token query parameter required")
				return
			}
		}

		claims, err := authx.ValidateToken(token, authx.AccessToken)
		if err != nil {
			if err == authx.ErrExpiredToken {
				httpx.RespondError(w, http.StatusUnauthorized, "Token expired")
				return
			}
			httpx.RespondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if user has the required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				httpx.RespondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				httpx.RespondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WebSocketAuth validates JWT token for WebSocket connections without writing HTTP responses
// This middleware is specifically designed for WebSocket upgrades where we cannot write
// regular HTTP responses before the upgrade happens
func WebSocketAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Try Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract token: support both "Bearer <token>" and raw token
			parts := strings.Split(authHeader, " ")
			switch {
			case len(parts) == 2 && parts[0] == "Bearer":
				// Standard format: "Bearer <token>"
				token = parts[1]
			case len(parts) == 1 && !strings.Contains(authHeader, " "):
				// Raw token without "Bearer" prefix
				token = authHeader
			default:
				// Return simple 401 for WebSocket
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
		} else {
			// Check for token in query parameters (for WebSocket connections)
			token = r.URL.Query().Get("token")
			if token == "" {
				// Return simple 401 for WebSocket
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}
		}

		claims, err := authx.ValidateToken(token, authx.AccessToken)
		if err != nil {
			if err == authx.ErrExpiredToken {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
