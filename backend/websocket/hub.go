package websocket

import (
	"backend/kafka"
	"backend/models"
	"backend/redis"
	"backend/service"
	"context"
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

// GroupMemberProvider interface for getting group members
type GroupMemberProvider interface {
	GetGroupMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error)
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

	// Message service for persisting messages (used when Kafka is disabled)
	messageService *service.MessageService

	// Group member provider for getting group members
	groupMemberProvider GroupMemberProvider

	// Kafka producer for publishing messages for persistence
	kafkaProducer *kafka.Producer

	// Redis Pub/Sub for real-time message fanout between instances
	redisPubSub *redis.PubSub

	// Instance ID for identifying this server
	instanceID string

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	From      uuid.UUID `json:"from"`
	FromEmail string    `json:"from_email"`
	To        uuid.UUID `json:"to,omitempty"`
	ToEmail   string    `json:"to_email,omitempty"`
	GroupID   uuid.UUID `json:"group_id,omitempty"`
	GroupName string    `json:"group_name,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "direct", "broadcast", "group", "system"
	Timestamp time.Time `json:"timestamp"`
}

// IncomingMessage represents a message received from client
type IncomingMessage struct {
	To      string `json:"to"`                 // Can be user ID or "all" for broadcast
	GroupID string `json:"group_id,omitempty"` // Group ID for group messages
	Content string `json:"content"`
	Type    string `json:"type,omitempty"` // Optional, defaults to "direct"
}

// OnlineUser represents an online user
type OnlineUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// NewHub creates a new Hub instance
func NewHub(messageService *service.MessageService) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID]*Client),
		broadcast:      make(chan *Message, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		messageService: messageService,
	}
}

// SetKafkaProducer sets the Kafka producer for message persistence
func (h *Hub) SetKafkaProducer(producer *kafka.Producer, instanceID string) {
	h.kafkaProducer = producer
	h.instanceID = instanceID
	log.Printf("Kafka producer set for hub (persistence), instanceID=%s", instanceID)
}

// SetRedisPubSub sets the Redis Pub/Sub for real-time fanout
func (h *Hub) SetRedisPubSub(pubsub *redis.PubSub) {
	h.redisPubSub = pubsub
	log.Println("Redis Pub/Sub set for hub (real-time fanout)")
}

// SetGroupMemberProvider sets the group member provider for group messages
func (h *Hub) SetGroupMemberProvider(provider GroupMemberProvider) {
	h.groupMemberProvider = provider
	log.Println("Group member provider set for hub")
}

// HandleRedisMessage processes a message received from Redis (from another instance)
func (h *Hub) HandleRedisMessage(msg *redis.ChatMessage) {
	// Convert Redis message to internal Message format
	message := &Message{
		From:      msg.From,
		FromEmail: msg.FromEmail,
		To:        msg.To,
		ToEmail:   msg.ToEmail,
		Content:   msg.Content,
		Type:      msg.Type,
		Timestamp: msg.Timestamp,
	}

	log.Printf("Processing Redis message: from=%s, to=%s, type=%s", message.FromEmail, message.ToEmail, message.Type)

	// Broadcast to local clients only (message already saved to DB via Kafka)
	h.broadcastToLocalClients(message)
}

// broadcastToLocalClients sends a message to clients connected to this instance
func (h *Hub) broadcastToLocalClients(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if message.Type == "direct" && message.To != uuid.Nil {
		// Send to recipient only (sender already has the message via optimistic update)
		if client, ok := h.clients[message.To]; ok {
			select {
			case client.Send <- messageJSON:
				log.Printf("Direct message sent to local client %s", message.ToEmail)
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	} else if message.Type == "group" && message.GroupID != uuid.Nil {
		// Send to all group members except sender (sender already has the message via optimistic update)
		if h.groupMemberProvider != nil {
			memberIDs, err := h.groupMemberProvider.GetGroupMemberIDs(message.GroupID)
			if err != nil {
				log.Printf("Error getting group members: %v", err)
				return
			}
			for _, memberID := range memberIDs {
				// Skip sender - they already have the message
				if memberID == message.From {
					continue
				}
				if client, ok := h.clients[memberID]; ok {
					select {
					case client.Send <- messageJSON:
						log.Printf("Group message sent to member %s", memberID)
					default:
						close(client.Send)
						delete(h.clients, memberID)
					}
				}
			}
		}
	} else {
		// Broadcast to all local clients except sender
		for id, client := range h.clients {
			if id == message.From {
				continue
			}
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(h.clients, id)
			}
		}
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
			log.Printf("Received message in broadcast channel: from=%s, to=%s, type=%s, content=%s", message.From, message.To, message.Type, message.Content)

			// Enrich message with email information
			h.EnrichMessage(message)

			// Handle message persistence
			if message.Type != "system" {
				log.Printf("Persisting message: type=%s, from=%s, to=%s, groupID=%s, content=%s",
					message.Type, message.From, message.To, message.GroupID, message.Content)
				if h.kafkaProducer != nil {
					log.Printf("Using Kafka for persistence")
					// Kafka enabled: publish to Kafka for persistence
					go func(msg *Message) {
						kafkaMsg := &kafka.ChatMessage{
							From:      msg.From,
							FromEmail: msg.FromEmail,
							To:        msg.To,
							ToEmail:   msg.ToEmail,
							GroupID:   msg.GroupID,
							Content:   msg.Content,
							Type:      msg.Type,
							Timestamp: msg.Timestamp,
						}
						if err := h.kafkaProducer.Publish(context.Background(), kafkaMsg); err != nil {
							log.Printf("Error publishing to Kafka: %v", err)
						}
					}(message)
				} else if h.messageService != nil {
					log.Printf("Using direct DB save for persistence")
					// No Kafka: save directly to database
					go func(msg *Message) {
						var savedMsg *models.ChatMessage
						var err error
						if msg.Type == "group" && msg.GroupID != uuid.Nil {
							log.Printf("Saving group message: groupID=%s", msg.GroupID)
							savedMsg, err = h.messageService.SaveGroupMessage(msg.From, msg.GroupID, msg.Content)
						} else {
							log.Printf("Saving direct message: from=%s, to=%s", msg.From, msg.To)
							savedMsg, err = h.messageService.SaveMessage(msg.From, msg.To, msg.Content, msg.Type)
						}
						if err != nil {
							log.Printf("Error saving message: %v", err)
						} else {
							log.Printf("Message saved successfully: id=%s, type=%s", savedMsg.ID, msg.Type)
						}
					}(message)
				} else {
					log.Printf("WARNING: No persistence method available (kafka=%v, messageService=%v)",
						h.kafkaProducer != nil, h.messageService != nil)
				}
			}

			// Handle real-time fanout
			if h.redisPubSub != nil {
				// Redis enabled: publish to Redis for fanout to other instances
				go func(msg *Message) {
					redisMsg := &redis.ChatMessage{
						From:      msg.From,
						FromEmail: msg.FromEmail,
						To:        msg.To,
						ToEmail:   msg.ToEmail,
						Content:   msg.Content,
						Type:      msg.Type,
						Timestamp: msg.Timestamp,
					}
					if err := h.redisPubSub.Publish(context.Background(), redisMsg); err != nil {
						log.Printf("Error publishing to Redis: %v", err)
					}
				}(message)
			}

			// Broadcast to local clients
			h.broadcastToLocalClients(message)
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

// GetOnlineUserIDs returns IDs of all online users (implements OnlineUsersProvider interface)
func (h *Hub) GetOnlineUserIDs() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(h.clients))
	for _, client := range h.clients {
		ids = append(ids, client.ID)
	}
	return ids
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

	// Check if this is a group message
	if incoming.GroupID != "" {
		groupID, err := uuid.Parse(incoming.GroupID)
		if err != nil {
			log.Printf("Error parsing group ID '%s': %v", incoming.GroupID, err)
			return nil, err
		}
		message.Type = "group"
		message.GroupID = groupID
		log.Printf("Parsed group message: groupID=%s, content=%s", message.GroupID, message.Content)
		return message, nil
	}

	// Determine message type for non-group messages
	if incoming.To == "" || incoming.To == "all" {
		message.Type = "broadcast"
	} else {
		message.Type = "direct"
		// Parse recipient ID
		recipientID, err := uuid.Parse(incoming.To)
		if err != nil {
			log.Printf("Error parsing recipient ID '%s': %v", incoming.To, err)
			return nil, err
		}
		message.To = recipientID
	}

	// Override type if explicitly provided (but still keep the parsed To)
	if incoming.Type != "" && incoming.Type != "direct" && incoming.Type != "group" {
		message.Type = incoming.Type
	}

	log.Printf("Parsed message: type=%s, to=%s, content=%s", message.Type, message.To, message.Content)

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
