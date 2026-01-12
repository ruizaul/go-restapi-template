package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"tacoshare-delivery-api/internal/websockets/models"
	"tacoshare-delivery-api/internal/websockets/services"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, validate origin properly
		return true
	},
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// WSHandler handles WebSocket connections
type WSHandler struct {
	hub *services.Hub
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(hub *services.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleConnection handles WebSocket connection requests
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok {
		userRole = "unknown"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create client
	client := &services.Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Role:     userRole,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client
	h.hub.Register <- client

	// Send connection confirmation
	connectedMsg, err := models.NewConnectedMessage(client.ID, userID.String(), userRole)
	if err == nil {
		msgData, _ := json.Marshal(connectedMsg)
		client.Send <- msgData
	}

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// readPump pumps messages from the WebSocket connection to the hub
func (h *WSHandler) readPump(client *services.Client) {
	defer func() {
		h.hub.Unregister <- client
		_ = client.Conn.Close()
	}()

	_ = client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetPongHandler(func(string) error {
		if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse incoming message
		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		// Handle different message types
		h.handleClientMessage(client, &wsMsg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *WSHandler) writePump(client *services.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				// Hub closed the channel
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					return
				}
				if _, err := w.Write(<-client.Send); err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage handles incoming messages from clients
func (h *WSHandler) handleClientMessage(client *services.Client, msg *models.WSMessage) {
	switch msg.Type {
	case models.MessageTypePing:
		// Respond with pong
		pongMsg, err := models.NewWSMessage(models.MessageTypePong, nil)
		if err == nil {
			msgData, _ := json.Marshal(pongMsg)
			client.Send <- msgData
		}

	default:
		// Unhandled message type
	}
}

// HandleOrderChannel handles WebSocket connections for order-specific channels
func (h *WSHandler) HandleOrderChannel(w http.ResponseWriter, r *http.Request) {
	// Get order ID from path
	orderIDStr := r.PathValue("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Get user info from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok {
		userRole = "unknown"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create client
	client := &services.Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Role:     userRole,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client
	h.hub.Register <- client

	// Subscribe to order channel
	channelName := "order:" + orderID.String()
	h.hub.SubscribeToChannel(client, channelName)

	// Send connection confirmation
	connectedMsg, err := models.NewConnectedMessage(client.ID, userID.String(), userRole)
	if err == nil {
		msgData, _ := json.Marshal(connectedMsg)
		client.Send <- msgData
	}

	// Start goroutines
	go h.writePump(client)
	go h.readPump(client)
}

// HandleDriverChannel handles WebSocket connections for driver-specific channels
func (h *WSHandler) HandleDriverChannel(w http.ResponseWriter, r *http.Request) {
	// Get driver ID from path
	driverIDStr := r.PathValue("driver_id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	// Get user info from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user is accessing their own channel or is admin
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if userID != driverID && userRole != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create client
	client := &services.Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Role:     userRole,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client
	h.hub.Register <- client

	// Subscribe to driver channel
	channelName := "driver:" + driverID.String()
	h.hub.SubscribeToChannel(client, channelName)

	// Send connection confirmation
	connectedMsg, err := models.NewConnectedMessage(client.ID, userID.String(), userRole)
	if err == nil {
		msgData, _ := json.Marshal(connectedMsg)
		client.Send <- msgData
	}

	// Start goroutines
	go h.writePump(client)
	go h.readPump(client)
}
