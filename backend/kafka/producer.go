package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// ChatMessage represents a message to be published to Kafka
type ChatMessage struct {
	From      uuid.UUID `json:"from"`
	FromEmail string    `json:"from_email"`
	To        uuid.UUID `json:"to,omitempty"`
	ToEmail   string    `json:"to_email,omitempty"`
	GroupID   uuid.UUID `json:"group_id,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SourceID  string    `json:"source_id"` // Instance ID that produced the message
}

// Producer handles publishing messages to Kafka
type Producer struct {
	writer   *kafka.Writer
	topic    string
	sourceID string
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string, sourceID string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,                      // Send immediately for real-time
		BatchTimeout: 10 * time.Millisecond,  // Low latency
		RequiredAcks: kafka.RequireOne,       // Wait for leader ack
		Async:        false,                  // Synchronous for reliability
	}

	log.Printf("Kafka producer initialized: brokers=%v, topic=%s", brokers, topic)

	return &Producer{
		writer:   writer,
		topic:    topic,
		sourceID: sourceID,
	}, nil
}

// Publish sends a message to Kafka
func (p *Producer) Publish(ctx context.Context, msg *ChatMessage) error {
	// Add source ID to track which instance produced the message
	msg.SourceID = p.sourceID

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Use sender ID as key for ordering per-user messages
	key := msg.From.String()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})

	if err != nil {
		log.Printf("Failed to publish message to Kafka: %v", err)
		return err
	}

	log.Printf("Message published to Kafka: from=%s, to=%s, type=%s", msg.FromEmail, msg.ToEmail, msg.Type)
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

