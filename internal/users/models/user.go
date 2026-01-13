package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Email     string     `json:"email" db:"email"`
	Name      string     `json:"name" db:"name"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// UserResponse represents a successful user response (JSend format)
type UserResponse struct {
	Status string `json:"status" example:"success"`
	Data   User   `json:"data"`
}

// UsersListResponse represents a successful list of users response
type UsersListResponse struct {
	Status string `json:"status" example:"success"`
	Data   []User `json:"data"`
}
