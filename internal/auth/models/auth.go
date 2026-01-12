package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	CreatedAt      time.Time  `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-01-15T10:30:00Z"`
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BirthDate      *time.Time `json:"birth_date,omitempty" example:"1990-05-15T00:00:00Z"`
	OTPExpiresAt   *time.Time `json:"-"`
	OTPLockedUntil *time.Time `json:"-"` // New: Lockout timestamp for rate limiting
	Name           string     `json:"name" example:"Juan Pérez"`
	FirstName      string     `json:"first_name,omitempty" example:"Juan"`
	LastName       string     `json:"last_name,omitempty" example:"Pérez"`
	MotherLastName string     `json:"mother_last_name,omitempty" example:"González"`
	Email          string     `json:"email" example:"juan.perez@example.com"`
	Phone          string     `json:"phone,omitempty" example:"+525512345678"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role" enums:"customer,merchant,driver,admin" example:"customer"`
	AccountStatus  string     `json:"account_status" enums:"pending,active,suspended" example:"active"`
	OTPCode        string     `json:"-"` // DEPRECATED: Use OTPHash instead
	OTPHash        string     `json:"-"` // New: SHA-256 hash of OTP
	PhoneVerified  bool       `json:"phone_verified" example:"true"`
	OTPAttempts    int        `json:"-"` // New: Failed verification attempt counter
}

// RegisterRequest represents the request body for user registration
// Supports two modes:
// Mode 1 (Send OTP): Only phone is provided
// Mode 2 (Complete Registration): All fields including OTP are provided
type RegisterRequest struct {
	// Step 1 fields (optional for OTP send)
	FirstName      string `json:"first_name,omitempty" binding:"omitempty,min=2,max=100" example:"Juan"`
	LastName       string `json:"last_name,omitempty" binding:"omitempty,min=2,max=100" example:"Pérez"`
	MotherLastName string `json:"mother_last_name,omitempty" binding:"omitempty,max=100" example:"González"`
	BirthDate      string `json:"birth_date,omitempty" binding:"omitempty" example:"1990-05-15"`

	// Step 2 fields (required for both modes)
	Phone string `json:"phone" binding:"required,e164" example:"+525512345678"`
	OTP   string `json:"otp,omitempty" binding:"omitempty,len=6,numeric" example:"123456"`

	// Step 3 fields (required for complete registration)
	Email    string `json:"email,omitempty" binding:"omitempty,email" example:"juan.perez@example.com"`
	Password string `json:"password,omitempty" binding:"omitempty,min=6,max=72" example:"SecurePass123!"`
}

// VerifyOTPRequest represents the request body for OTP verification
type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,e164" example:"+525512345678"`
	OTP   string `json:"otp" binding:"required,len=6,numeric" example:"123456"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"juan.perez@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAiLCJleHAiOjE3Mzk5Nzk2MDAsImlhdCI6MTczNzM4NzYwMCwidHlwZSI6InJlZnJlc2gifQ.dGhpc2lzYW1vY2tzaWduYXR1cmU"`
}

// AuthResponse represents the response body for authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAiLCJleHAiOjE3MzczOTA2MDAsImlhdCI6MTczNzM4NzYwMCwidHlwZSI6ImFjY2VzcyJ9.dGhpc2lzYW1vY2tzaWduYXR1cmU"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAiLCJleHAiOjE3Mzk5Nzk2MDAsImlhdCI6MTczNzM4NzYwMCwidHlwZSI6InJlZnJlc2gifQ.dGhpc2lzYW1vY2tzaWduYXR1cmU"`
	User         User   `json:"user"`
}

// LoginResponse wraps login data in JSend format
type LoginResponse struct {
	Status string       `json:"status" example:"success"`
	Data   AuthResponse `json:"data"`
}

// RefreshResponse wraps refresh token data in JSend format
type RefreshResponse struct {
	Status string       `json:"status" example:"success"`
	Data   AuthResponse `json:"data"`
}

// OTPSentResponse represents the response after OTP is sent
type OTPSentResponse struct {
	OTPExpiresAt time.Time `json:"otp_expires_at" example:"2025-01-20T12:35:00Z"`
	Phone        string    `json:"phone" example:"+525512345678"`
	Message      string    `json:"message" example:"OTP sent to phone"`
}

// RegisterResponse wraps registration data in JSend format
type RegisterResponse struct {
	Status string          `json:"status" example:"success"`
	Data   OTPSentResponse `json:"data"`
}

// VerifyOTPResponse represents the response after OTP verification
type VerifyOTPResponse struct {
	Phone    string `json:"phone" example:"+525512345678"`
	Message  string `json:"message" example:"Phone verified successfully"`
	Verified bool   `json:"verified" example:"true"`
}

// VerifyOTPResponseWrapper wraps OTP verification in JSend format
type VerifyOTPResponseWrapper struct {
	Status string            `json:"status" example:"success"`
	Data   VerifyOTPResponse `json:"data"`
}

// CompleteRegistrationResponse wraps complete registration with tokens in JSend format
type CompleteRegistrationResponse struct {
	Status string       `json:"status" example:"success"`
	Data   AuthResponse `json:"data"`
}

// RefreshToken represents a stored refresh token in the database
type RefreshToken struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	TokenHash     string     `json:"token_hash"`
	DeviceInfo    string     `json:"device_info,omitempty"`
	DeviceID      string     `json:"device_id,omitempty"` // New: Strong device identifier
	IPAddress     string     `json:"ip_address,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"` // New: Token reuse detection
	Revoked       bool       `json:"revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason string     `json:"revoked_reason,omitempty"` // New: Why was it revoked
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`     // New: Soft delete
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LogoutResponse represents the response after logout
type LogoutResponse struct {
	Message string `json:"message" example:"Sesión cerrada exitosamente"`
}

// LogoutResponseWrapper wraps logout response in JSend format
type LogoutResponseWrapper struct {
	Status string         `json:"status" example:"success"`
	Data   LogoutResponse `json:"data"`
}

// ActiveSession represents an active user session
type ActiveSession struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceInfo string    `json:"device_info,omitempty" example:"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)"`
	IPAddress  string    `json:"ip_address,omitempty" example:"192.168.1.1"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T10:30:00Z"`
	ExpiresAt  time.Time `json:"expires_at" example:"2025-04-15T10:30:00Z"`
}

// ActiveSessionsResponse represents the response with all active sessions
type ActiveSessionsResponse struct {
	Sessions []ActiveSession `json:"sessions"`
}

// ActiveSessionsResponseWrapper wraps active sessions in JSend format
type ActiveSessionsResponseWrapper struct {
	Status string                 `json:"status" example:"success"`
	Data   ActiveSessionsResponse `json:"data"`
}
