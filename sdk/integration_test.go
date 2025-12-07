package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Test with a real WebSocket server (integration test)
func TestSDKIntegration(t *testing.T) {
	// Create a test WebSocket server
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token is passed
		token := r.URL.Query().Get("token")
		if token == "" {
			t.Error("Token not received")
		}

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Echo back with "from" field added
		var received Message
		json.Unmarshal(msg, &received)
		response := Message{
			From:    "server",
			To:      received.To,
			Content: "Echo: " + received.Content,
			Type:    "direct",
		}
		data, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, data)

		// Keep connection open for a bit
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create SDK client
	client, err := NewClient(Config{
		APIKey:   "test-token",
		Endpoint: wsURL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Track received messages
	received := make(chan Message, 1)
	client.OnMessage(func(msg Message) {
		received <- msg
	})

	// Connect
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Send message
	err = client.SendMessage("user123", "Hello SDK!")
	if err != nil {
		t.Fatalf("Failed to send: %v", err)
	}

	// Wait for echo response
	select {
	case msg := <-received:
		if msg.Content != "Echo: Hello SDK!" {
			t.Errorf("Expected 'Echo: Hello SDK!', got '%s'", msg.Content)
		}
		if msg.From != "server" {
			t.Errorf("Expected from 'server', got '%s'", msg.From)
		}
		t.Logf("Received: %+v", msg)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for response")
	}

	// Close
	client.Close()
	t.Log("SDK integration test passed!")
}
