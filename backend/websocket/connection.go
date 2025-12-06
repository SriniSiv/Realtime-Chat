package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Connection wraps the WebSocket connection
type Connection struct {
	ws *websocket.Conn
}

// NewConnection creates a new Connection
func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{ws: ws}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.ws.Close()
	}()

	c.Conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.ws.SetPongHandler(func(string) error {
		c.Conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.Conn.ws.SetReadLimit(maxMessageSize)
	log.Printf("ReadPump started for client: %s (%s)", c.ID, c.Email)

	for {
		_, rawMessage, err := c.Conn.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for %s: %v", c.Email, err)
			}
			log.Printf("ReadPump ending for client: %s (%s), error: %v", c.ID, c.Email, err)
			break
		}

		log.Printf("ReadPump received raw message from %s: %s", c.Email, string(rawMessage))

		// Parse the incoming message
		message, err := ParseIncomingMessage(rawMessage, c.ID)
		if err != nil {
			log.Printf("Error parsing message from %s: %v", c.Email, err)
			continue
		}

		log.Printf("Sending message to broadcast channel from %s", c.Email)
		// Send message to hub for routing
		c.Hub.broadcast <- message
		log.Printf("Message sent to broadcast channel from %s", c.Email)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
