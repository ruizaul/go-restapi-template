package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-api-template/internal/auth/models"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey         []byte
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// jwtHeader represents the JWT header
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string) (*models.TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessToken, err := s.generateToken(userID, email, "access", now, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(userID, email, "refresh", now, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

// generateToken creates a JWT token
func (s *JWTService) generateToken(userID uuid.UUID, email, tokenType string, now time.Time, ttl time.Duration) (string, error) {
	header := jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	claims := models.Claims{
		UserID: userID,
		Email:  email,
		Type:   tokenType,
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64URLEncode(headerJSON)

	// Encode payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64URLEncode(payloadJSON)

	// Create signature
	signatureInput := headerEncoded + "." + payloadEncoded
	signature := s.sign([]byte(signatureInput))
	signatureEncoded := base64URLEncode(signature)

	return signatureInput + "." + signatureEncoded, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*models.Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Verify signature
	signatureInput := parts[0] + "." + parts[1]
	expectedSignature := s.sign([]byte(signatureInput))
	providedSignature, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !hmac.Equal(expectedSignature, providedSignature) {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims models.Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

// ValidateAccessToken validates an access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*models.Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*models.Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// sign creates an HMAC-SHA256 signature
func (s *JWTService) sign(data []byte) []byte {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write(data)
	return h.Sum(nil)
}

// base64URLEncode encodes data using URL-safe base64
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// base64URLDecode decodes URL-safe base64 data
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// GetAccessTokenTTL returns the access token TTL
func (s *JWTService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token TTL
func (s *JWTService) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
