// Package otp provides one-time password generation and SMS delivery via Twilio
package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"tacoshare-delivery-api/pkg/validator"
	"time"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

const (
	// OTPLength is the number of digits in the OTP
	OTPLength = 6
	// OTPExpirationMinutes is how long the OTP is valid
	OTPExpirationMinutes = 10
	// MaxOTPAttempts is the maximum number of failed verification attempts before lockout
	MaxOTPAttempts = 3
	// OTPLockoutMinutes is how long to lock the account after max attempts
	OTPLockoutMinutes = 15
)

var (
	// ErrOTPLocked indicates the account is temporarily locked due to too many failed attempts
	ErrOTPLocked = errors.New("account temporarily locked due to too many failed OTP attempts")
	// ErrOTPExpired indicates the OTP has expired
	ErrOTPExpired = errors.New("OTP has expired")
	// ErrInvalidOTP indicates the OTP code is invalid
	ErrInvalidOTP = errors.New("invalid OTP code")
)

var (
	// twilioClient is the global Twilio client
	twilioClient *twilio.RestClient
	// twilioFromPhone is the Twilio phone number to send from
	twilioFromPhone string
	// twilioEnabled indicates if Twilio is configured
	twilioEnabled bool
)

// InitializeTwilio initializes the Twilio client with credentials
// If credentials are empty, runs in mock mode
func InitializeTwilio(accountSID, apiKey, apiSecret, fromPhone string, enabled bool) {
	if !enabled || accountSID == "" || apiKey == "" || apiSecret == "" || fromPhone == "" {
		twilioEnabled = false
		return
	}

	twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   apiKey,
		Password:   apiSecret,
		AccountSid: accountSID,
	})
	twilioFromPhone = fromPhone
	twilioEnabled = true
}

// GenerateOTP generates a random 6-digit OTP code
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000) // 0-999999 range
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Pad with leading zeros if necessary
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// GetExpirationTime returns the expiration time for an OTP
func GetExpirationTime() time.Time {
	return time.Now().Add(OTPExpirationMinutes * time.Minute)
}

// IsExpired checks if an OTP has expired
func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// SendOTP sends an OTP via SMS using Twilio (or logs it in mock mode)
func SendOTP(phone, code string) error {
	// Normalize phone to E.164 format (+526621816014)
	normalizedPhone := validator.NormalizePhone(phone)

	// Mock mode - just return without sending
	if !twilioEnabled || twilioClient == nil {
		return nil
	}

	// Prepare SMS message
	messageBody := fmt.Sprintf("Your TacoShare verification code is: %s (expires in %d minutes)", code, OTPExpirationMinutes)

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(normalizedPhone)
	params.SetFrom(twilioFromPhone)
	params.SetBody(messageBody)

	// Send SMS via Twilio
	_, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send OTP SMS: %w", err)
	}

	return nil
}

// ValidateOTPFormat checks if an OTP code has valid format (6 digits)
func ValidateOTPFormat(code string) bool {
	if len(code) != OTPLength {
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

// HashOTP creates a SHA-256 hash of the OTP with server-side pepper
// This ensures OTPs are never stored in plaintext
func HashOTP(otpCode string) string {
	// Get pepper from environment (server-side secret, never in DB)
	pepper := os.Getenv("OTP_PEPPER")
	if pepper == "" {
		// Fallback to JWT_SECRET if OTP_PEPPER not set (but should set dedicated pepper)
		pepper = os.Getenv("JWT_SECRET")
	}

	// Combine OTP + pepper before hashing
	combined := otpCode + pepper
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifyOTPHash verifies if the provided OTP matches the stored hash
func VerifyOTPHash(otpCode, storedHash string) bool {
	computedHash := HashOTP(otpCode)
	return computedHash == storedHash
}

// GetLockoutTime returns the lockout expiration time
func GetLockoutTime() time.Time {
	return time.Now().Add(OTPLockoutMinutes * time.Minute)
}

// IsLocked checks if an account is currently locked
func IsLocked(lockedUntil *time.Time) bool {
	if lockedUntil == nil {
		return false
	}
	return time.Now().Before(*lockedUntil)
}
