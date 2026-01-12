package repositories

import (
	"database/sql"
	"fmt"

	"tacoshare-delivery-api/internal/merchants/models"

	"github.com/google/uuid"
)

// MerchantRepository handles database operations for merchants
type MerchantRepository struct {
	db *sql.DB
}

// NewMerchantRepository creates a new merchant repository
func NewMerchantRepository(db *sql.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

// Create creates a new merchant
func (r *MerchantRepository) Create(merchant *models.Merchant) error {
	query := `
		INSERT INTO merchants (
			user_id, business_name, business_type, phone, email,
			address, latitude, longitude, city, state, postal_code, country
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, status, rating, total_orders, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		merchant.UserID,
		merchant.BusinessName,
		merchant.BusinessType,
		merchant.Phone,
		merchant.Email,
		merchant.Address,
		merchant.Latitude,
		merchant.Longitude,
		merchant.City,
		merchant.State,
		merchant.PostalCode,
		merchant.Country,
	).Scan(
		&merchant.ID,
		&merchant.Status,
		&merchant.Rating,
		&merchant.TotalOrders,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	return nil
}

// FindByID finds a merchant by ID
func (r *MerchantRepository) FindByID(id uuid.UUID) (*models.Merchant, error) {
	query := `
		SELECT id, user_id, business_name, business_type, phone, email,
			address, latitude, longitude, city, state, postal_code, country,
			status, rating, total_orders, created_at, updated_at
		FROM merchants
		WHERE id = $1
	`

	merchant := &models.Merchant{}
	err := r.db.QueryRow(query, id).Scan(
		&merchant.ID,
		&merchant.UserID,
		&merchant.BusinessName,
		&merchant.BusinessType,
		&merchant.Phone,
		&merchant.Email,
		&merchant.Address,
		&merchant.Latitude,
		&merchant.Longitude,
		&merchant.City,
		&merchant.State,
		&merchant.PostalCode,
		&merchant.Country,
		&merchant.Status,
		&merchant.Rating,
		&merchant.TotalOrders,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %w", err)
	}

	return merchant, nil
}

// FindByUserID finds a merchant by user ID
func (r *MerchantRepository) FindByUserID(userID uuid.UUID) (*models.Merchant, error) {
	query := `
		SELECT id, user_id, business_name, business_type, phone, email,
			address, latitude, longitude, city, state, postal_code, country,
			status, rating, total_orders, created_at, updated_at
		FROM merchants
		WHERE user_id = $1
	`

	merchant := &models.Merchant{}
	err := r.db.QueryRow(query, userID).Scan(
		&merchant.ID,
		&merchant.UserID,
		&merchant.BusinessName,
		&merchant.BusinessType,
		&merchant.Phone,
		&merchant.Email,
		&merchant.Address,
		&merchant.Latitude,
		&merchant.Longitude,
		&merchant.City,
		&merchant.State,
		&merchant.PostalCode,
		&merchant.Country,
		&merchant.Status,
		&merchant.Rating,
		&merchant.TotalOrders,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant by user ID: %w", err)
	}

	return merchant, nil
}

// Update updates merchant information
func (r *MerchantRepository) Update(merchant *models.Merchant) error {
	query := `
		UPDATE merchants
		SET business_name = $1, phone = $2, email = $3, address = $4,
			latitude = $5, longitude = $6, postal_code = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		merchant.BusinessName,
		merchant.Phone,
		merchant.Email,
		merchant.Address,
		merchant.Latitude,
		merchant.Longitude,
		merchant.PostalCode,
		merchant.ID,
	).Scan(&merchant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	return nil
}

// UpdateStatus updates merchant status
func (r *MerchantRepository) UpdateStatus(id uuid.UUID, status string) error {
	query := `UPDATE merchants SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update merchant status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// IncrementTotalOrders increments the total orders count
func (r *MerchantRepository) IncrementTotalOrders(id uuid.UUID) error {
	query := `UPDATE merchants SET total_orders = total_orders + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment total orders: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// FindAll finds all merchants with optional filters
func (r *MerchantRepository) FindAll(city string, businessType string, status string) ([]models.Merchant, error) {
	query := `
		SELECT id, user_id, business_name, business_type, phone, email,
			address, latitude, longitude, city, state, postal_code, country,
			status, rating, total_orders, created_at, updated_at
		FROM merchants
		WHERE 1=1
	`
	args := []any{}
	argCount := 1

	if city != "" {
		query += fmt.Sprintf(" AND city = $%d", argCount)
		args = append(args, city)
		argCount++
	}

	if businessType != "" {
		query += fmt.Sprintf(" AND business_type = $%d", argCount)
		args = append(args, businessType)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	query += " ORDER BY business_name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find merchants: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	merchants := []models.Merchant{}
	for rows.Next() {
		var m models.Merchant
		err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.BusinessName,
			&m.BusinessType,
			&m.Phone,
			&m.Email,
			&m.Address,
			&m.Latitude,
			&m.Longitude,
			&m.City,
			&m.State,
			&m.PostalCode,
			&m.Country,
			&m.Status,
			&m.Rating,
			&m.TotalOrders,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return merchants, nil
}
