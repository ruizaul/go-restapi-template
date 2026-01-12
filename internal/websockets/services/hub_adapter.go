package services

import (
	"encoding/json"

	"tacoshare-delivery-api/internal/websockets/models"

	"github.com/google/uuid"
)

// HubAdapter adapts the Hub to work with the assignment service interface
type HubAdapter struct {
	hub *Hub
}

// NewHubAdapter creates a new hub adapter
func NewHubAdapter(hub *Hub) *HubAdapter {
	return &HubAdapter{hub: hub}
}

// BroadcastToChannel broadcasts a message to a specific channel
func (a *HubAdapter) BroadcastToChannel(channel string, message any) error {
	// Convert message to WSMessage
	var wsMsg *models.WSMessage

	// Check if message is already a WSMessage
	if msg, ok := message.(*models.WSMessage); ok {
		wsMsg = msg
	} else {
		// Convert generic message to WSMessage
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}

		// Determine message type from data
		var msgData map[string]any
		if err := json.Unmarshal(data, &msgData); err != nil {
			return err
		}

		msgType := models.MessageTypeNewOrder
		if typeStr, ok := msgData["type"].(string); ok {
			msgType = models.MessageType(typeStr)
		}

		wsMsg, err = models.NewWSMessage(msgType, message)
		if err != nil {
			return err
		}

		// Extract timeout_seconds if present
		if timeoutSeconds, ok := msgData["timeout_seconds"].(float64); ok {
			timeoutInt := int(timeoutSeconds)
			wsMsg.TimeoutSeconds = &timeoutInt
		}
	}

	return a.hub.BroadcastToChannel(channel, wsMsg)
}

// SendToUser sends a message to a specific user
func (a *HubAdapter) SendToUser(userID uuid.UUID, message any) error {
	// Convert message to WSMessage
	var wsMsg *models.WSMessage

	// Check if message is already a WSMessage
	if msg, ok := message.(*models.WSMessage); ok {
		wsMsg = msg
	} else {
		// Convert generic message to WSMessage
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}

		// Determine message type from data
		var msgData map[string]any
		if err := json.Unmarshal(data, &msgData); err != nil {
			return err
		}

		msgType := models.MessageTypeNewOrder
		if typeStr, ok := msgData["type"].(string); ok {
			msgType = models.MessageType(typeStr)
		}

		wsMsg, err = models.NewWSMessage(msgType, message)
		if err != nil {
			return err
		}

		// Extract timeout_seconds if present
		if timeoutSeconds, ok := msgData["timeout_seconds"].(float64); ok {
			timeoutInt := int(timeoutSeconds)
			wsMsg.TimeoutSeconds = &timeoutInt
		}
	}

	return a.hub.SendToUser(userID, wsMsg)
}
