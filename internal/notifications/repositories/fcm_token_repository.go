package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tacoshare-delivery-api/internal/notifications/models"

	"github.com/google/uuid"
)

// FCMTokenRepository handles data access for FCM tokens
type FCMTokenRepository struct {
	db *sql.DB
}

// NewFCMTokenRepository creates a new FCM token repository
func NewFCMTokenRepository(db *sql.DB) *FCMTokenRepository {
	return &FCMTokenRepository{db: db}
}

// Create creates a new FCM token or updates if it already exists
func (r *FCMTokenRepository) Create(ctx context.Context, userID uuid.UUID, token string, deviceType models.DeviceType, deviceID *string) (*models.FCMToken, error) {
	query := `
		INSERT INTO fcm_tokens (user_id, token, device_type, device_id, is_active, created_at, updated_at, last_used_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW(), NOW())
		ON CONFLICT (token)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			device_type = EXCLUDED.device_type,
			device_id = EXCLUDED.device_id,
			is_active = true,
			updated_at = NOW(),
			last_used_at = NOW()
		RETURNING id, user_id, token, device_type, device_id, is_active, created_at, updated_at, last_used_at
	`

	var fcmToken models.FCMToken
	err := r.db.QueryRowContext(ctx, query, userID, token, deviceType, deviceID).Scan(
		&fcmToken.ID,
		&fcmToken.UserID,
		&fcmToken.Token,
		&fcmToken.DeviceType,
		&fcmToken.DeviceID,
		&fcmToken.IsActive,
		&fcmToken.CreatedAt,
		&fcmToken.UpdatedAt,
		&fcmToken.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create fcm token: %w", err)
	}

	return &fcmToken, nil
}

// FindByToken finds a FCM token by its token string
func (r *FCMTokenRepository) FindByToken(ctx context.Context, token string) (*models.FCMToken, error) {
	query := `
		SELECT id, user_id, token, device_type, device_id, is_active, created_at, updated_at, last_used_at
		FROM fcm_tokens
		WHERE token = $1
	`

	var fcmToken models.FCMToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&fcmToken.ID,
		&fcmToken.UserID,
		&fcmToken.Token,
		&fcmToken.DeviceType,
		&fcmToken.DeviceID,
		&fcmToken.IsActive,
		&fcmToken.CreatedAt,
		&fcmToken.UpdatedAt,
		&fcmToken.LastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find fcm token: %w", err)
	}

	return &fcmToken, nil
}

// FindActiveByUserID finds all active FCM tokens for a user
func (r *FCMTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]models.FCMToken, error) {
	query := `
		SELECT id, user_id, token, device_type, device_id, is_active, created_at, updated_at, last_used_at
		FROM fcm_tokens
		WHERE user_id = $1 AND is_active = true
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active fcm tokens: %w", err)
	}
	//nolint:errcheck // rows.Close() error is not critical in defer
	defer func() { _ = rows.Close() }()

	var tokens []models.FCMToken
	for rows.Next() {
		var token models.FCMToken
		if err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.DeviceType,
			&token.DeviceID,
			&token.IsActive,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fcm token: %w", err)
		}
		tokens = append(tokens, token)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating fcm token rows: %w", err)
	}

	return tokens, nil
}

// Deactivate marks a FCM token as inactive
func (r *FCMTokenRepository) Deactivate(ctx context.Context, token string) error {
	query := `
		UPDATE fcm_tokens
		SET is_active = false, updated_at = NOW()
		WHERE token = $1
	`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to deactivate fcm token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("fcm token not found")
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for a token
func (r *FCMTokenRepository) UpdateLastUsed(ctx context.Context, token string) error {
	query := `
		UPDATE fcm_tokens
		SET last_used_at = $1
		WHERE token = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

// DeleteInactiveOlderThan deletes inactive tokens older than the specified duration
func (r *FCMTokenRepository) DeleteInactiveOlderThan(ctx context.Context, duration time.Duration) error {
	query := `
		DELETE FROM fcm_tokens
		WHERE is_active = false AND updated_at < $1
	`

	cutoffTime := time.Now().Add(-duration)
	_, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to delete inactive tokens: %w", err)
	}

	return nil
}
