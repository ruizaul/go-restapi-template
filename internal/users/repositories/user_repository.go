package repositories

import (
	"database/sql"
	"errors"

	"tacoshare-delivery-api/internal/users/models"

	"github.com/google/uuid"
)

// UserRepository handles data access for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*models.UserProfile, error) {
	query := `
		SELECT id, name, email, phone, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.UserProfile{}
	var phone sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}

	return user, nil
}
