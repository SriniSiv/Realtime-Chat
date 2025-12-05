package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Client represents a WebSocket client
type Client struct {
	ID    uuid.UUID
	Email string
	Conn  *Connection
	Send  chan []byte
	Hub   *Hub
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[uuid.UUID]*Client

	// Inbound messages from clients
	broadcast chan *Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	From      uuid.UUID `json:"from"`
	FromEmail string    `json:"from_email"`
	To        uuid.UUID `json:"to,omitempty"`
	ToEmail   string    `json:"to_email,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "direct", "broadcast", "system"
	Timestamp time.Time `json:"timestamp"`
}

// IncomingMessage represents a message received from client
type IncomingMessage struct {
	To      string `json:"to"` // Can be user ID or "all" for broadcast
	Content string `json:"content"`
	Type    string `json:"type,omitempty"` // Optional, defaults to "direct"
}

// OnlineUser represents an online user
type OnlineUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (%s)", client.ID, client.Email)

			// Send system message to notify user is online
			h.notifyUserOnline(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("Client unregistered: %s (%s)", client.ID, client.Email)

				// Send system message to notify user is offline
				h.notifyUserOffline(client)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Enrich message with email information
			h.EnrichMessage(message)

			h.mu.RLock()

			// Convert message to JSON
			messageJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				h.mu.RUnlock()
				continue
			}

			if message.Type == "direct" && message.To != uuid.Nil {
				// Send to specific client (direct message)
				if client, ok := h.clients[message.To]; ok {
					select {
					case client.Send <- messageJSON:
						log.Printf("Direct message sent from %s to %s", message.FromEmail, message.ToEmail)
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				} else {
					log.Printf("Recipient %s not found or offline", message.To)
				}
			} else {
				// Broadcast to all clients except sender
				for id, client := range h.clients {
					if id == message.From {
						continue // Don't send message back to sender
					}
					select {
					case client.Send <- messageJSON:
					default:
						close(client.Send)
						delete(h.clients, id)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// GetActiveClients returns the number of active clients
func (h *Hub) GetActiveClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClientByID returns a client by ID
func (h *Hub) GetClientByID(id uuid.UUID) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[id]
	return client, ok
}

// GetOnlineUsers returns a list of all online users
func (h *Hub) GetOnlineUsers() []OnlineUser {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]OnlineUser, 0, len(h.clients))
	for _, client := range h.clients {
		users = append(users, OnlineUser{
			ID:    client.ID,
			Email: client.Email,
		})
	}
	return users
}

// ParseIncomingMessage parses a raw message from client
func ParseIncomingMessage(rawMessage []byte, senderID uuid.UUID) (*Message, error) {
	var incoming IncomingMessage
	if err := json.Unmarshal(rawMessage, &incoming); err != nil {
		// If not JSON, treat as plain text broadcast
		return &Message{
			From:      senderID,
			Content:   string(rawMessage),
			Type:      "broadcast",
			Timestamp: time.Now(),
		}, nil
	}

	message := &Message{
		From:      senderID,
		Content:   incoming.Content,
		Timestamp: time.Now(),
	}

	// Determine message type
	if incoming.Type != "" {
		message.Type = incoming.Type
	} else if incoming.To == "" || incoming.To == "all" {
		message.Type = "broadcast"
	} else {
		message.Type = "direct"
		// Parse recipient ID
		recipientID, err := uuid.Parse(incoming.To)
		if err != nil {
			return nil, err
		}
		message.To = recipientID
	}

	return message, nil
}

// EnrichMessage adds sender and recipient email information to the message
func (h *Hub) EnrichMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Add sender email
	if sender, ok := h.clients[message.From]; ok {
		message.FromEmail = sender.Email
	}

	// Add recipient email for direct messages
	if message.Type == "direct" && message.To != uuid.Nil {
		if recipient, ok := h.clients[message.To]; ok {
			message.ToEmail = recipient.Email
		}
	}
}

// notifyUserOnline sends a system message when a user comes online
func (h *Hub) notifyUserOnline(client *Client) {
	systemMsg := &Message{
		From:      client.ID,
		FromEmail: client.Email,
		Content:   client.Email + " is now online",
		Type:      "system",
		Timestamp: time.Now(),
	}

	messageJSON, _ := json.Marshal(systemMsg)

	// Broadcast to all other clients
	for id, c := range h.clients {
		if id != client.ID {
			select {
			case c.Send <- messageJSON:
			default:
			}
		}
	}
}

// notifyUserOffline sends a system message when a user goes offline
func (h *Hub) notifyUserOffline(client *Client) {
	systemMsg := &Message{
		From:      client.ID,
		FromEmail: client.Email,
		Content:   client.Email + " is now offline",
		Type:      "system",
		Timestamp: time.Now(),
	}

	messageJSON, _ := json.Marshal(systemMsg)

	// Broadcast to all remaining clients
	for id, c := range h.clients {
		if id != client.ID {
			select {
			case c.Send <- messageJSON:
			default:
			}
		}
	}
}
