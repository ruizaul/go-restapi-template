// Package authx provides JWT authentication utilities and password hashing
package authx

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken indicates the token is malformed or invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken indicates the token has expired
	ErrExpiredToken = errors.New("token expired")
	// ErrInvalidTokenType indicates the token type doesn't match expected type
	ErrInvalidTokenType = errors.New("invalid token type")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken represents a short-lived access token
	AccessToken TokenType = "access"
	// RefreshToken represents a long-lived refresh token
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT claims with user information
type Claims struct {
	jwt.RegisteredClaims
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	UserID uuid.UUID `json:"user_id"`
}

// GenerateAccessToken creates a new JWT access token
func GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	expiryStr := os.Getenv("JWT_ACCESS_EXPIRY")
	if expiryStr == "" {
		expiryStr = "15m"
	}

	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 15 * time.Minute
	}

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a new JWT refresh token
func GenerateRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	expiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if expiryStr == "" {
		expiryStr = "168h" // 7 days
	}

	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 168 * time.Hour
	}

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// HashRefreshToken creates a SHA-256 hash of a refresh token for secure storage
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
