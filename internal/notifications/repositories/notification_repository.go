package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tacoshare-delivery-api/internal/notifications/models"

	"github.com/google/uuid"
)

// NotificationRepository handles data access for notifications
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, title, body, data, notification_type, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, false, NOW())
		RETURNING id, user_id, title, body, data, notification_type, is_read, read_at, created_at
	`

	var notification models.Notification
	err := r.db.QueryRowContext(ctx, query,
		req.UserID,
		req.Title,
		req.Body,
		req.Data,
		req.NotificationType,
	).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Data,
		&notification.NotificationType,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &notification, nil
}

// FindByID finds a notification by ID
func (r *NotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, title, body, data, notification_type, is_read, read_at, created_at
		FROM notifications
		WHERE id = $1
	`

	var notification models.Notification
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Data,
		&notification.NotificationType,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	return &notification, nil
}

// FindByUserID finds all notifications for a user with pagination
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, user_id, title, body, data, notification_type, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find notifications: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Ignore error as it's a cleanup operation
			_ = closeErr
		}
	}()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.Data,
			&notification.NotificationType,
			&notification.IsRead,
			&notification.ReadAt,
			&notification.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating notification rows: %w", err)
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE user_id = $2 AND is_read = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// CountUnread counts unread notifications for a user
func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}
