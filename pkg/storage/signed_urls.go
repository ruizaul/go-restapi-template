// Package storage provides utilities for Cloudflare R2 storage with signed URLs
package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"time"
)

const (
	// DefaultURLExpiry is the default expiration time for signed URLs (15 minutes)
	DefaultURLExpiry = 15 * time.Minute
)

// SignedURLConfig holds configuration for generating signed URLs
type SignedURLConfig struct {
	R2PublicURL   string
	R2SecretKey   string
	DefaultExpiry time.Duration
}

// NewSignedURLConfig creates a new signed URL configuration from environment
func NewSignedURLConfig() *SignedURLConfig {
	return &SignedURLConfig{
		R2PublicURL:   os.Getenv("R2_PUBLIC_URL"),
		R2SecretKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		DefaultExpiry: DefaultURLExpiry,
	}
}

// GenerateSignedURL generates a time-limited signed URL for an R2 object key
// This prevents unauthorized access to sensitive documents (KYC, IDs, etc.)
func (c *SignedURLConfig) GenerateSignedURL(objectKey string, expiresIn time.Duration) (string, error) {
	if c.R2PublicURL == "" {
		return "", fmt.Errorf("R2_PUBLIC_URL not configured")
	}
	if c.R2SecretKey == "" {
		return "", fmt.Errorf("R2_SECRET_ACCESS_KEY not configured")
	}

	if expiresIn == 0 {
		expiresIn = c.DefaultExpiry
	}

	// Calculate expiration timestamp
	expiresAt := time.Now().Add(expiresIn).Unix()

	// Build base URL
	baseURL := fmt.Sprintf("%s/%s", c.R2PublicURL, objectKey)

	// Create signature payload: objectKey + expiresAt + secretKey
	payload := fmt.Sprintf("%s:%d:%s", objectKey, expiresAt, c.R2SecretKey)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(c.R2SecretKey))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	// Build signed URL with query parameters
	signedURL := fmt.Sprintf("%s?expires=%d&signature=%s", baseURL, expiresAt, signature)

	return signedURL, nil
}

// VerifySignedURL verifies if a signed URL is valid and not expired
// Use this in middleware to protect R2 access
func (c *SignedURLConfig) VerifySignedURL(fullURL string) error {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Extract query parameters
	expiresStr := parsedURL.Query().Get("expires")
	signature := parsedURL.Query().Get("signature")

	if expiresStr == "" || signature == "" {
		return fmt.Errorf("missing expires or signature parameter")
	}

	// Parse expiration timestamp
	var expiresAt int64
	if _, err := fmt.Sscanf(expiresStr, "%d", &expiresAt); err != nil {
		return fmt.Errorf("invalid expires parameter: %w", err)
	}

	// Check if URL has expired
	if time.Now().Unix() > expiresAt {
		return fmt.Errorf("URL has expired")
	}

	// Extract object key from path
	objectKey := parsedURL.Path[1:] // Remove leading slash

	// Recreate signature payload
	payload := fmt.Sprintf("%s:%d:%s", objectKey, expiresAt, c.R2SecretKey)

	// Compute expected signature
	h := hmac.New(sha256.New, []byte(c.R2SecretKey))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures (constant-time comparison)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// GetObjectKey extracts the object key from a full R2 URL or path
func GetObjectKey(urlOrPath string) string {
	// If it's a full URL, parse it
	if parsedURL, err := url.Parse(urlOrPath); err == nil && parsedURL.Host != "" {
		return parsedURL.Path[1:] // Remove leading slash
	}
	// Otherwise, assume it's already an object key
	return urlOrPath
}

// BuildObjectKey constructs an R2 object key for document storage
// Format: documents/{user_id}/{document_type}_{timestamp}.{extension}
func BuildObjectKey(userID, documentType, extension string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("documents/%s/%s_%d.%s", userID, documentType, timestamp, extension)
}
