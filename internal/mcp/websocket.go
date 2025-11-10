// Package mcp provides WebSocket support for real-time MCP server communication.
// This enables event-driven updates for agent status changes and new messages,
// reducing polling overhead and improving responsiveness.
//
// Example usage:
//
//	wsClient := mcp.NewWebSocketClient("ws://localhost:8765/ws")
//	eventChan := wsClient.Events()
//	
//	if err := wsClient.Connect(); err != nil {
//	    log.Fatal(err)
//	}
//	defer wsClient.Close()
//	
//	for event := range eventChan {
//	    switch event.Type {
//	    case mcp.EventAgentStatus:
//	        fmt.Printf("Agent %s status changed\n", event.AgentStatus.Name)
//	    case mcp.EventNewMessage:
//	        fmt.Printf("New message: %s\n", event.Message.Content)
//	    }
//	}
package mcp

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EventType represents the type of WebSocket event received from the MCP server.
type EventType string

const (
	EventAgentStatus EventType = "agent_status"
	EventNewMessage  EventType = "new_message"
	EventError       EventType = "error"
	EventConnected   EventType = "connected"
	EventDisconnected EventType = "disconnected"
)

// Event represents a WebSocket event from the MCP server.
type Event struct {
	Type        EventType    `json:"type"`
	AgentStatus *AgentStatus `json:"agent_status,omitempty"`
	Message     *Message     `json:"message,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// WebSocketClient manages a WebSocket connection to the MCP server
// with automatic reconnection and event distribution.
type WebSocketClient struct {
	url            string
	conn           *websocket.Conn
	connMutex      sync.RWMutex
	events         chan Event
	done           chan struct{}
	reconnectDelay time.Duration
	maxReconnectDelay time.Duration
	connected      bool
	connectedMutex sync.RWMutex
}

// NewWebSocketClient creates a new WebSocket client for the MCP server.
// The client starts disconnected and must be connected via Connect().
//
// Example:
//
//	client := mcp.NewWebSocketClient("ws://localhost:8765/ws")
func NewWebSocketClient(url string) *WebSocketClient {
	return &WebSocketClient{
		url:               url,
		events:            make(chan Event, 100), // Buffer events to prevent blocking
		done:              make(chan struct{}),
		reconnectDelay:    1 * time.Second,
		maxReconnectDelay: 30 * time.Second,
		connected:         false,
	}
}

// Connect establishes a WebSocket connection to the MCP server.
// It starts background goroutines for reading messages and handling reconnection.
// Returns an error if the initial connection fails.
func (c *WebSocketClient) Connect() error {
	if err := c.connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Start message reader
	go c.readLoop()

	// Start connection health monitor
	go c.healthMonitor()

	return nil
}

// connect establishes the WebSocket connection
func (c *WebSocketClient) connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.setConnected(true)

	// Send connected event
	c.events <- Event{Type: EventConnected}

	// Subscribe to agent status changes
	if err := c.subscribe("agent_status"); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to subscribe to agent_status: %w", err)
	}

	// Subscribe to new messages
	if err := c.subscribe("new_message"); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to subscribe to new_message: %w", err)
	}

	return nil
}

// subscribe sends a subscription message for a specific event type
func (c *WebSocketClient) subscribe(eventType string) error {
	subscribeMsg := map[string]interface{}{
		"action": "subscribe",
		"event":  eventType,
	}

	return c.conn.WriteJSON(subscribeMsg)
}

// readLoop continuously reads messages from the WebSocket connection
func (c *WebSocketClient) readLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
			c.connMutex.RLock()
			conn := c.conn
			c.connMutex.RUnlock()

			if conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			var event Event
			err := conn.ReadJSON(&event)
			if err != nil {
				// Connection error - trigger reconnection
				c.setConnected(false)
				c.events <- Event{
					Type:  EventDisconnected,
					Error: err.Error(),
				}

				// Wait before attempting reconnection
				time.Sleep(c.reconnectDelay)
				c.attemptReconnect()
				continue
			}

			// Send event to channel
			select {
			case c.events <- event:
			case <-c.done:
				return
			}
		}
	}
}

// attemptReconnect tries to reconnect with exponential backoff
func (c *WebSocketClient) attemptReconnect() {
	delay := c.reconnectDelay

	for {
		select {
		case <-c.done:
			return
		default:
			if c.IsConnected() {
				return
			}

			if err := c.connect(); err == nil {
				// Successfully reconnected
				return
			}

			// Exponential backoff
			time.Sleep(delay)
			delay *= 2
			if delay > c.maxReconnectDelay {
				delay = c.maxReconnectDelay
			}
		}
	}
}

// healthMonitor periodically checks connection health
func (c *WebSocketClient) healthMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if !c.IsConnected() {
				continue
			}

			// Send ping to check connection health
			c.connMutex.RLock()
			conn := c.conn
			c.connMutex.RUnlock()

			if conn != nil {
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.setConnected(false)
					c.events <- Event{
						Type:  EventDisconnected,
						Error: "ping failed",
					}
				}
			}
		}
	}
}

// Events returns the channel for receiving WebSocket events.
// The channel is buffered and will not block unless the buffer is full.
func (c *WebSocketClient) Events() <-chan Event {
	return c.events
}

// IsConnected returns true if the WebSocket connection is active.
func (c *WebSocketClient) IsConnected() bool {
	c.connectedMutex.RLock()
	defer c.connectedMutex.RUnlock()
	return c.connected
}

// setConnected updates the connection status
func (c *WebSocketClient) setConnected(connected bool) {
	c.connectedMutex.Lock()
	defer c.connectedMutex.Unlock()
	c.connected = connected
}

// Close closes the WebSocket connection and stops all background goroutines.
// After calling Close, the client cannot be reused.
func (c *WebSocketClient) Close() error {
	close(c.done)

	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.setConnected(false)
		return err
	}

	return nil
}
