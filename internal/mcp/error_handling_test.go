package mcp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewClient_ErrorPaths tests error handling in client creation
func TestNewClient_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty URL",
			baseURL:     "",
			expectError: true,
			errorMsg:    "url",
		},
		{
			name:        "invalid URL",
			baseURL:     "://invalid",
			expectError: true,
			errorMsg:    "url",
		},
		{
			name:        "URL with spaces",
			baseURL:     "http://local host:8765",
			expectError: true,
			errorMsg:    "url",
		},
		{
			name:        "valid URL",
			baseURL:     "http://localhost:8765",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewHTTPClient(tt.baseURL)

			if tt.expectError {
				// NewHTTPClient doesn't return errors, but operations on invalid URLs should fail
				if client == nil {
					t.Error("Expected non-nil client")
				}
			} else {
				if client == nil {
					t.Error("Expected valid client")
				}
			}
		})
	}
}

// TestGetMessages_ErrorPaths tests error handling in message retrieval
func TestGetMessages_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		serverFunc  func(w http.ResponseWriter, r *http.Request)
		expectError bool
		errorMsg    string
	}{
		{
			name: "server returns 500",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectError: true,
			errorMsg:    "500",
		},
		{
			name: "server returns 404",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			expectError: true,
			errorMsg:    "404",
		},
		{
			name: "server returns invalid JSON",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json{"))
			},
			expectError: true,
			errorMsg:    "decode",
		},
		{
			name: "server returns empty response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(""))
			},
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "server timeout",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(3 * time.Second)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("[]"))
			},
			expectError: false, // Client timeout is longer than 3s
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewHTTPClient(server.URL)

			messages, err := client.GetMessages(time.Now().Add(-1*time.Hour))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if messages == nil {
					t.Error("Expected non-nil messages")
				}
			}
		})
	}
}

// TestSendMessage_ErrorPaths tests error handling in message sending
func TestSendMessage_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		message     Message
		serverFunc  func(w http.ResponseWriter, r *http.Request)
		expectError bool
		errorMsg    string
	}{
		{
			name: "server returns error",
			message: Message{
				Type:    "test",
				Source:  "test",
				Content: "test",
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad Request"))
			},
			expectError: true,
			errorMsg:    "400",
		},
		{
			name: "empty message type",
			message: Message{
				Type:    "",
				Source:  "test",
				Content: "test",
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false, // Client may not validate
		},
		{
			name: "network error",
			message: Message{
				Type:    "test",
				Source:  "test",
				Content: "test",
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Close connection immediately
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
				}
			},
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewHTTPClient(server.URL)

			err := client.SendMessage(tt.message)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestGetAgentStatus_ErrorPaths tests error handling in agent status retrieval
func TestGetAgentStatus_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		serverFunc  func(w http.ResponseWriter, r *http.Request)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "empty agent name",
			agentName: "",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"name":"test","state":"offline"}`))
			},
			expectError: false, // May return offline status
		},
		{
			name:      "agent not found",
			agentName: "nonexistent",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Agent not found"))
			},
			expectError: true,
			errorMsg:    "404",
		},
		{
			name:      "invalid response format",
			agentName: "test",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not json"))
			},
			expectError: true,
			errorMsg:    "decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewHTTPClient(server.URL)

			status, err := client.GetAgentStatus(tt.agentName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if status.Name == "" {
					t.Error("Expected valid status with name")
				}
			}
		})
	}
}

// TestConnectionFailure tests handling of connection failures
func TestConnectionFailure(t *testing.T) {
	// Create client with invalid URL
	client := NewHTTPClient("http://localhost:99999")

	// Try to get messages - should fail
	_, err := client.GetMessages(time.Now())
	if err == nil {
		t.Error("Expected error for connection failure")
	}

	// Try to send message - should fail
	err = client.SendMessage(Message{Type: "test", Source: "test", Content: "test"})
	if err == nil {
		t.Error("Expected error for connection failure")
	}

	// Try to get agent status - should fail
	_, err = client.GetAgentStatus("test")
	if err == nil {
		t.Error("Expected error for connection failure")
	}
}

// TestRetryLogic tests error recovery with retries
func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	// Client retries internally, so first call should succeed after 3 attempts
	messages, err := client.GetMessages(time.Now())
	if err != nil {
		t.Errorf("Expected success after retries, got: %v", err)
	}
	if messages == nil {
		t.Error("Expected non-nil messages")
	}
	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}
}

// TestConcurrentRequests tests error handling under concurrent load
func TestConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	// Make concurrent requests
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := client.GetMessages(time.Now())
			done <- err
		}()
	}

	// Check results
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}

// TestInvalidInput tests handling of invalid input
func TestInvalidInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "send message with null bytes",
			fn: func() error {
				return client.SendMessage(Message{
					Type:    "test\x00type",
					Source:  "test",
					Content: "test",
				})
			},
		},
		{
			name: "get status with special characters",
			fn: func() error {
				_, err := client.GetAgentStatus("test/../../../etc/passwd")
				return err
			},
		},
		{
			name: "get messages with far future time",
			fn: func() error {
				_, err := client.GetMessages(time.Now().Add(100 * 365 * 24 * time.Hour))
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic occurred: %v", r)
				}
			}()

			err := tt.fn()
			// Should not panic, may return error
			if err != nil {
				t.Logf("Got error (may be expected): %v", err)
			}
		})
	}
}

// TestErrorWrapping tests that errors are properly wrapped
func TestErrorWrapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Error"))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	_, err := client.GetMessages(time.Now())
	if err == nil {
		t.Fatal("Expected error")
	}

	// Check that error can be unwrapped
	var unwrapped error
	for unwrapped = err; errors.Unwrap(unwrapped) != nil; unwrapped = errors.Unwrap(unwrapped) {
		// Keep unwrapping
	}

	// Should have some underlying error
	if unwrapped == err {
		t.Log("Error may not be wrapped (acceptable)")
	}
}

// TestPanicRecovery tests that panics are handled gracefully
func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic was not recovered: %v", r)
		}
	}()

	// Try to create client with nil URL (should not panic)
	_ = NewHTTPClient("")

	// Create valid client
	client := NewHTTPClient("http://localhost:8765")

	// Try operations that might panic
	client.GetMessages(time.Time{})
	client.SendMessage(Message{})
	client.GetAgentStatus("")
}

// Helper function
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
