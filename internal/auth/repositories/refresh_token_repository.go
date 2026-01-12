package repositories

import (
	"database/sql"
	"errors"
	"time"

	"tacoshare-delivery-api/internal/auth/models"

	"github.com/google/uuid"
)

// RefreshTokenRepository handles data access for refresh tokens
type RefreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// SaveRefreshToken stores a new refresh token in the database
func (r *RefreshTokenRepository) SaveRefreshToken(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, device_info, device_id, ip_address, expires_at, created_at, revoked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.DeviceInfo,
		token.DeviceID,
		token.IPAddress,
		token.ExpiresAt,
		token.CreatedAt,
		token.Revoked,
	)
	return err
}

// FindByTokenHash finds a refresh token by its hash
func (r *RefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, device_id, ip_address,
		       expires_at, created_at, last_used_at, revoked, revoked_at, revoked_reason
		FROM refresh_tokens
		WHERE token_hash = $1 AND deleted_at IS NULL
	`

	token := &models.RefreshToken{}
	var deviceInfo, deviceID, ipAddress, revokedReason sql.NullString
	var revokedAt, lastUsedAt sql.NullTime

	err := r.db.QueryRow(query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&deviceInfo,
		&deviceID,
		&ipAddress,
		&token.ExpiresAt,
		&token.CreatedAt,
		&lastUsedAt,
		&token.Revoked,
		&revokedAt,
		&revokedReason,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if deviceInfo.Valid {
		token.DeviceInfo = deviceInfo.String
	}
	if deviceID.Valid {
		token.DeviceID = deviceID.String
	}
	if ipAddress.Valid {
		token.IPAddress = ipAddress.String
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	if revokedReason.Valid {
		token.RevokedReason = revokedReason.String
	}

	return token, nil
}

// RevokeToken marks a refresh token as revoked
func (r *RefreshTokenRepository) RevokeToken(tokenHash string) error {
	return r.RevokeTokenWithReason(tokenHash, "logout")
}

// RevokeTokenWithReason marks a refresh token as revoked with a specific reason
func (r *RefreshTokenRepository) RevokeTokenWithReason(tokenHash, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE, revoked_at = $1, revoked_reason = $2
		WHERE token_hash = $3
	`
	_, err := r.db.Exec(query, time.Now(), reason, tokenHash)
	return err
}

// UpdateLastUsedAt updates the last_used_at timestamp for token reuse detection
func (r *RefreshTokenRepository) UpdateLastUsedAt(tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET last_used_at = $1
		WHERE token_hash = $2
	`
	_, err := r.db.Exec(query, time.Now(), tokenHash)
	return err
}

// RevokeAllUserTokens marks all refresh tokens for a user as revoked
func (r *RefreshTokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	return r.RevokeAllUserTokensWithReason(userID, "logout_all_devices")
}

// RevokeAllUserTokensWithReason marks all refresh tokens for a user as revoked with a reason
func (r *RefreshTokenRepository) RevokeAllUserTokensWithReason(userID uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE, revoked_at = $1, revoked_reason = $2
		WHERE user_id = $3 AND revoked = FALSE
	`
	_, err := r.db.Exec(query, time.Now(), reason, userID)
	return err
}

// GetUserActiveSessions retrieves all active (non-revoked, non-expired) sessions for a user
func (r *RefreshTokenRepository) GetUserActiveSessions(userID uuid.UUID) ([]models.ActiveSession, error) {
	query := `
		SELECT id, device_info, ip_address, created_at, expires_at
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked = FALSE AND expires_at > $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID, time.Now())
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't override the return error
			_ = closeErr
		}
	}()

	var sessions []models.ActiveSession
	for rows.Next() {
		var session models.ActiveSession
		var deviceInfo, ipAddress sql.NullString

		if err := rows.Scan(
			&session.ID,
			&deviceInfo,
			&ipAddress,
			&session.CreatedAt,
			&session.ExpiresAt,
		); err != nil {
			return nil, err
		}

		if deviceInfo.Valid {
			session.DeviceInfo = deviceInfo.String
		}
		if ipAddress.Valid {
			session.IPAddress = ipAddress.String
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *RefreshTokenRepository) CleanupExpiredTokens() (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1
	`
	result, err := r.db.Exec(query, time.Now())
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
