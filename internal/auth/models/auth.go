package models

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword123"`
	Name     string `json:"name" example:"John Doe"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"900"`
}

// AuthUser represents authenticated user data (without sensitive info)
type AuthUser struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Claims represents JWT claims for authentication
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Type   string    `json:"type"` // "access" or "refresh"
	Exp    int64     `json:"exp"`
	Iat    int64     `json:"iat"`
}

// AuthResponse represents a successful authentication response (JSend format)
type AuthResponse struct {
	Status string        `json:"status" example:"success"`
	Data   *AuthRespData `json:"data"`
}

// AuthRespData contains user and token data
type AuthRespData struct {
	User   AuthUser  `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

// ProfileResponse represents a successful profile response (JSend format)
type ProfileResponse struct {
	Status string   `json:"status" example:"success"`
	Data   AuthUser `json:"data"`
}

// MessageResponse represents a simple message response (JSend format)
type MessageResponse struct {
	Status string            `json:"status" example:"success"`
	Data   map[string]string `json:"data"`
}
