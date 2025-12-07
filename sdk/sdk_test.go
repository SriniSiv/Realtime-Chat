package sdk

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// MOCK TRANSPORT (for testing - demonstrates Strategy Pattern)

type MockTransport struct {
	mu        sync.Mutex
	connected bool
	sent      [][]byte
	recvChan  chan []byte
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		sent:     make([][]byte, 0),
		recvChan: make(chan []byte, 10),
	}
}

func (t *MockTransport) Connect(endpoint, token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = true
	return nil
}

func (t *MockTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.connected {
		return ErrNotConnected
	}
	t.sent = append(t.sent, data)
	return nil
}

func (t *MockTransport) Receive() ([]byte, error) {
	select {
	case data := <-t.recvChan:
		return data, nil
	case <-time.After(50 * time.Millisecond):
		return nil, nil
	}
}

func (t *MockTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = false
	return nil
}

func (t *MockTransport) SimulateMessage(msg Message) {
	data, _ := json.Marshal(msg)
	t.recvChan <- data
}

// =============================================================================
// TESTS
// =============================================================================

func TestNewClient(t *testing.T) {
	// Valid config
	client, err := NewClient(Config{APIKey: "key", Endpoint: "ws://localhost/ws"})
	if err != nil || client == nil {
		t.Error("Expected client, got error")
	}

	// Missing APIKey
	_, err = NewClient(Config{Endpoint: "ws://localhost/ws"})
	if err != ErrMissingAPIKey {
		t.Error("Expected ErrMissingAPIKey")
	}

	// Missing Endpoint
	_, err = NewClient(Config{APIKey: "key"})
	if err != ErrMissingEndpoint {
		t.Error("Expected ErrMissingEndpoint")
	}
}

func TestSendMessage(t *testing.T) {
	client, _ := NewClient(Config{APIKey: "key", Endpoint: "ws://localhost/ws"})
	mock := NewMockTransport()
	client.SetTransport(mock)

	// Connect and send
	client.Connect()
	err := client.SendMessage("user1", "Hello!")
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}

	// Verify sent data
	if len(mock.sent) != 1 {
		t.Error("Expected 1 message sent")
	}

	var msg map[string]string
	json.Unmarshal(mock.sent[0], &msg)
	if msg["to"] != "user1" || msg["content"] != "Hello!" {
		t.Error("Message content mismatch")
	}
}

func TestOnMessage(t *testing.T) {
	client, _ := NewClient(Config{APIKey: "key", Endpoint: "ws://localhost/ws"})
	mock := NewMockTransport()
	client.SetTransport(mock)

	received := make(chan Message, 1)
	client.OnMessage(func(msg Message) {
		received <- msg
	})

	client.Connect()
	mock.SimulateMessage(Message{From: "user1", Content: "Hi!", Type: "direct"})

	select {
	case msg := <-received:
		if msg.Content != "Hi!" {
			t.Error("Message content mismatch")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestClose(t *testing.T) {
	client, _ := NewClient(Config{APIKey: "key", Endpoint: "ws://localhost/ws"})
	mock := NewMockTransport()
	client.SetTransport(mock)

	client.Connect()
	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
