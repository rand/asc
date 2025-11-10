package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// mockWebSocketServer creates a test WebSocket server
func mockWebSocketServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()
		handler(conn)
	}))
}

func TestWebSocketClient_Connect(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			// Verify subscription format
			if msg["action"] != "subscribe" {
				t.Errorf("Expected subscribe action, got %v", msg["action"])
			}
		}
		// Keep connection open
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)
	defer client.Close()

	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connected event
	select {
	case event := <-client.Events():
		if event.Type != EventConnected {
			t.Errorf("Expected connected event, got %v", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for connected event")
	}

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}
}

func TestWebSocketClient_ReceiveAgentStatusEvent(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			conn.ReadJSON(&msg)
		}

		// Send agent status event
		event := Event{
			Type: EventAgentStatus,
			AgentStatus: &AgentStatus{
				Name:        "test-agent",
				State:       StateWorking,
				CurrentTask: "task-123",
				LastSeen:    time.Now(),
			},
		}
		conn.WriteJSON(event)

		// Keep connection open
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connected event
	<-client.Events()

	// Wait for agent status event
	select {
	case event := <-client.Events():
		if event.Type != EventAgentStatus {
			t.Errorf("Expected agent status event, got %v", event.Type)
		}
		if event.AgentStatus == nil {
			t.Fatal("Agent status should not be nil")
		}
		if event.AgentStatus.Name != "test-agent" {
			t.Errorf("Expected agent name 'test-agent', got %s", event.AgentStatus.Name)
		}
		if event.AgentStatus.State != StateWorking {
			t.Errorf("Expected state working, got %s", event.AgentStatus.State)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for agent status event")
	}
}

func TestWebSocketClient_ReceiveNewMessageEvent(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			conn.ReadJSON(&msg)
		}

		// Send new message event
		event := Event{
			Type: EventNewMessage,
			Message: &Message{
				Timestamp: time.Now(),
				Type:      TypeBeads,
				Source:    "test-agent",
				Content:   "test message content",
			},
		}
		conn.WriteJSON(event)

		// Keep connection open
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connected event
	<-client.Events()

	// Wait for new message event
	select {
	case event := <-client.Events():
		if event.Type != EventNewMessage {
			t.Errorf("Expected new message event, got %v", event.Type)
		}
		if event.Message == nil {
			t.Fatal("Message should not be nil")
		}
		if event.Message.Source != "test-agent" {
			t.Errorf("Expected source 'test-agent', got %s", event.Message.Source)
		}
		if event.Message.Content != "test message content" {
			t.Errorf("Expected content 'test message content', got %s", event.Message.Content)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for new message event")
	}
}

func TestWebSocketClient_Reconnection(t *testing.T) {
	// Create a server that closes connection after first message
	connectionCount := 0
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		connectionCount++
		
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
		}

		if connectionCount == 1 {
			// Close connection immediately to trigger reconnection
			conn.Close()
		} else {
			// Keep second connection open
			time.Sleep(200 * time.Millisecond)
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for initial connected event
	<-client.Events()

	// Wait for disconnected event
	select {
	case event := <-client.Events():
		if event.Type != EventDisconnected {
			t.Errorf("Expected disconnected event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for disconnected event")
	}

	// Wait for reconnected event
	select {
	case event := <-client.Events():
		if event.Type != EventConnected {
			t.Errorf("Expected reconnected event, got %v", event.Type)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for reconnected event")
	}

	if connectionCount < 2 {
		t.Errorf("Expected at least 2 connections, got %d", connectionCount)
	}
}

func TestWebSocketClient_Close(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			conn.ReadJSON(&msg)
		}
		// Keep connection open
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connected event
	<-client.Events()

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// Close the client
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	if client.IsConnected() {
		t.Error("Client should be disconnected after close")
	}
}

func TestWebSocketClient_EventBuffering(t *testing.T) {
	server := mockWebSocketServer(t, func(conn *websocket.Conn) {
		// Read subscription messages
		for i := 0; i < 2; i++ {
			var msg map[string]interface{}
			conn.ReadJSON(&msg)
		}

		// Send multiple events rapidly
		for i := 0; i < 10; i++ {
			event := Event{
				Type: EventNewMessage,
				Message: &Message{
					Timestamp: time.Now(),
					Type:      TypeMessage,
					Source:    "test-agent",
					Content:   "test message",
				},
			}
			conn.WriteJSON(event)
		}

		// Keep connection open
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWebSocketClient(wsURL)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connected event
	<-client.Events()

	// Collect events
	eventCount := 0
	timeout := time.After(2 * time.Second)
	
	for eventCount < 10 {
		select {
		case event := <-client.Events():
			if event.Type == EventNewMessage {
				eventCount++
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for events, received %d out of 10", eventCount)
		}
	}

	if eventCount != 10 {
		t.Errorf("Expected 10 events, got %d", eventCount)
	}
}
