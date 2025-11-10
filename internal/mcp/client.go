// Package mcp provides a client for interacting with the MCP agent mail server.
// It handles HTTP communication for sending messages, retrieving agent status,
// and tracking heartbeats, with support for retries and error handling.
//
// Example usage:
//
//	client := mcp.NewHTTPClient("http://localhost:8765")
//	messages, err := client.GetMessages(time.Now().Add(-5 * time.Minute))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, msg := range messages {
//	    fmt.Printf("[%s] %s: %s\n", msg.Type, msg.Source, msg.Content)
//	}
package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MessageType represents the type of MCP message.
type MessageType string

const (
	TypeLease   MessageType = "lease"
	TypeBeads   MessageType = "beads"
	TypeError   MessageType = "error"
	TypeMessage MessageType = "message"
)

// AgentState represents the current state of an agent in the system.
type AgentState string

const (
	StateIdle    AgentState = "idle"
	StateWorking AgentState = "working"
	StateError   AgentState = "error"
	StateOffline AgentState = "offline"
)

// Message represents an MCP message exchanged between agents or services.
type Message struct {
	Timestamp time.Time   `json:"timestamp"`
	Type      MessageType `json:"type"`
	Source    string      `json:"source"`
	Content   string      `json:"content"`
}

// AgentStatus represents the status of an agent including its current state,
// task, and last seen timestamp.
type AgentStatus struct {
	Name        string     `json:"name"`
	State       AgentState `json:"state"`
	CurrentTask string     `json:"current_task"`
	LastSeen    time.Time  `json:"last_seen"`
}

// MCPClient defines the interface for interacting with the MCP server
type MCPClient interface {
	GetMessages(since time.Time) ([]Message, error)
	SendMessage(msg Message) error
	GetAgentStatus(agentName string) (AgentStatus, error)
}

// HTTPClient implements the MCPClient interface using HTTP requests.
// It includes retry logic and configurable timeouts.
type HTTPClient struct {
	baseURL    string        // Base URL of the MCP server
	httpClient *http.Client  // HTTP client with timeout
	maxRetries int           // Maximum number of retry attempts
	retryDelay time.Duration // Base delay between retries
}

// NewHTTPClient creates a new HTTP-based MCP client with the specified base URL.
// The client is configured with a 10-second timeout and 3 retry attempts.
//
// Example:
//
//	client := mcp.NewHTTPClient("http://localhost:8765")
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// GetMessages retrieves messages from the MCP server since the given timestamp.
// Returns an empty slice if no messages are available. Retries on network errors.
func (c *HTTPClient) GetMessages(since time.Time) ([]Message, error) {
	url := fmt.Sprintf("%s/messages?since=%d", c.baseURL, since.Unix())
	
	var messages []Message
	err := c.doRequestWithRetry("GET", url, nil, &messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	return messages, nil
}

// SendMessage sends a message to the MCP server.
// The message is serialized to JSON and sent via HTTP POST. Retries on network errors.
func (c *HTTPClient) SendMessage(msg Message) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)
	
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	err = c.doRequestWithRetry("POST", url, jsonData, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	return nil
}

// GetAgentStatus retrieves the status of a specific agent by name.
// Returns an error if the agent is not found or the request fails.
func (c *HTTPClient) GetAgentStatus(agentName string) (AgentStatus, error) {
	url := fmt.Sprintf("%s/agents/%s/status", c.baseURL, agentName)
	
	var status AgentStatus
	err := c.doRequestWithRetry("GET", url, nil, &status)
	if err != nil {
		return AgentStatus{}, fmt.Errorf("failed to get agent status: %w", err)
	}
	
	return status, nil
}

// doRequestWithRetry performs an HTTP request with retry logic
func (c *HTTPClient) doRequestWithRetry(method, url string, body []byte, result interface{}) error {
	var lastErr error
	
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}
		
		err := c.doRequest(method, url, body, result)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Don't retry on client errors (4xx)
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return lastErr
		}
	}
	
	return fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// doRequest performs a single HTTP request
func (c *HTTPClient) doRequest(method, url string, body []byte, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}
	
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// HTTPError represents an HTTP error response with status code and message.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Heartbeat represents an agent heartbeat message used to track agent liveness.
type Heartbeat struct {
	AgentName   string     `json:"agent_name"`
	State       AgentState `json:"state"`
	CurrentTask string     `json:"current_task,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

// GetHeartbeats retrieves agent heartbeats from the MCP server.
// Heartbeats are used to determine agent liveness and current state.
func (c *HTTPClient) GetHeartbeats() ([]Heartbeat, error) {
	url := fmt.Sprintf("%s/heartbeats", c.baseURL)
	
	var heartbeats []Heartbeat
	err := c.doRequestWithRetry("GET", url, nil, &heartbeats)
	if err != nil {
		return nil, fmt.Errorf("failed to get heartbeats: %w", err)
	}
	
	return heartbeats, nil
}

// GetAllAgentStatuses retrieves the status of all agents based on heartbeats.
// Agents that haven't sent a heartbeat within the offlineThreshold are marked as offline.
func (c *HTTPClient) GetAllAgentStatuses(offlineThreshold time.Duration) ([]AgentStatus, error) {
	heartbeats, err := c.GetHeartbeats()
	if err != nil {
		return nil, err
	}
	
	statuses := make([]AgentStatus, 0, len(heartbeats))
	now := time.Now()
	
	for _, hb := range heartbeats {
		status := c.heartbeatToStatus(hb, now, offlineThreshold)
		statuses = append(statuses, status)
	}
	
	return statuses, nil
}

// heartbeatToStatus converts a heartbeat to an AgentStatus
func (c *HTTPClient) heartbeatToStatus(hb Heartbeat, now time.Time, offlineThreshold time.Duration) AgentStatus {
	status := AgentStatus{
		Name:        hb.AgentName,
		State:       hb.State,
		CurrentTask: hb.CurrentTask,
		LastSeen:    hb.Timestamp,
	}
	
	// Check if agent is offline based on last seen time
	if now.Sub(hb.Timestamp) > offlineThreshold {
		status.State = StateOffline
	}
	
	return status
}

// TrackAgentStatus polls the MCP server for a specific agent's status.
// Returns the agent status based on its most recent heartbeat, or marks it
// as offline if no heartbeat is found or it exceeds the offline threshold.
func (c *HTTPClient) TrackAgentStatus(agentName string, offlineThreshold time.Duration) (AgentStatus, error) {
	heartbeats, err := c.GetHeartbeats()
	if err != nil {
		return AgentStatus{}, err
	}
	
	now := time.Now()
	
	// Find the heartbeat for the specified agent
	for _, hb := range heartbeats {
		if hb.AgentName == agentName {
			return c.heartbeatToStatus(hb, now, offlineThreshold), nil
		}
	}
	
	// Agent not found in heartbeats - consider it offline
	return AgentStatus{
		Name:     agentName,
		State:    StateOffline,
		LastSeen: time.Time{},
	}, nil
}
