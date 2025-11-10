package mcp

import (
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient("http://localhost:8765")
	if client == nil {
		t.Fatal("NewHTTPClient returned nil")
	}

	if client.baseURL != "http://localhost:8765" {
		t.Errorf("baseURL = %v, want http://localhost:8765", client.baseURL)
	}
	if client.maxRetries != 3 {
		t.Errorf("maxRetries = %v, want 3", client.maxRetries)
	}
	if client.retryDelay != 1*time.Second {
		t.Errorf("retryDelay = %v, want 1s", client.retryDelay)
	}
}

func TestMessage(t *testing.T) {
	now := time.Now()
	msg := Message{
		Timestamp: now,
		Type:      TypeLease,
		Source:    "agent-1",
		Content:   "Leasing file.go",
	}

	if msg.Timestamp != now {
		t.Errorf("Timestamp mismatch")
	}
	if msg.Type != TypeLease {
		t.Errorf("Type = %v, want %v", msg.Type, TypeLease)
	}
	if msg.Source != "agent-1" {
		t.Errorf("Source = %v, want agent-1", msg.Source)
	}
	if msg.Content != "Leasing file.go" {
		t.Errorf("Content = %v, want Leasing file.go", msg.Content)
	}
}

func TestMessageTypes(t *testing.T) {
	types := []MessageType{TypeLease, TypeBeads, TypeError, TypeMessage}
	expected := []string{"lease", "beads", "error", "message"}

	for i, msgType := range types {
		if string(msgType) != expected[i] {
			t.Errorf("MessageType %d = %v, want %v", i, msgType, expected[i])
		}
	}
}

func TestAgentStates(t *testing.T) {
	states := []AgentState{StateIdle, StateWorking, StateError, StateOffline}
	expected := []string{"idle", "working", "error", "offline"}

	for i, state := range states {
		if string(state) != expected[i] {
			t.Errorf("AgentState %d = %v, want %v", i, state, expected[i])
		}
	}
}

func TestAgentStatus(t *testing.T) {
	now := time.Now()
	status := AgentStatus{
		Name:        "test-agent",
		State:       StateWorking,
		CurrentTask: "task-123",
		LastSeen:    now,
	}

	if status.Name != "test-agent" {
		t.Errorf("Name = %v, want test-agent", status.Name)
	}
	if status.State != StateWorking {
		t.Errorf("State = %v, want %v", status.State, StateWorking)
	}
	if status.CurrentTask != "task-123" {
		t.Errorf("CurrentTask = %v, want task-123", status.CurrentTask)
	}
	if status.LastSeen != now {
		t.Errorf("LastSeen mismatch")
	}
}

func TestHeartbeat(t *testing.T) {
	now := time.Now()
	hb := Heartbeat{
		AgentName:   "agent-1",
		State:       StateIdle,
		CurrentTask: "",
		Timestamp:   now,
	}

	if hb.AgentName != "agent-1" {
		t.Errorf("AgentName = %v, want agent-1", hb.AgentName)
	}
	if hb.State != StateIdle {
		t.Errorf("State = %v, want %v", hb.State, StateIdle)
	}
	if hb.CurrentTask != "" {
		t.Errorf("CurrentTask should be empty")
	}
	if hb.Timestamp != now {
		t.Errorf("Timestamp mismatch")
	}
}

func TestHTTPError(t *testing.T) {
	err := &HTTPError{
		StatusCode: 404,
		Message:    "Not Found",
	}

	expected := "HTTP 404: Not Found"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestHTTPErrorCodes(t *testing.T) {
	tests := []struct {
		code    int
		message string
	}{
		{400, "Bad Request"},
		{401, "Unauthorized"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{503, "Service Unavailable"},
	}

	for _, tt := range tests {
		err := &HTTPError{
			StatusCode: tt.code,
			Message:    tt.message,
		}
		if err.StatusCode != tt.code {
			t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.code)
		}
		if err.Message != tt.message {
			t.Errorf("Message = %v, want %v", err.Message, tt.message)
		}
	}
}

func TestClientInterface(t *testing.T) {
	var _ MCPClient = (*HTTPClient)(nil)
}

func TestMultipleMessages(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{Timestamp: now, Type: TypeLease, Source: "agent-1", Content: "Lease file1.go"},
		{Timestamp: now, Type: TypeBeads, Source: "agent-2", Content: "Claimed task-1"},
		{Timestamp: now, Type: TypeMessage, Source: "agent-3", Content: "Ready"},
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	types := []MessageType{TypeLease, TypeBeads, TypeMessage}
	for i, msg := range messages {
		if msg.Type != types[i] {
			t.Errorf("Message %d: Type = %v, want %v", i, msg.Type, types[i])
		}
	}
}

func TestAgentStatusIdle(t *testing.T) {
	status := AgentStatus{
		Name:        "idle-agent",
		State:       StateIdle,
		CurrentTask: "",
		LastSeen:    time.Now(),
	}

	if status.State != StateIdle {
		t.Errorf("State = %v, want %v", status.State, StateIdle)
	}
	if status.CurrentTask != "" {
		t.Errorf("CurrentTask should be empty for idle agent")
	}
}

func TestAgentStatusWorking(t *testing.T) {
	status := AgentStatus{
		Name:        "working-agent",
		State:       StateWorking,
		CurrentTask: "task-456",
		LastSeen:    time.Now(),
	}

	if status.State != StateWorking {
		t.Errorf("State = %v, want %v", status.State, StateWorking)
	}
	if status.CurrentTask == "" {
		t.Errorf("CurrentTask should not be empty for working agent")
	}
}

func TestAgentStatusOffline(t *testing.T) {
	oldTime := time.Now().Add(-10 * time.Minute)
	status := AgentStatus{
		Name:     "offline-agent",
		State:    StateOffline,
		LastSeen: oldTime,
	}

	if status.State != StateOffline {
		t.Errorf("State = %v, want %v", status.State, StateOffline)
	}
}

func TestHeartbeatWithTask(t *testing.T) {
	hb := Heartbeat{
		AgentName:   "busy-agent",
		State:       StateWorking,
		CurrentTask: "task-789",
		Timestamp:   time.Now(),
	}

	if hb.State != StateWorking {
		t.Errorf("State = %v, want %v", hb.State, StateWorking)
	}
	if hb.CurrentTask != "task-789" {
		t.Errorf("CurrentTask = %v, want task-789", hb.CurrentTask)
	}
}

func TestMessageTimestamps(t *testing.T) {
	now := time.Now()
	past := now.Add(-5 * time.Minute)
	future := now.Add(5 * time.Minute)

	messages := []Message{
		{Timestamp: past, Type: TypeMessage, Source: "agent-1", Content: "Old message"},
		{Timestamp: now, Type: TypeMessage, Source: "agent-2", Content: "Current message"},
		{Timestamp: future, Type: TypeMessage, Source: "agent-3", Content: "Future message"},
	}

	if !messages[0].Timestamp.Before(messages[1].Timestamp) {
		t.Errorf("Past message should be before current")
	}
	if !messages[1].Timestamp.Before(messages[2].Timestamp) {
		t.Errorf("Current message should be before future")
	}
}

func TestClientBaseURLVariations(t *testing.T) {
	urls := []string{
		"http://localhost:8765",
		"http://127.0.0.1:8765",
		"http://0.0.0.0:8765",
		"http://example.com:8765",
	}

	for _, url := range urls {
		client := NewHTTPClient(url)
		if client.baseURL != url {
			t.Errorf("baseURL = %v, want %v", client.baseURL, url)
		}
	}
}

func TestReleaseAgentLeases(t *testing.T) {
	// This test verifies that the ReleaseAgentLeases method is properly defined
	// and can be called on the HTTPClient. Integration testing with a real MCP
	// server would be needed to verify the actual HTTP request.
	client := NewHTTPClient("http://localhost:8765")
	
	// Verify the method exists and returns an error (expected since no server is running)
	err := client.ReleaseAgentLeases("test-agent")
	
	// We expect an error since there's no actual server running
	if err == nil {
		t.Log("Note: ReleaseAgentLeases succeeded (unexpected without a running server)")
	}
	
	// The important thing is that the method exists and can be called
	// The actual functionality would be tested in integration tests
}
