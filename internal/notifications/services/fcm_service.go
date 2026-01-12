package services

import (
	"context"
	"encoding/json"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging operations
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service from a credentials file path
func NewFCMService(ctx context.Context, credentialsPath string) (*FCMService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %w", err)
	}

	return &FCMService{client: client}, nil
}

// NewFCMServiceFromJSON creates a new FCM service from a JSON credentials string
func NewFCMServiceFromJSON(ctx context.Context, credentialsJSON string) (*FCMService, error) {
	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %w", err)
	}

	return &FCMService{client: client}, nil
}

// SendNotification sends a notification to a single device token
func (s *FCMService) SendNotification(ctx context.Context, token, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
		},
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

// SendNotificationToMultiple sends a notification to multiple device tokens
func (s *FCMService) SendNotificationToMultiple(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
		},
	}

	response, err := s.client.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("error sending multicast message: %w", err)
	}

	return response, nil
}

// SendDataOnlyNotification sends a data-only notification (silent notification)
func (s *FCMService) SendDataOnlyNotification(ctx context.Context, token string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending data message: %w", err)
	}

	return nil
}

// SendToTopic sends a notification to all devices subscribed to a topic
func (s *FCMService) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending topic message: %w", err)
	}

	return nil
}

// SubscribeToTopic subscribes device tokens to a topic
func (s *FCMService) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := s.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("error subscribing to topic: %w", err)
	}

	return nil
}

// UnsubscribeFromTopic unsubscribes device tokens from a topic
func (s *FCMService) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := s.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("error unsubscribing from topic: %w", err)
	}

	return nil
}

// ConvertDataToStringMap converts arbitrary data to string map for FCM
func ConvertDataToStringMap(data any) (map[string]string, error) {
	if data == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %w", err)
	}

	stringMap := make(map[string]string)
	for k, v := range result {
		stringMap[k] = fmt.Sprintf("%v", v)
	}

	return stringMap, nil
}
