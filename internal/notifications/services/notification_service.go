package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"tacoshare-delivery-api/internal/notifications/models"
	"tacoshare-delivery-api/internal/notifications/repositories"

	"github.com/google/uuid"
)

const (
	// Error messages
	errNotificationNotFound     = "notification not found"
	errUnauthorizedNotification = "unauthorized access to notification"
)

var (
	// Error variables for staticcheck SA1006 compliance
	errNotificationNotFoundVar     = errors.New(errNotificationNotFound)
	errUnauthorizedNotificationVar = errors.New(errUnauthorizedNotification)
)

// NotificationService handles business logic for notifications
type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
	fcmTokenRepo     *repositories.FCMTokenRepository
	fcmService       *FCMService
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo *repositories.NotificationRepository,
	fcmTokenRepo *repositories.FCMTokenRepository,
	fcmService *FCMService,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		fcmTokenRepo:     fcmTokenRepo,
		fcmService:       fcmService,
	}
}

// CreateAndSend creates a notification in the database and sends it via FCM
func (s *NotificationService) CreateAndSend(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	// Create notification in database
	notification, err := s.notificationRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Get all active FCM tokens for the user
	tokens, err := s.fcmTokenRepo.FindActiveByUserID(ctx, req.UserID)
	if err != nil {
		return notification, nil // Return notification even if FCM fails
	}

	if len(tokens) == 0 {
		return notification, nil
	}

	// Convert data to string map for FCM
	var dataMap map[string]string
	if req.Data != nil {
		var tempData map[string]any
		if err := json.Unmarshal(req.Data, &tempData); err == nil {
			var convertErr error
			dataMap, convertErr = ConvertDataToStringMap(tempData)
			_ = convertErr
		}
	}

	// Add notification ID to data
	if dataMap == nil {
		dataMap = make(map[string]string)
	}
	dataMap["notification_id"] = fmt.Sprintf("%d", notification.ID)
	dataMap["notification_type"] = string(notification.NotificationType)

	// Extract token strings
	tokenStrings := make([]string, len(tokens))
	for i, token := range tokens {
		tokenStrings[i] = token.Token
	}

	// Send notification via FCM
	if len(tokenStrings) == 1 {
		err = s.fcmService.SendNotification(ctx, tokenStrings[0], req.Title, req.Body, dataMap)
		_ = err
	} else {
		_, err = s.fcmService.SendNotificationToMultiple(ctx, tokenStrings, req.Title, req.Body, dataMap)
		_ = err
	}

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error) {
	notification, err := s.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if notification == nil {
		return nil, errNotificationNotFoundVar
	}

	// Verify the notification belongs to the user
	if notification.UserID != userID {
		return nil, errUnauthorizedNotificationVar
	}

	return notification, nil
}

// ListNotifications retrieves notifications for a user with pagination
func (s *NotificationService) ListNotifications(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Notification, models.PaginationMetadata, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	notifications, total, err := s.notificationRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, models.PaginationMetadata{}, fmt.Errorf("failed to list notifications: %w", err)
	}

	totalPages := (total + limit - 1) / limit

	pagination := models.PaginationMetadata{
		CurrentPage: page,
		PerPage:     limit,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	if pagination.HasNext {
		pagination.NextURL = fmt.Sprintf("/api/v1/notifications?page=%d&limit=%d", page+1, limit)
	}
	if pagination.HasPrevious {
		pagination.PreviousURL = fmt.Sprintf("/api/v1/notifications?page=%d&limit=%d", page-1, limit)
	}

	return notifications, pagination, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Verify the notification belongs to the user
	notification, err := s.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification == nil {
		return errNotificationNotFoundVar
	}

	if notification.UserID != userID {
		return errUnauthorizedNotificationVar
	}

	return s.notificationRepo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Verify the notification belongs to the user
	notification, err := s.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}

	if notification == nil {
		return errNotificationNotFoundVar
	}

	if notification.UserID != userID {
		return errUnauthorizedNotificationVar
	}

	return s.notificationRepo.Delete(ctx, id)
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}

// RegisterToken registers a new FCM token for a user
func (s *NotificationService) RegisterToken(ctx context.Context, userID uuid.UUID, token string, deviceType models.DeviceType, deviceID *string) (*models.FCMToken, error) {
	return s.fcmTokenRepo.Create(ctx, userID, token, deviceType, deviceID)
}

// UnregisterToken deactivates a FCM token
func (s *NotificationService) UnregisterToken(ctx context.Context, token string) error {
	return s.fcmTokenRepo.Deactivate(ctx, token)
}
