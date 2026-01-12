package services

import (
	"encoding/json"
	"sync"

	"tacoshare-delivery-api/internal/websockets/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   uuid.UUID
	Role     string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	Channels map[string]bool // Channels this client is subscribed to
	mu       sync.RWMutex
}

// Hub maintains active WebSocket connections and handles message broadcasting
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients by user ID for direct messaging
	clientsByUser map[uuid.UUID][]*Client

	// Channel subscriptions (channel_name -> clients)
	channels map[string]map[*Client]bool

	// Register requests from clients (exported for handlers)
	Register chan *Client

	// Unregister requests from clients (exported for handlers)
	Unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan []byte

	// Broadcast to specific channel
	channelBroadcast chan *ChannelMessage

	// Send to specific user
	userMessage chan *UserMessage

	mu sync.RWMutex
}

// ChannelMessage represents a message to broadcast to a channel
type ChannelMessage struct {
	Channel string
	Message []byte
}

// UserMessage represents a message to send to a specific user
type UserMessage struct {
	UserID  uuid.UUID
	Message []byte
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		clientsByUser:    make(map[uuid.UUID][]*Client),
		channels:         make(map[string]map[*Client]bool),
		Register:         make(chan *Client),
		Unregister:       make(chan *Client),
		broadcast:        make(chan []byte, 256),
		channelBroadcast: make(chan *ChannelMessage, 256),
		userMessage:      make(chan *UserMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case channelMsg := <-h.channelBroadcast:
			h.broadcastToChannel(channelMsg.Channel, channelMsg.Message)

		case userMsg := <-h.userMessage:
			h.sendToUser(userMsg.UserID, userMsg.Message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to user map
	h.clientsByUser[client.UserID] = append(h.clientsByUser[client.UserID], client)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove from clients map
		delete(h.clients, client)

		// Remove from user map
		userClients := h.clientsByUser[client.UserID]
		for i, c := range userClients {
			if c == client {
				h.clientsByUser[client.UserID] = append(userClients[:i], userClients[i+1:]...)
				break
			}
		}
		if len(h.clientsByUser[client.UserID]) == 0 {
			delete(h.clientsByUser, client.UserID)
		}

		// Remove from all channels
		client.mu.RLock()
		for channel := range client.Channels {
			if clients, ok := h.channels[channel]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.channels, channel)
				}
			}
		}
		client.mu.RUnlock()

		close(client.Send)
	}
}

// broadcastToAll broadcasts a message to all connected clients
func (h *Hub) broadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send buffer is full, close connection
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// broadcastToChannel broadcasts a message to all clients subscribed to a channel
func (h *Hub) broadcastToChannel(channel string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.channels[channel]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// sendToUser sends a message to all connections of a specific user
func (h *Hub) sendToUser(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clientsByUser[userID]
	if !ok {
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// SubscribeToChannel subscribes a client to a channel
func (h *Hub) SubscribeToChannel(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]bool)
	}

	h.channels[channel][client] = true

	client.mu.Lock()
	client.Channels[channel] = true
	client.mu.Unlock()
}

// UnsubscribeFromChannel unsubscribes a client from a channel
func (h *Hub) UnsubscribeFromChannel(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[channel]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.channels, channel)
		}
	}

	client.mu.Lock()
	delete(client.Channels, channel)
	client.mu.Unlock()
}

// BroadcastToAll broadcasts a message to all clients
func (h *Hub) BroadcastToAll(message *models.WSMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

// BroadcastToChannel broadcasts a message to a specific channel
func (h *Hub) BroadcastToChannel(channel string, message *models.WSMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.channelBroadcast <- &ChannelMessage{
		Channel: channel,
		Message: data,
	}
	return nil
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uuid.UUID, message *models.WSMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.userMessage <- &UserMessage{
		UserID:  userID,
		Message: data,
	}
	return nil
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetChannelClientCount returns the number of clients in a channel
func (h *Hub) GetChannelClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.channels[channel]; ok {
		return len(clients)
	}
	return 0
}

// GetUserClientCount returns the number of connections for a user
func (h *Hub) GetUserClientCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clientsByUser[userID])
}
