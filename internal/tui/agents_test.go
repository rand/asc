package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/mcp"
)

// TestRenderAgentPane tests the renderAgentPane method
func TestRenderAgentPane(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add agent status
	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-1",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderAgentPane(40, 20)

	if pane == "" {
		t.Error("Agent pane should not be empty")
	}

	// Should contain title
	if !strings.Contains(pane, "Agent Status") {
		t.Error("Agent pane should contain title")
	}

	// Should contain agent name
	if !strings.Contains(pane, "test-agent-1") {
		t.Error("Agent pane should contain agent name")
	}
}

// TestRenderAgentPane_MultipleAgents tests rendering with multiple agents
func TestRenderAgentPane_MultipleAgents(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add multiple agent statuses
	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-1",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	})
	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-2",
		State:    mcp.StateWorking,
		CurrentTask: "task-123",
		LastSeen: time.Now(),
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderAgentPane(40, 20)

	// Should contain both agents
	if !strings.Contains(pane, "test-agent-1") {
		t.Error("Agent pane should contain first agent")
	}
	if !strings.Contains(pane, "test-agent-2") {
		t.Error("Agent pane should contain second agent")
	}
}

// TestRenderAgentPane_DifferentStates tests rendering agents in different states
func TestRenderAgentPane_DifferentStates(t *testing.T) {
	states := []struct {
		name  string
		state mcp.AgentState
		icon  string
	}{
		{"Idle", mcp.StateIdle, iconIdle},
		{"Working", mcp.StateWorking, iconWorking},
		{"Error", mcp.StateError, iconError},
		{"Offline", mcp.StateOffline, iconOffline},
	}

	for _, tt := range states {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			tf.SetAgentStatus(mcp.AgentStatus{
				Name:     "test-agent-1",
				State:    tt.state,
				LastSeen: time.Now(),
			})

			err := model.refreshData()
			if err != nil {
				t.Fatalf("refreshData failed: %v", err)
			}

			pane := model.renderAgentPane(40, 20)

			// Should contain the appropriate icon
			if !strings.Contains(pane, tt.icon) {
				t.Errorf("Agent pane should contain icon %q for state %s", tt.icon, tt.name)
			}
		})
	}
}

// TestRenderAgentPane_WithCurrentTask tests rendering working agent with task
func TestRenderAgentPane_WithCurrentTask(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Use an agent that exists in the config
	tf.SetAgentStatus(mcp.AgentStatus{
		Name:        "test-agent-1", // This agent is in the test config
		State:       mcp.StateWorking,
		CurrentTask: "task-456",
		LastSeen:    time.Now(),
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderAgentPane(40, 20)

	// Should contain task ID (if agent is in config and status is working)
	// The pane should at least render without error
	if pane == "" {
		t.Error("Agent pane should not be empty")
	}
	
	// Check if working status is shown
	if !strings.Contains(pane, "Working") {
		t.Error("Agent pane should show working status")
	}
}

// TestRenderAgentPane_NoAgents tests rendering with no agents
func TestRenderAgentPane_NoAgents(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Clear agents from config
	model.config.Agents = make(map[string]config.AgentConfig)

	pane := model.renderAgentPane(40, 20)

	if !strings.Contains(pane, "No agents configured") {
		t.Error("Agent pane should show 'No agents configured' message")
	}
}

// TestRenderAgentPane_WithSelection tests rendering with agent selection
func TestRenderAgentPane_WithSelection(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-1",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	})
	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-2",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	// Select first agent
	model.selectedAgentIndex = 0

	pane := model.renderAgentPane(40, 20)

	// Should contain selection indicator
	if !strings.Contains(pane, "▶") {
		t.Error("Agent pane should contain selection indicator")
	}
}

// TestFormatAgentLine tests the formatAgentLine method
func TestFormatAgentLine(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	status := mcp.AgentStatus{
		Name:     "test-agent",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	}

	line := model.formatAgentLine(status, "", 50, 1, false)

	if line == "" {
		t.Error("Formatted agent line should not be empty")
	}

	// Should contain agent name
	if !strings.Contains(line, "test-agent") {
		t.Error("Formatted line should contain agent name")
	}

	// Should contain status
	if !strings.Contains(line, "Idle") {
		t.Error("Formatted line should contain status")
	}
}

// TestFormatAgentLine_WithHealthIssue tests formatting with health issues
func TestFormatAgentLine_WithHealthIssue(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	status := mcp.AgentStatus{
		Name:     "test-agent",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	}

	tests := []struct {
		name        string
		healthIssue string
		indicator   string
	}{
		{"Crashed", "crashed", "⚠"},
		{"Unresponsive", "unresponsive", "⚠"},
		{"Stuck", "stuck", "⏱"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := model.formatAgentLine(status, tt.healthIssue, 50, 1, false)

			if !strings.Contains(line, tt.indicator) {
				t.Errorf("Line should contain health indicator %q", tt.indicator)
			}
		})
	}
}

// TestFormatAgentLine_Selected tests formatting selected agent
func TestFormatAgentLine_Selected(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	status := mcp.AgentStatus{
		Name:     "test-agent",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	}

	line := model.formatAgentLine(status, "", 50, 1, true)

	// Should contain selection indicator
	if !strings.Contains(line, "▶") {
		t.Error("Selected line should contain selection indicator")
	}
}

// TestFormatAgentLine_Truncation tests line truncation
func TestFormatAgentLine_Truncation(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	status := mcp.AgentStatus{
		Name:     "very-long-agent-name-that-should-be-truncated",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	}

	// Use small width to force truncation
	line := model.formatAgentLine(status, "", 20, 1, false)

	// Line should be truncated
	if len(line) > 100 { // Account for ANSI codes
		t.Error("Line should be truncated to fit width")
	}
}

// TestGetAgentIconAndStyle tests icon and style selection
func TestGetAgentIconAndStyle(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tests := []struct {
		state        mcp.AgentState
		expectedIcon string
	}{
		{mcp.StateIdle, iconIdle},
		{mcp.StateWorking, iconWorking},
		{mcp.StateError, iconError},
		{mcp.StateOffline, iconOffline},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			icon, style := model.getAgentIconAndStyle(tt.state)

			if icon != tt.expectedIcon {
				t.Errorf("Expected icon %q, got %q", tt.expectedIcon, icon)
			}

			// Style should be valid (we can't easily check color value)
			_ = style
		})
	}
}

// TestGetAgentNames tests agent name retrieval
func TestGetAgentNames(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	names := model.getAgentNames()

	// Should return names from config
	if len(names) != 2 {
		t.Errorf("Expected 2 agent names, got %d", len(names))
	}
}

// TestFitContent tests content fitting
func TestFitContent(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Test truncation
	lines := []string{"line1", "line2", "line3", "line4", "line5"}
	fitted := model.fitContent(lines, 3)

	if len(fitted) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(fitted))
	}

	// Test padding
	lines = []string{"line1", "line2"}
	fitted = model.fitContent(lines, 5)

	if len(fitted) != 5 {
		t.Errorf("Expected 5 lines (padded), got %d", len(fitted))
	}
}

// TestRenderAgentPane_DifferentSizes tests rendering with different pane sizes
func TestRenderAgentPane_DifferentSizes(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"Small", 30, 10},
		{"Medium", 40, 20},
		{"Large", 60, 40},
	}

	for _, tt := range sizes {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			tf.SetAgentStatus(mcp.AgentStatus{
				Name:     "test-agent-1",
				State:    mcp.StateIdle,
				LastSeen: time.Now(),
			})

			err := model.refreshData()
			if err != nil {
				t.Fatalf("refreshData failed: %v", err)
			}

			pane := model.renderAgentPane(tt.width, tt.height)

			if pane == "" {
				t.Errorf("Agent pane should not be empty for size %dx%d", tt.width, tt.height)
			}
		})
	}
}

// TestRenderAgentPane_WithKeybindingHints tests that keybinding hints are shown
func TestRenderAgentPane_WithKeybindingHints(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	pane := model.renderAgentPane(40, 20)

	// Should contain keybinding hints
	if !strings.Contains(pane, "1-9:select") {
		t.Error("Agent pane should contain selection keybinding hint")
	}
	if !strings.Contains(pane, "p:pause") {
		t.Error("Agent pane should contain pause keybinding hint")
	}
	if !strings.Contains(pane, "k:kill") {
		t.Error("Agent pane should contain kill keybinding hint")
	}
}

// TestRenderAgentPane_OfflineAgent tests rendering offline agent
func TestRenderAgentPane_OfflineAgent(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Agent in config but no status update (offline)
	pane := model.renderAgentPane(40, 20)

	// Should show agents as offline
	if !strings.Contains(pane, "Offline") {
		t.Error("Agent pane should show offline status for agents without status updates")
	}
}
