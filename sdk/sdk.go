// Package sdk provides a Golang SDK for Realtime Messaging/Streaming API.
//
// # Quick Start
//
//	client, _ := sdk.NewClient(sdk.Config{
//	    APIKey:   "your-api-key",
//	    Endpoint: "wss://example.com/realtime",
//	})
//	client.OnMessage(func(msg sdk.Message) {
//	    fmt.Println("Received:", msg.Content)
//	})
//	client.Connect()
//	client.SendMessage("channel", "Hello!")
//	defer client.Close()
//
// # Design Patterns Used
//
//   - Strategy Pattern: Transport interface for testability
//   - Observer Pattern: OnMessage handler for real-time events
package sdk

import (
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ERRORS
var (
	ErrNotConnected    = errors.New("sdk: not connected")
	ErrMissingAPIKey   = errors.New("sdk: missing API key")
	ErrMissingEndpoint = errors.New("sdk: missing endpoint")
)

// TYPES

// Config holds SDK configuration
type Config struct {
	APIKey   string // JWT token for authentication
	Endpoint string // WebSocket URL (e.g., wss://example.com/ws)
}

// Message represents a chat message or stream event
type Message struct {
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
	GroupID string    `json:"group_id,omitempty"`
	Content string    `json:"content"`
	Type    string    `json:"type"` // "direct", "group", "broadcast"
	Time    time.Time `json:"timestamp,omitempty"`
}

// TRANSPORT INTERFACE (STRATEGY PATTERN)

// Transport defines connection strategy (Strategy Pattern)
type Transport interface {
	Connect(endpoint, token string) error
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}

// WebSocketTransport is the default transport implementation
type WebSocketTransport struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (t *WebSocketTransport) Connect(endpoint, token string) error {
	u, _ := url.Parse(endpoint)
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.conn = conn
	t.mu.Unlock()
	return nil
}

func (t *WebSocketTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return ErrNotConnected
	}
	return t.conn.WriteMessage(websocket.TextMessage, data)
}

func (t *WebSocketTransport) Receive() ([]byte, error) {
	_, data, err := t.conn.ReadMessage()
	return data, err
}

func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn != nil {
		t.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return t.conn.Close()
	}
	return nil
}

// CLIENT (OBSERVER PATTERN)

// MessageHandler is callback for incoming messages (Observer Pattern)
type MessageHandler func(msg Message)

// Client is the SDK client
type Client struct {
	config    Config
	transport Transport
	handler   MessageHandler
	done      chan struct{}
	mu        sync.RWMutex
}

// NewClient creates a new SDK client
//
// Usage:
//
//	client, _ := sdk.NewClient(sdk.Config{
//	    APIKey:   "my-api-key",
//	    Endpoint: "wss://example.com/realtime",
//	})
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	if config.Endpoint == "" {
		return nil, ErrMissingEndpoint
	}
	return &Client{
		config:    config,
		transport: &WebSocketTransport{},
		done:      make(chan struct{}),
	}, nil
}

// Connect establishes connection to the server
func (c *Client) Connect() error {
	if err := c.transport.Connect(c.config.Endpoint, c.config.APIKey); err != nil {
		return err
	}
	go c.readLoop()
	return nil
}

// OnMessage registers a message handler (Observer Pattern)
func (c *Client) OnMessage(handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handler = handler
}

// SendMessage sends a message to a channel/user
//
// Usage:
//
//	err := sdk.SendMessage("general", "Hello, world!")
func (c *Client) SendMessage(to, content string) error {
	msg := map[string]string{"to": to, "content": content, "type": "direct"}
	data, _ := json.Marshal(msg)
	return c.transport.Send(data)
}

// Close disconnects gracefully
//
// Usage:
//
//	sdk.Close()
func (c *Client) Close() error {
	close(c.done)
	return c.transport.Close()
}

// readLoop handles incoming messages (concurrency)
func (c *Client) readLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
			data, err := c.transport.Receive()
			if err != nil {
				return
			}
			c.mu.RLock()
			h := c.handler
			c.mu.RUnlock()
			if h != nil {
				var msg Message
				if json.Unmarshal(data, &msg) == nil {
					go h(msg) // Non-blocking handler call
				}
			}
		}
	}
}

// SetTransport allows setting custom transport (for testing)
func (c *Client) SetTransport(t Transport) {
	c.transport = t
}
