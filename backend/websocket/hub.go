package websocket

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

// Client represents a WebSocket client
type Client struct {
	ID     uuid.UUID
	Email  string
	Conn   *Connection
	Send   chan []byte
	Hub    *Hub
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
	From    uuid.UUID `json:"from"`
	To      uuid.UUID `json:"to,omitempty"`
	Content string    `json:"content"`
	Type    string    `json:"type"` // "broadcast", "direct", "system"
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

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("Client unregistered: %s (%s)", client.ID, client.Email)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if message.Type == "direct" && message.To != uuid.Nil {
				// Send to specific client
				if client, ok := h.clients[message.To]; ok {
					select {
					case client.Send <- []byte(message.Content):
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			} else {
				// Broadcast to all clients
				for id, client := range h.clients {
					select {
					case client.Send <- []byte(message.Content):
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

