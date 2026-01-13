package services

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"go-api-template/internal/auth/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrNameRequired       = errors.New("name is required")
)

// emailRegex is a simple email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// AuthService handles authentication business logic
type AuthService struct {
	db         *sql.DB
	jwtService *JWTService
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB, jwtService *JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
	// Validate input
	if err := s.validateRegistration(req); err != nil {
		return nil, nil, err
	}

	// Check if email already exists
	var exists bool
	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)",
		req.Email,
	).Scan(&exists)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// Create user
	user := &models.AuthUser{
		ID:    uuid.New(),
		Email: req.Email,
		Name:  req.Name,
	}
	now := time.Now().UTC()

	err = s.db.QueryRowContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		user.ID, user.Email, user.Name, string(hashedPassword), now, now,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthUser, *models.TokenPair, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, nil, ErrInvalidCredentials
	}

	// Get user by email
	var user models.AuthUser
	var passwordHash string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at
		 FROM users
		 WHERE email = $1 AND deleted_at IS NULL`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Name, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// RefreshTokens generates new tokens from a valid refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*models.AuthUser, *models.TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	// Get user from database to ensure they still exist and are not deleted
	var user models.AuthUser
	err = s.db.QueryRowContext(ctx,
		`SELECT id, email, name, created_at, updated_at
		 FROM users
		 WHERE id = $1 AND deleted_at IS NULL`,
		claims.UserID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrUserNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	// Generate new tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// GetProfile retrieves the user profile by ID
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.AuthUser, error) {
	var user models.AuthUser

	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, created_at, updated_at
		 FROM users
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// validateRegistration validates registration input
func (s *AuthService) validateRegistration(req *models.RegisterRequest) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if !emailRegex.MatchString(req.Email) {
		return ErrInvalidEmail
	}

	if len(req.Password) < 8 {
		return ErrWeakPassword
	}

	return nil
}
