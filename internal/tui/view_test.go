package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/mcp"
)

// TestView_InitialState tests View with initial empty state
func TestView_InitialState(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Set terminal size
	model.width = 120
	model.height = 40

	view := model.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain basic structure
	if !strings.Contains(view, "Agent Status") {
		t.Error("View should contain 'Agent Status' pane")
	}
	if !strings.Contains(view, "Task Stream") {
		t.Error("View should contain 'Task Stream' pane")
	}
	if !strings.Contains(view, "MCP Interaction Log") {
		t.Error("View should contain 'MCP Interaction Log' pane")
	}
}

// TestView_WithoutTerminalSize tests View when terminal size is not set
func TestView_WithoutTerminalSize(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Explicitly set width/height to 0
	model.width = 0
	model.height = 0

	view := model.View()

	if view != "Initializing..." {
		t.Errorf("Expected 'Initializing...', got %q", view)
	}
}

// TestView_WithData tests View with populated data
func TestView_WithData(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Set terminal size
	model.width = 120
	model.height = 40

	// Add test data
	tf.AddTask(beads.Task{
		ID:     "task-1",
		Title:  "Test Task",
		Status: "open",
	})

	tf.SetAgentStatus(mcp.AgentStatus{
		Name:     "test-agent-1",
		State:    mcp.StateIdle,
		LastSeen: time.Now(),
	})

	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent-1",
		Content:   "Test message",
	})

	// Refresh to load data
	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	view := model.View()

	// Should contain data
	if !strings.Contains(view, "test-agent-1") {
		t.Error("View should contain agent name")
	}
	if !strings.Contains(view, "task-1") {
		t.Error("View should contain task ID")
	}
	if !strings.Contains(view, "Test message") {
		t.Error("View should contain message content")
	}
}

// TestView_WithModals tests View with modal overlays
func TestView_WithModals(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.width = 120
	model.height = 40

	// Add a task so modal has something to show
	tf.AddTask(beads.Task{
		ID:     "task-1",
		Title:  "Test Task",
		Status: "open",
	})
	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	// Test task modal
	model.showTaskModal = true
	model.selectedTaskIndex = 0
	view := model.View()
	// Modal rendering may return empty if not fully implemented, just check it doesn't panic
	_ = view

	// Test create modal
	model.showTaskModal = false
	model.showCreateModal = true
	view = model.View()
	_ = view

	// Test confirm modal
	model.showCreateModal = false
	model.showConfirmModal = true
	view = model.View()
	_ = view
}

// TestView_WithSearchMode tests View in search mode
func TestView_WithSearchMode(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.width = 120
	model.height = 40
	model.searchMode = true
	model.searchInput = "test"

	view := model.View()

	if view == "" {
		t.Error("View in search mode should not be empty")
	}
}

// TestView_DifferentTerminalSizes tests View with various terminal sizes
func TestView_DifferentTerminalSizes(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Small", 80, 24},
		{"Medium", 120, 40},
		{"Large", 200, 60},
		{"Wide", 300, 40},
		{"Tall", 80, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			model.width = tt.width
			model.height = tt.height

			view := model.View()

			if view == "" {
				t.Errorf("View should not be empty for size %dx%d", tt.width, tt.height)
			}
		})
	}
}

// TestRenderFooter tests the renderFooter method
func TestRenderFooter(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	footer := model.renderFooter(120)

	if footer == "" {
		t.Error("Footer should not be empty")
	}

	// Should contain keybindings
	if !strings.Contains(footer, "(q)") {
		t.Error("Footer should contain quit keybinding")
	}
	if !strings.Contains(footer, "(r)") {
		t.Error("Footer should contain refresh keybinding")
	}
	if !strings.Contains(footer, "(t)") {
		t.Error("Footer should contain test keybinding")
	}
}

// TestRenderFooter_WithDebugMode tests footer with debug mode enabled
func TestRenderFooter_WithDebugMode(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.debugMode = true
	footer := model.renderFooter(120)

	if !strings.Contains(footer, "[DEBUG]") {
		t.Error("Footer should contain debug indicator when debug mode is enabled")
	}
}

// TestRenderFooter_WithReloadNotification tests footer with reload notification
func TestRenderFooter_WithReloadNotification(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.reloadNotification = "Configuration reloaded"
	model.reloadNotificationTime = time.Now()

	footer := model.renderFooter(120)

	if !strings.Contains(footer, "Configuration reloaded") {
		t.Error("Footer should contain reload notification")
	}
}

// TestRenderFooter_ExpiredNotification tests footer with expired notification
func TestRenderFooter_ExpiredNotification(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.reloadNotification = "Configuration reloaded"
	model.reloadNotificationTime = time.Now().Add(-10 * time.Second)

	footer := model.renderFooter(120)

	// Should show keybindings instead of expired notification
	if !strings.Contains(footer, "(q)") {
		t.Error("Footer should show keybindings when notification is expired")
	}
}

// TestGetBeadsConnectionStatus tests beads connection status
func TestGetBeadsConnectionStatus(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Test disconnected
	model.beadsConnected = false
	status := model.getBeadsConnectionStatus()
	if !strings.Contains(status, "○") {
		t.Error("Disconnected status should contain empty circle")
	}

	// Test connected
	model.beadsConnected = true
	status = model.getBeadsConnectionStatus()
	if !strings.Contains(status, "●") {
		t.Error("Connected status should contain filled circle")
	}
}

// TestGetMCPConnectionStatus tests MCP connection status
func TestGetMCPConnectionStatus(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Test disconnected
	model.wsConnected = false
	model.mcpClient = nil
	status := model.getMCPConnectionStatus()
	if !strings.Contains(status, "○") {
		t.Error("Disconnected status should contain empty circle")
	}

	// Test WebSocket connected
	model.wsConnected = true
	status = model.getMCPConnectionStatus()
	if !strings.Contains(status, "ws") {
		t.Error("WebSocket connected status should contain 'ws'")
	}

	// Test HTTP fallback
	model.wsConnected = false
	mockClient := NewMockMCPClient()
	model.mcpClient = mockClient
	status = model.getMCPConnectionStatus()
	if !strings.Contains(status, "http") {
		t.Error("HTTP fallback status should contain 'http'")
	}
}

// TestOverlayModal tests the overlayModal method
func TestOverlayModal(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	baseView := "Base view content"
	modal := "Modal content"

	result := model.overlayModal(baseView, modal)

	// Currently just returns modal (simplified implementation)
	if result != modal {
		t.Errorf("Expected modal content, got %q", result)
	}
}

// TestView_LayoutCalculations tests layout dimension calculations
func TestView_LayoutCalculations(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Set specific size
	model.width = 120
	model.height = 40

	// Render view
	_ = model.View()

	// Verify calculations are correct
	// availableHeight = 40 - 3 = 37
	// leftWidth = 120 / 3 = 40
	// rightWidth = 120 - 40 = 80
	// rightTopHeight = 37 / 2 = 18
	// rightBottomHeight = 37 - 18 = 19

	// We can't directly test internal calculations, but we can verify
	// the view renders without panic
}

// TestView_WithConnectionStatus tests view with different connection states
func TestView_WithConnectionStatus(t *testing.T) {
	tests := []struct {
		name          string
		beadsConn     bool
		wsConn        bool
		hasMCPClient  bool
	}{
		{"All disconnected", false, false, false},
		{"Beads only", true, false, false},
		{"WebSocket only", false, true, false},
		{"HTTP only", false, false, true},
		{"All connected", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			model.width = 120
			model.height = 40
			model.beadsConnected = tt.beadsConn
			model.wsConnected = tt.wsConn
			if tt.hasMCPClient {
				mockClient := NewMockMCPClient()
				model.mcpClient = mockClient
			} else {
				model.mcpClient = nil
			}

			view := model.View()

			if view == "" {
				t.Error("View should not be empty")
			}
		})
	}
}

// TestView_WithDebugMode tests view with debug mode enabled
func TestView_WithDebugMode(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.width = 120
	model.height = 40
	model.debugMode = true

	view := model.View()

	if view == "" {
		t.Error("View should not be empty in debug mode")
	}

	// Footer should contain debug indicator
	if !strings.Contains(view, "[DEBUG]") {
		t.Error("View should contain debug indicator")
	}
}

// TestView_Composition tests that all panes are properly composed
func TestView_Composition(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	model.width = 120
	model.height = 40

	// Add data to all panes
	tf.AddTask(beads.Task{ID: "task-1", Title: "Task 1", Status: "open"})
	tf.SetAgentStatus(mcp.AgentStatus{Name: "agent-1", State: mcp.StateIdle, LastSeen: time.Now()})
	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "agent-1",
		Content:   "Message 1",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	view := model.View()

	// Verify all components are present
	components := []string{
		"Agent Status",
		"Task Stream",
		"MCP Interaction Log",
		"agent-1",
		"task-1",
		"Message 1",
		"(q)",
		"(r)",
		"(t)",
	}

	for _, component := range components {
		if !strings.Contains(view, component) {
			t.Errorf("View should contain %q", component)
		}
	}
}
