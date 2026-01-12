package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WSMessage represents a WebSocket message (simplified for external use)
type WSMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	MessageID string                 `json:"message_id"`
}

// NewWSMessage creates a new WebSocket message
func NewWSMessage(msgType string, data map[string]interface{}) *WSMessage {
	return &WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}
}

// ToJSON converts the message to JSON bytes
func (m *WSMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
