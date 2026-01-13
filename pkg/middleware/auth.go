// Package middleware provides HTTP middleware functions.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-api-template/internal/auth/handlers"
	"go-api-template/internal/auth/services"
	"go-api-template/pkg/response"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, map[string]string{"authorization": "Missing authorization header"})
				return
			}

			// Check Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Unauthorized(w, map[string]string{"authorization": "Invalid authorization header format"})
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				response.Unauthorized(w, map[string]string{"authorization": "Missing token"})
				return
			}

			// Validate token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				switch err {
				case services.ErrExpiredToken:
					response.Unauthorized(w, map[string]string{"token": "Token has expired"})
				case services.ErrInvalidTokenType:
					response.Unauthorized(w, map[string]string{"token": "Invalid token type"})
				default:
					response.Unauthorized(w, map[string]string{"token": "Invalid token"})
				}
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), handlers.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, handlers.UserEmailKey, claims.Email)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth wraps a handler function with auth middleware
func RequireAuth(jwtService *services.JWTService, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, map[string]string{"authorization": "Missing authorization header"})
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(w, map[string]string{"authorization": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			response.Unauthorized(w, map[string]string{"authorization": "Missing token"})
			return
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			switch err {
			case services.ErrExpiredToken:
				response.Unauthorized(w, map[string]string{"token": "Token has expired"})
			case services.ErrInvalidTokenType:
				response.Unauthorized(w, map[string]string{"token": "Invalid token type"})
			default:
				response.Unauthorized(w, map[string]string{"token": "Invalid token"})
			}
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), handlers.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, handlers.UserEmailKey, claims.Email)

		// Call handler with updated context
		handler(w, r.WithContext(ctx))
	}
}
