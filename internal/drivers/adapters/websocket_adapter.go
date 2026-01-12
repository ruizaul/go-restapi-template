package adapters

import (
	"tacoshare-delivery-api/internal/drivers/services"
	wsModels "tacoshare-delivery-api/internal/websockets/models"
	"tacoshare-delivery-api/internal/websockets/models/ws"
	wsServices "tacoshare-delivery-api/internal/websockets/services"

	"github.com/google/uuid"
)

// WebSocketHubAdapter adapts the WebSocket Hub to the minimal interface needed by LocationService
type WebSocketHubAdapter struct {
	hub *wsServices.Hub
}

// NewWebSocketHubAdapter creates a new WebSocket hub adapter
func NewWebSocketHubAdapter(hub *wsServices.Hub) *WebSocketHubAdapter {
	return &WebSocketHubAdapter{hub: hub}
}

// SendToUser sends a message to a specific user (adapter implementation)
func (a *WebSocketHubAdapter) SendToUser(userID uuid.UUID, message *ws.WSMessage) error {
	// Convert the simple ws.WSMessage to the full wsModels.WSMessage
	fullMessage, err := wsModels.NewWSMessage(wsModels.MessageType(message.Type), message.Data)
	if err != nil {
		return err
	}

	return a.hub.SendToUser(userID, fullMessage)
}

// Compile-time check to ensure WebSocketHubAdapter implements services.WebSocketHub
var _ services.WebSocketHub = (*WebSocketHubAdapter)(nil)
