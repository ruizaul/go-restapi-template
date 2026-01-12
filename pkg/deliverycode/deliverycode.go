// Package deliverycode provides cryptographically secure delivery code generation and verification
package deliverycode

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	// CodeLength is the length of the delivery code (4 digits)
	CodeLength = 4
	// MaxAttempts is the maximum number of failed verification attempts
	MaxAttempts = 3
)

var (
	// ErrMaxAttemptsReached indicates too many failed verification attempts
	ErrMaxAttemptsReached = errors.New("maximum delivery code attempts reached")
	// ErrInvalidCode indicates the delivery code is invalid
	ErrInvalidCode = errors.New("invalid delivery code")
	// ErrInvalidFormat indicates the code format is invalid
	ErrInvalidFormat = errors.New("invalid delivery code format")
)

// GenerateCode generates a cryptographically secure random 4-digit code
// Uses crypto/rand instead of math/rand for security
func GenerateCode() (string, error) {
	// Generate random number between 0 and 9999
	max := big.NewInt(10000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure delivery code: %w", err)
	}

	// Format as 4-digit string with leading zeros
	code := fmt.Sprintf("%04d", n.Int64())
	return code, nil
}

// ValidateFormat checks if a delivery code has valid format (4 digits)
func ValidateFormat(code string) bool {
	if len(code) != CodeLength {
		return false
	}

	// Check if all characters are digits
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// VerifyCode checks if the provided code matches the expected code
func VerifyCode(providedCode, expectedCode string) bool {
	// Constant-time comparison to prevent timing attacks
	if len(providedCode) != len(expectedCode) {
		return false
	}

	match := true
	for i := 0; i < len(providedCode); i++ {
		if providedCode[i] != expectedCode[i] {
			match = false
		}
	}

	return match
}
