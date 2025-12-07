// Package sdk provides a simple Golang SDK for Realtime Messaging API.
//
// Example:
//
//	client, _ := sdk.NewClient(sdk.Config{
//	    APIKey:   "your-api-key",
//	    Endpoint: "ws://localhost:8080/ws",
//	})
//	client.OnMessage(func(msg sdk.Message) { fmt.Println(msg.Content) })
//	client.Connect()
//	client.SendMessage("user-id", "Hello!")
//	client.Close()
package sdk

import (
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ============================================================================
// Errors
// ============================================================================

var (
	ErrNotConnected    = errors.New("not connected to server")
	ErrMissingAPIKey   = errors.New("API key is required")
	ErrMissingEndpoint = errors.New("endpoint URL is required")
)

// ============================================================================
// Config - SDK configuration
// ============================================================================

type Config struct {
	APIKey   string // JWT token for authentication
	Endpoint string // WebSocket URL (e.g., ws://localhost:8080/ws)
}

// ============================================================================
// Message - Chat message structure
// ============================================================================

type Message struct {
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
	GroupID string    `json:"group_id,omitempty"`
	Content string    `json:"content"`
	Type    string    `json:"type"`
	Time    time.Time `json:"timestamp,omitempty"`
}

// ============================================================================
// Transport - Strategy Pattern for connection handling
// ============================================================================

// Transport interface allows swapping connection implementations (e.g., for testing)
type Transport interface {
	Connect(endpoint, token string) error
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}

// wsTransport implements Transport using WebSocket
type wsTransport struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (t *wsTransport) Connect(endpoint, token string) error {
	// Add token to URL query
	u, _ := url.Parse(endpoint)
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	// Connect via WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.conn = conn
	t.mu.Unlock()
	return nil
}

func (t *wsTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return ErrNotConnected
	}
	return t.conn.WriteMessage(websocket.TextMessage, data)
}

func (t *wsTransport) Receive() ([]byte, error) {
	_, data, err := t.conn.ReadMessage()
	return data, err
}

func (t *wsTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return nil
	}

	// Send close message before disconnecting
	t.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return t.conn.Close()
}

// ============================================================================
// Client - Main SDK client
// ============================================================================

// MessageHandler is called when a message is received (Observer Pattern)
type MessageHandler func(msg Message)

// Client provides methods to interact with the Realtime Messaging API
type Client struct {
	config    Config
	transport Transport
	handler   MessageHandler
	done      chan struct{}
	mu        sync.RWMutex
}

// NewClient creates a new SDK client
func NewClient(config Config) (*Client, error) {
	// Validate config
	if config.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	if config.Endpoint == "" {
		return nil, ErrMissingEndpoint
	}

	return &Client{
		config:    config,
		transport: &wsTransport{},
		done:      make(chan struct{}),
	}, nil
}

// Connect establishes WebSocket connection to the server
func (c *Client) Connect() error {
	err := c.transport.Connect(c.config.Endpoint, c.config.APIKey)
	if err != nil {
		return err
	}

	// Start listening for messages in background
	go c.listenForMessages()
	return nil
}

// OnMessage registers a handler for incoming messages (Observer Pattern)
func (c *Client) OnMessage(handler MessageHandler) {
	c.mu.Lock()
	c.handler = handler
	c.mu.Unlock()
}

// SendMessage sends a direct message to a user
func (c *Client) SendMessage(to, content string) error {
	msg := Message{
		To:      to,
		Content: content,
		Type:    "direct",
	}
	data, _ := json.Marshal(msg)
	return c.transport.Send(data)
}

// SendGroupMessage sends a message to a group
func (c *Client) SendGroupMessage(groupID, content string) error {
	msg := Message{
		GroupID: groupID,
		Content: content,
		Type:    "group",
	}
	data, _ := json.Marshal(msg)
	return c.transport.Send(data)
}

// Close gracefully disconnects from the server
func (c *Client) Close() error {
	close(c.done)
	return c.transport.Close()
}

// SetTransport sets a custom transport (useful for testing with mock)
func (c *Client) SetTransport(t Transport) {
	c.transport = t
}

// ============================================================================
// Internal methods
// ============================================================================

// listenForMessages runs in background and handles incoming messages
func (c *Client) listenForMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			// Read message from server
			data, err := c.transport.Receive()
			if err != nil {
				return
			}

			// Parse and dispatch to handler
			c.dispatchMessage(data)
		}
	}
}

// dispatchMessage parses message and calls the handler
func (c *Client) dispatchMessage(data []byte) {
	c.mu.RLock()
	handler := c.handler
	c.mu.RUnlock()

	if handler == nil {
		return
	}

	var msg Message
	if json.Unmarshal(data, &msg) == nil {
		go handler(msg) // Non-blocking call
	}
}
