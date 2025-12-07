package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// MessageHandler is called when a message is received from Kafka
// Used for persistence - saves message to database
type MessageHandler func(msg *ChatMessage)

// Consumer handles consuming messages from Kafka for persistence
// Uses a single consumer group so each message is processed exactly once
type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewConsumer creates a new Kafka consumer for message persistence
// groupID should be the SAME across all instances for exactly-once processing
func NewConsumer(brokers []string, topic string, groupID string, handler MessageHandler) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID, // Same group = messages distributed across instances
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        100 * time.Millisecond,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Kafka persistence consumer initialized: brokers=%v, topic=%s, groupID=%s", brokers, topic, groupID)

	return &Consumer{
		reader:  reader,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Start begins consuming messages in a goroutine
func (c *Consumer) Start() {
	go c.consumeLoop()
}

// consumeLoop continuously reads messages from Kafka for persistence
func (c *Consumer) consumeLoop() {
	log.Println("Kafka persistence consumer started")

	for {
		select {
		case <-c.ctx.Done():
			log.Println("Kafka persistence consumer stopped")
			return
		default:
			msg, err := c.reader.ReadMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					return // Context cancelled
				}
				log.Printf("Error reading from Kafka: %v", err)
				time.Sleep(time.Second) // Back off on error
				continue
			}

			var chatMsg ChatMessage
			if err := json.Unmarshal(msg.Value, &chatMsg); err != nil {
				log.Printf("Error unmarshaling Kafka message: %v", err)
				continue
			}

			log.Printf("Kafka persistence: processing message from=%s, to=%s, type=%s",
				chatMsg.FromEmail, chatMsg.ToEmail, chatMsg.Type)

			// Call the handler to save to database
			// Only one instance in the consumer group will process each message
			if c.handler != nil {
				c.handler(&chatMsg)
			}
		}
	}
}

// Close stops the consumer and closes the reader
func (c *Consumer) Close() error {
	c.cancel()
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

