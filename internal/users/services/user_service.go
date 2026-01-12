package services

import (
	"errors"

	"tacoshare-delivery-api/internal/users/models"
	"tacoshare-delivery-api/internal/users/repositories"

	"github.com/google/uuid"
)

// UserService handles business logic for users
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserProfile retrieves a user's profile by ID
func (s *UserService) GetUserProfile(userID uuid.UUID) (*models.UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
