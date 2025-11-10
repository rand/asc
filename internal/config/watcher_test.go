package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_Basic(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")
	
	// Write initial config
	initialConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "claude"
phases = ["planning"]
`
	
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}
	
	// Create watcher
	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	
	// Modify the config file
	updatedConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "gemini"
phases = ["planning", "implementation"]
`
	
	// Wait a bit before writing to ensure watcher is ready
	time.Sleep(100 * time.Millisecond)
	
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}
	
	// Wait for the event with timeout
	select {
	case newConfig := <-watcher.Events():
		if newConfig == nil {
			t.Fatal("Received nil config")
		}
		
		// Verify the config was updated
		if newConfig.Agents["test-agent"].Model != "gemini" {
			t.Errorf("Expected model 'gemini', got '%s'", newConfig.Agents["test-agent"].Model)
		}
		
		if len(newConfig.Agents["test-agent"].Phases) != 2 {
			t.Errorf("Expected 2 phases, got %d", len(newConfig.Agents["test-agent"].Phases))
		}
		
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for config reload event")
	}
}

func TestWatcher_InvalidConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")
	
	// Write initial valid config
	initialConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "claude"
phases = ["planning"]
`
	
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}
	
	// Create watcher
	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	
	// Write invalid config (should not trigger event)
	invalidConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "invalid-model"
phases = ["planning"]
`
	
	// Wait a bit before writing
	time.Sleep(100 * time.Millisecond)
	
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	
	// Should not receive an event (or timeout is acceptable)
	select {
	case <-watcher.Events():
		t.Fatal("Should not receive event for invalid config")
	case <-time.After(1 * time.Second):
		// Expected - invalid config should not trigger event
	}
}

func TestWatcher_MultipleChanges(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")
	
	// Write initial config
	initialConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "claude"
phases = ["planning"]
`
	
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}
	
	// Create watcher
	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()
	
	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	
	// Make multiple rapid changes (should be debounced)
	for i := 0; i < 5; i++ {
		updatedConfig := `
[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo test"
model = "gemini"
phases = ["planning"]
`
		if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	
	// Should receive only one event due to debouncing
	eventCount := 0
	timeout := time.After(2 * time.Second)
	
	for {
		select {
		case <-watcher.Events():
			eventCount++
			// Continue draining events for a bit
			time.Sleep(100 * time.Millisecond)
		case <-timeout:
			// Check that we got at least one event but not too many
			if eventCount == 0 {
				t.Fatal("Expected at least one event")
			}
			if eventCount > 2 {
				t.Errorf("Expected debouncing to limit events, got %d", eventCount)
			}
			return
		}
	}
}
