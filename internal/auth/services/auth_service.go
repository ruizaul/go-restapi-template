package services

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"tacoshare-delivery-api/internal/auth/models"
	"tacoshare-delivery-api/internal/auth/repositories"
	"tacoshare-delivery-api/pkg/authx"
	"tacoshare-delivery-api/pkg/otp"
	"tacoshare-delivery-api/pkg/validator"

	"github.com/google/uuid"
)

const (
	// defaultRefreshTokenExpiry is the default expiry duration for refresh tokens (90 days)
	defaultRefreshTokenExpiry = "2160h"
)

// AuthService handles business logic for authentication
type AuthService struct {
	userRepo         *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repositories.UserRepository, refreshTokenRepo *repositories.RefreshTokenRepository) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Login authenticates a user and stores refresh token in DB
func (s *AuthService) Login(req *models.LoginRequest, deviceInfo, ipAddress string) (*models.AuthResponse, error) {
	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Compare password
	// Verify password
	if err := authx.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	// Generate tokens
	accessToken, err := authx.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authx.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	tokenHash := authx.HashRefreshToken(refreshToken)
	expiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if expiryStr == "" {
		expiryStr = defaultRefreshTokenExpiry
	}
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 2160 * time.Hour
	}

	refreshTokenModel := &models.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  tokenHash,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		ExpiresAt:  time.Now().Add(expiry),
		CreatedAt:  time.Now(),
		Revoked:    false,
	}

	_ = s.refreshTokenRepo.SaveRefreshToken(refreshTokenModel)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// RefreshToken generates new tokens from a refresh token with rotation and theft detection
func (s *AuthService) RefreshToken(refreshToken, deviceInfo, ipAddress, deviceID string) (*models.AuthResponse, error) {
	// Validate refresh token JWT
	claims, err := authx.ValidateToken(refreshToken, authx.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token exists in database and is not revoked
	tokenHash := authx.HashRefreshToken(refreshToken)
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if storedToken == nil {
		return nil, errors.New("refresh token not found")
	}

	// THEFT DETECTION: If token is already revoked and being reused, it's likely stolen
	if storedToken.Revoked {
		// Revoke ALL tokens for this user (force re-login everywhere)
		_ = s.refreshTokenRepo.RevokeAllUserTokensWithReason(claims.UserID, "token_theft_detected")
		return nil, errors.New("refresh token has been revoked")
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}

	// DEVICE BINDING: Verify device_id matches (if device_id was stored)
	if storedToken.DeviceID != "" && deviceID != "" && storedToken.DeviceID != deviceID {
		// Revoke this token and require re-authentication
		_ = s.refreshTokenRepo.RevokeTokenWithReason(tokenHash, "device_mismatch")
		return nil, errors.New("device mismatch detected")
	}

	// Update last_used_at for reuse detection
	_ = s.refreshTokenRepo.UpdateLastUsedAt(tokenHash)

	// Get user from database to ensure they still exist
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// REVOKE old refresh token (rotation security)
	_ = s.refreshTokenRepo.RevokeTokenWithReason(tokenHash, "token_rotated")

	// Generate new tokens
	accessToken, err := authx.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := authx.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Store new refresh token in database with same device_id (binding continuity)
	newTokenHash := authx.HashRefreshToken(newRefreshToken)
	expiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if expiryStr == "" {
		expiryStr = defaultRefreshTokenExpiry
	}
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 2160 * time.Hour
	}

	newRefreshTokenModel := &models.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  newTokenHash,
		DeviceInfo: deviceInfo,
		DeviceID:   deviceID, // Preserve device_id for binding
		IPAddress:  ipAddress,
		ExpiresAt:  time.Now().Add(expiry),
		CreatedAt:  time.Now(),
		Revoked:    false,
	}

	_ = s.refreshTokenRepo.SaveRefreshToken(newRefreshTokenModel)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

// Register handles user registration with OTP support
// Mode 1: If only phone is provided, sends OTP
// Mode 2: If all fields + OTP are provided, completes registration
// Register handles two-step registration: OTP sending and verification
func (s *AuthService) Register(req *models.RegisterRequest) (any, error) {
	// Validate phone format
	if !validator.IsValidPhone(req.Phone) {
		return nil, errors.New("invalid phone format")
	}

	// Mode 1: Send OTP (only phone provided, no email/password)
	// Mode 1: Send OTP (when only phone is provided)
	if req.Email == "" && req.Password == "" && req.OTP == "" {
		return s.sendOTP(req.Phone)
	}

	// Mode 2: Complete registration (all fields + OTP provided)
	// Mode 2: Complete registration (with phone, email, password, and OTP)
	if req.Email != "" && req.Password != "" && req.OTP != "" {
		return s.completeRegistration(req)
	}

	return nil, errors.New("invalid request: provide either only phone (to send OTP) or all fields including OTP (to complete registration)")
}

// sendOTP generates and sends OTP to the provided phone number
// sendOTP generates and sends an OTP to the user's phone
func (s *AuthService) sendOTP(phone string) (*models.OTPSentResponse, error) {
	// Check if phone already exists and is verified
	exists, err := s.userRepo.PhoneExists(phone)
	if err != nil {
		return nil, err
	}

	// If phone exists, check if it's already verified and check lockout status
	if exists {
		user, err := s.userRepo.FindByPhoneWithOTPHash(phone)
		if err != nil {
			return nil, err
		}
		if user != nil {
			if otp.IsLocked(user.OTPLockedUntil) {
				return nil, otp.ErrOTPLocked
			}

			// Check if already fully registered
			if user.PhoneVerified && user.Email != "" {
				return nil, errors.New("phone number already registered")
			}
		}
	}

	// Generate OTP
	otpCode, err := otp.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash the OTP before storing (NEVER store plaintext)
	otpHash := otp.HashOTP(otpCode)
	expiresAt := otp.GetExpirationTime()

	// Save OTP hash to database (create or update pending user)
	if exists {
		// Update existing pending user
		err = s.userRepo.SaveOTPHash(phone, otpHash, sql.NullTime{Time: expiresAt, Valid: true})
	} else {
		// Create new pending user
		err = s.userRepo.CreatePendingUserWithHash(phone, otpHash, sql.NullTime{Time: expiresAt, Valid: true})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save OTP: %w", err)
	}

	// Send OTP via SMS (send plaintext to user, but we stored hash)
	if err := otp.SendOTP(phone, otpCode); err != nil {
		return nil, fmt.Errorf("failed to send OTP: %w", err)
	}

	return &models.OTPSentResponse{
		Phone:        phone,
		Message:      "OTP sent to phone",
		OTPExpiresAt: expiresAt,
	}, nil
}

// completeRegistration verifies OTP and creates the user account
//
//nolint:gocyclo // Complex registration completion with multiple validation steps
func (s *AuthService) completeRegistration(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate all required fields
	if req.FirstName == "" || req.LastName == "" || req.BirthDate == "" {
		return nil, errors.New("first_name, last_name, and birth_date are required")
	}

	// Validate email format
	if !validator.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validate password length
	if len(req.Password) < 6 || len(req.Password) > 72 {
		return nil, errors.New("password must be between 6 and 72 characters")
	}

	// Validate OTP format
	if !otp.ValidateOTPFormat(req.OTP) {
		return nil, errors.New("invalid OTP format")
	}

	// Parse birth date
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, errors.New("invalid birth_date format (use YYYY-MM-DD)")
	}

	// Validate age (must be 18+)
	age := time.Now().Year() - birthDate.Year()
	if age < 18 {
		return nil, errors.New("user must be at least 18 years old")
	}

	// Find user by phone
	user, err := s.userRepo.FindByPhoneWithOTPHash(req.Phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("phone number not found - please request OTP first")
	}

	// Check if phone is already verified
	if !user.PhoneVerified {
		return nil, errors.New("phone not verified - please verify OTP first")
	}

	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := authx.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Build full name
	fullName := req.FirstName + " " + req.LastName
	if req.MotherLastName != "" {
		fullName += " " + req.MotherLastName
	}

	// Complete user registration
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.MotherLastName = req.MotherLastName
	user.BirthDate = &birthDate
	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user.PasswordHash = hashedPassword
	user.Name = fullName
	user.Role = "driver"
	user.AccountStatus = "active"
	user.PhoneVerified = true

	// Update user in database
	if err := s.userRepo.CompleteRegistration(user); err != nil {
		return nil, fmt.Errorf("failed to complete registration: %w", err)
	}

	// Generate tokens
	accessToken, err := authx.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authx.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	tokenHash := authx.HashRefreshToken(refreshToken)
	expiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if expiryStr == "" {
		expiryStr = defaultRefreshTokenExpiry
	}
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 2160 * time.Hour
	}

	refreshTokenModel := &models.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  tokenHash,
		DeviceInfo: "", // Device info not available in registration flow
		IPAddress:  "", // IP address not available in registration flow
		ExpiresAt:  time.Now().Add(expiry),
		CreatedAt:  time.Now(),
		Revoked:    false,
	}

	_ = s.refreshTokenRepo.SaveRefreshToken(refreshTokenModel)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// VerifyOTP verifies the OTP code for a phone number with rate limiting
// VerifyOTP verifies an OTP code for a phone number
func (s *AuthService) VerifyOTP(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	// Validate phone format
	if !validator.IsValidPhone(req.Phone) {
		return nil, errors.New("invalid phone format")
	}

	// Validate OTP format
	if !otp.ValidateOTPFormat(req.OTP) {
		return nil, errors.New("invalid OTP format")
	}

	// Find user by phone
	user, err := s.userRepo.FindByPhoneWithOTPHash(req.Phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("phone number not found")
	}

	// Check if account is locked due to too many failed attempts
	if otp.IsLocked(user.OTPLockedUntil) {
		return nil, otp.ErrOTPLocked
	}

	// Check if OTP hash exists
	if user.OTPHash == "" {
		return nil, errors.New("no OTP found for this phone number")
	}

	// Check if OTP is expired
	if user.OTPExpiresAt == nil || otp.IsExpired(*user.OTPExpiresAt) {
		return nil, otp.ErrOTPExpired
	}

	// Verify OTP code against hash
	if !otp.VerifyOTPHash(req.OTP, user.OTPHash) {
		// Increment failed attempt counter
		_ = s.userRepo.IncrementOTPAttempts(req.Phone)

		// Check if max attempts reached
		if user.OTPAttempts+1 >= otp.MaxOTPAttempts {
			lockoutTime := otp.GetLockoutTime()
			_ = s.userRepo.LockOTPAccount(req.Phone, sql.NullTime{Time: lockoutTime, Valid: true})
			return nil, otp.ErrOTPLocked
		}

		return nil, otp.ErrInvalidOTP
	}

	// OTP verification successful - clear OTP data and mark phone as verified
	if err := s.userRepo.ClearOTPData(req.Phone); err != nil {
		return nil, fmt.Errorf("failed to verify phone: %w", err)
	}

	return &models.VerifyOTPResponse{
		Phone:    req.Phone,
		Verified: true,
		Message:  "Phone verified successfully",
	}, nil
}

// Logout revokes a specific refresh token
func (s *AuthService) Logout(refreshToken string) error {
	// Hash the token to find it in DB
	tokenHash := authx.HashRefreshToken(refreshToken)

	// Check if token exists
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return err
	}
	if storedToken == nil {
		return errors.New("refresh token not found")
	}

	// Revoke the token
	if err := s.refreshTokenRepo.RevokeToken(tokenHash); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// LogoutAllDevices revokes all refresh tokens for a user
func (s *AuthService) LogoutAllDevices(userID uuid.UUID) error {
	if err := s.refreshTokenRepo.RevokeAllUserTokens(userID); err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}
	return nil
}

// GetActiveSessions retrieves all active sessions for a user
func (s *AuthService) GetActiveSessions(userID uuid.UUID) ([]models.ActiveSession, error) {
	sessions, err := s.refreshTokenRepo.GetUserActiveSessions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active sessions: %w", err)
	}
	return sessions, nil
}
