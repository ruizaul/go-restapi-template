package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"go-api-template/internal/users/models"
	"go-api-template/internal/users/repositories"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// UserService handles business logic for users
type UserService struct {
	repo *repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

// List retrieves all users with pagination
func (s *UserService) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

// Update updates a user's information
func (s *UserService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check if new email already exists (if changing email)
	if req.Email != "" && req.Email != user.Email {
		existing, err := s.repo.GetByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete soft deletes a user
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return ErrUserNotFound
	}
	return err
}
