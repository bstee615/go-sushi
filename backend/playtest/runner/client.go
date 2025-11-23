package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientSimulator simulates a WebSocket client
type ClientSimulator struct {
	ID        string
	conn      *websocket.Conn
	messages  []json.RawMessage
	gameState map[string]interface{}
	mu        sync.Mutex
	receiveCh chan json.RawMessage
	errorCh   chan error
}

// NewClientSimulator creates a new client simulator and connects to the server
func NewClientSimulator(id string, serverURL string) (*ClientSimulator, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serverURL, err)
	}

	client := &ClientSimulator{
		ID:        id,
		conn:      conn,
		messages:  make([]json.RawMessage, 0),
		gameState: make(map[string]interface{}),
		receiveCh: make(chan json.RawMessage, 10),
		errorCh:   make(chan error, 1),
	}

	// Start reading messages in background
	go client.readPump()

	return client, nil
}

// readPump continuously reads messages from the WebSocket connection
func (c *ClientSimulator) readPump() {
	defer close(c.receiveCh)
	defer close(c.errorCh)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.errorCh <- fmt.Errorf("websocket error: %w", err)
			}
			return
		}

		c.receiveCh <- json.RawMessage(message)
	}
}

// SendMessage sends a message to the server
func (c *ClientSimulator) SendMessage(msg map[string]interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("[Client %s] Sending: %s", c.ID, string(data))

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// ReceiveMessage waits for and returns the next message from the server
func (c *ClientSimulator) ReceiveMessage(timeout time.Duration) (json.RawMessage, error) {
	select {
	case msg, ok := <-c.receiveCh:
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		// Store the message
		c.messages = append(c.messages, msg)

		// Parse and log the message
		var parsed map[string]interface{}
		if err := json.Unmarshal(msg, &parsed); err == nil {
			msgType, _ := parsed["type"].(string)
			log.Printf("[Client %s] Received: %s", c.ID, msgType)

			// Update game state if this is a game_state message
			if msgType == "game_state" {
				if payload, ok := parsed["payload"].(map[string]interface{}); ok {
					c.gameState = payload
				}
			}
		}

		return msg, nil

	case err := <-c.errorCh:
		return nil, err

	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

// Close closes the WebSocket connection
func (c *ClientSimulator) Close() error {
	return c.conn.Close()
}

// GetMessages returns all messages received by this client
func (c *ClientSimulator) GetMessages() []json.RawMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]json.RawMessage{}, c.messages...)
}

// GetGameState returns the current game state for this client
func (c *ClientSimulator) GetGameState() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.gameState
}
