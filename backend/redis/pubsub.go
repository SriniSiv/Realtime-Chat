package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ChatMessage represents a message for Redis Pub/Sub
type ChatMessage struct {
	From      uuid.UUID `json:"from"`
	FromEmail string    `json:"from_email"`
	To        uuid.UUID `json:"to,omitempty"`
	ToEmail   string    `json:"to_email,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SourceID  string    `json:"source_id"` // Instance ID that published the message
}

// MessageHandler is called when a message is received from Redis
type MessageHandler func(msg *ChatMessage)

// PubSub handles Redis Pub/Sub operations
type PubSub struct {
	client   *redis.Client
	channel  string
	sourceID string
	handler  MessageHandler
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewPubSub creates a new Redis Pub/Sub client
func NewPubSub(addr, password string, db int, channel, sourceID string, handler MessageHandler) (*PubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	pubsubCtx, cancel := context.WithCancel(ctx)

	log.Printf("Redis Pub/Sub connected: addr=%s, channel=%s, sourceID=%s", addr, channel, sourceID)

	return &PubSub{
		client:   client,
		channel:  channel,
		sourceID: sourceID,
		handler:  handler,
		ctx:      pubsubCtx,
		cancel:   cancel,
	}, nil
}

// Publish sends a message to the Redis channel
func (p *PubSub) Publish(ctx context.Context, msg *ChatMessage) error {
	msg.SourceID = p.sourceID

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = p.client.Publish(ctx, p.channel, data).Err()
	if err != nil {
		log.Printf("Failed to publish to Redis: %v", err)
		return err
	}

	log.Printf("Message published to Redis: from=%s, to=%s, type=%s", msg.FromEmail, msg.ToEmail, msg.Type)
	return nil
}

// Subscribe starts listening for messages from Redis
func (p *PubSub) Subscribe() {
	go p.subscribeLoop()
}

// subscribeLoop continuously listens for messages
func (p *PubSub) subscribeLoop() {
	pubsub := p.client.Subscribe(p.ctx, p.channel)
	defer pubsub.Close()

	log.Printf("Redis Pub/Sub subscribed to channel: %s", p.channel)

	ch := pubsub.Channel()

	for {
		select {
		case <-p.ctx.Done():
			log.Println("Redis Pub/Sub subscriber stopped")
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var chatMsg ChatMessage
			if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
				log.Printf("Error unmarshaling Redis message: %v", err)
				continue
			}

			// Skip messages from this same instance (already handled locally)
			if chatMsg.SourceID == p.sourceID {
				continue
			}

			log.Printf("Received message from Redis: from=%s, to=%s, type=%s, sourceID=%s",
				chatMsg.FromEmail, chatMsg.ToEmail, chatMsg.Type, chatMsg.SourceID)

			if p.handler != nil {
				p.handler(&chatMsg)
			}
		}
	}
}

// Close closes the Redis connection
func (p *PubSub) Close() error {
	p.cancel()
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

