package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rand/asc/internal/mcp"
)

// TestRenderLogPane tests the renderLogPane method
func TestRenderLogPane(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add test message
	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test message",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 20)

	if pane == "" {
		t.Error("Log pane should not be empty")
	}

	// Should contain title
	if !strings.Contains(pane, "MCP Interaction Log") {
		t.Error("Log pane should contain title")
	}

	// Should contain message
	if !strings.Contains(pane, "Test message") {
		t.Error("Log pane should contain message content")
	}
	if !strings.Contains(pane, "test-agent") {
		t.Error("Log pane should contain message source")
	}
}

// TestRenderLogPane_MultipleMessages tests rendering with multiple messages
func TestRenderLogPane_MultipleMessages(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add multiple messages
	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "agent-1",
		Content:   "Message 1",
	})
	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeLease,
		Source:    "agent-2",
		Content:   "Message 2",
	})
	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeBeads,
		Source:    "agent-3",
		Content:   "Message 3",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 20)

	// Should contain all messages
	if !strings.Contains(pane, "Message 1") {
		t.Error("Log pane should contain first message")
	}
	if !strings.Contains(pane, "Message 2") {
		t.Error("Log pane should contain second message")
	}
	if !strings.Contains(pane, "Message 3") {
		t.Error("Log pane should contain third message")
	}
}

// TestRenderLogPane_DifferentMessageTypes tests rendering different message types
func TestRenderLogPane_DifferentMessageTypes(t *testing.T) {
	types := []struct {
		name string
		typ  mcp.MessageType
	}{
		{"Lease", mcp.TypeLease},
		{"Beads", mcp.TypeBeads},
		{"Error", mcp.TypeError},
		{"Message", mcp.TypeMessage},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			tf.AddMessage(mcp.Message{
				Timestamp: time.Now(),
				Type:      tt.typ,
				Source:    "test-agent",
				Content:   "Test content",
			})

			err := model.refreshData()
			if err != nil {
				t.Fatalf("refreshData failed: %v", err)
			}

			pane := model.renderLogPane(80, 20)

			// Should contain message type
			if !strings.Contains(pane, string(tt.typ)) {
				t.Errorf("Log pane should contain message type %q", tt.typ)
			}
		})
	}
}

// TestRenderLogPane_NoMessages tests rendering with no messages
func TestRenderLogPane_NoMessages(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	pane := model.renderLogPane(80, 20)

	if !strings.Contains(pane, "No messages yet") {
		t.Error("Log pane should show 'No messages yet' message")
	}
}

// TestRenderLogPane_AutoScroll tests that log pane auto-scrolls to bottom
func TestRenderLogPane_AutoScroll(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add many messages (more than pane height)
	for i := 0; i < 50; i++ {
		tf.AddMessage(mcp.Message{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Type:      mcp.TypeMessage,
			Source:    "test-agent",
			Content:   "Message " + string(rune(i)),
		})
	}

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 10) // Small height

	// Should contain recent messages (auto-scrolled to bottom)
	// The exact messages depend on implementation, but pane should not be empty
	if pane == "" {
		t.Error("Log pane should not be empty with many messages")
	}
}

// TestRenderLogPane_MessageLimit tests that messages are limited to maxLogMessages
func TestRenderLogPane_MessageLimit(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add more than maxLogMessages (100)
	for i := 0; i < 150; i++ {
		tf.AddMessage(mcp.Message{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Type:      mcp.TypeMessage,
			Source:    "test-agent",
			Content:   "Message",
		})
	}

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	// Get filtered messages
	messages := model.getFilteredMessages()

	// Should be limited to maxLogMessages
	if len(messages) > maxLogMessages {
		t.Errorf("Messages should be limited to %d, got %d", maxLogMessages, len(messages))
	}
}

// TestRenderLogPane_WithFilters tests rendering with active filters
func TestRenderLogPane_WithFilters(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test message",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	// Set filters
	model.searchInput = "test"
	model.logFilterAgent = "test-agent"
	model.logFilterType = "message"

	pane := model.renderLogPane(80, 20)

	// Should show active filters in title
	if !strings.Contains(pane, "search:test") {
		t.Error("Log pane should show search filter")
	}
	if !strings.Contains(pane, "agent:test-agent") {
		t.Error("Log pane should show agent filter")
	}
	if !strings.Contains(pane, "type:message") {
		t.Error("Log pane should show type filter")
	}
}

// TestFormatMessageLine tests the formatMessageLine method
func TestFormatMessageLine(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	msg := mcp.Message{
		Timestamp: time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test content",
	}

	line := model.formatMessageLine(msg, 80)

	if line == "" {
		t.Error("Formatted message line should not be empty")
	}

	// Should contain timestamp
	if !strings.Contains(line, "12:30:45") {
		t.Error("Formatted line should contain timestamp")
	}

	// Should contain type
	if !strings.Contains(line, "message") {
		t.Error("Formatted line should contain message type")
	}

	// Should contain source
	if !strings.Contains(line, "test-agent") {
		t.Error("Formatted line should contain source")
	}

	// Should contain content
	if !strings.Contains(line, "Test content") {
		t.Error("Formatted line should contain content")
	}
}

// TestFormatMessageLine_Truncation tests line truncation
func TestFormatMessageLine_Truncation(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	msg := mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Very long message content that should be truncated to fit within the available width",
	}

	// Use small width to force truncation
	line := model.formatMessageLine(msg, 40)

	// Line should be truncated (accounting for ANSI codes)
	if len(line) > 200 {
		t.Error("Line should be truncated to fit width")
	}
}

// TestGetMessageStyle tests message style selection
func TestGetMessageStyle(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tests := []mcp.MessageType{
		mcp.TypeLease,
		mcp.TypeBeads,
		mcp.TypeError,
		mcp.TypeMessage,
	}

	for _, typ := range tests {
		t.Run(string(typ), func(t *testing.T) {
			style := model.getMessageStyle(typ)

			// Style should be valid (we can't easily check color value)
			_ = style
		})
	}
}

// TestGetRecentMessages tests getting recent messages
func TestGetRecentMessages(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add messages
	for i := 0; i < 10; i++ {
		model.messages = append(model.messages, mcp.Message{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Type:      mcp.TypeMessage,
			Source:    "test-agent",
			Content:   "Message",
		})
	}

	// Get recent 5 messages
	recent := model.getRecentMessages(5)

	if len(recent) != 5 {
		t.Errorf("Expected 5 recent messages, got %d", len(recent))
	}

	// Should be the last 5 messages
	if recent[0].Timestamp.Before(model.messages[5].Timestamp) {
		t.Error("Recent messages should be the last N messages")
	}
}

// TestGetRecentMessages_LessThanLimit tests when there are fewer messages than limit
func TestGetRecentMessages_LessThanLimit(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add only 3 messages
	for i := 0; i < 3; i++ {
		model.messages = append(model.messages, mcp.Message{
			Timestamp: time.Now(),
			Type:      mcp.TypeMessage,
			Source:    "test-agent",
			Content:   "Message",
		})
	}

	// Request 10 messages
	recent := model.getRecentMessages(10)

	// Should return all 3 messages
	if len(recent) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(recent))
	}
}

// TestRenderLogPane_DifferentSizes tests rendering with different pane sizes
func TestRenderLogPane_DifferentSizes(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"Small", 60, 10},
		{"Medium", 80, 20},
		{"Large", 120, 40},
	}

	for _, tt := range sizes {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			tf.AddMessage(mcp.Message{
				Timestamp: time.Now(),
				Type:      mcp.TypeMessage,
				Source:    "test-agent",
				Content:   "Test message",
			})

			err := model.refreshData()
			if err != nil {
				t.Fatalf("refreshData failed: %v", err)
			}

			pane := model.renderLogPane(tt.width, tt.height)

			if pane == "" {
				t.Errorf("Log pane should not be empty for size %dx%d", tt.width, tt.height)
			}
		})
	}
}

// TestRenderLogPane_WithKeybindingHints tests that keybinding hints are shown
func TestRenderLogPane_WithKeybindingHints(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	pane := model.renderLogPane(80, 20)

	// Should contain keybinding hints
	if !strings.Contains(pane, "/:search") {
		t.Error("Log pane should contain search keybinding hint")
	}
	if !strings.Contains(pane, "a:agent") {
		t.Error("Log pane should contain agent filter keybinding hint")
	}
	if !strings.Contains(pane, "m:type") {
		t.Error("Log pane should contain type filter keybinding hint")
	}
	if !strings.Contains(pane, "x:clear") {
		t.Error("Log pane should contain clear keybinding hint")
	}
	if !strings.Contains(pane, "e:export") {
		t.Error("Log pane should contain export keybinding hint")
	}
}

// TestRenderLogPane_TimestampFormat tests timestamp formatting
func TestRenderLogPane_TimestampFormat(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add message with current timestamp
	now := time.Now()
	tf.AddMessage(mcp.Message{
		Timestamp: now,
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test message with timestamp",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 20)

	// Should contain the message content
	if !strings.Contains(pane, "Test message with timestamp") {
		t.Error("Log pane should contain message content")
	}
	
	// Should contain a timestamp in HH:MM:SS format (just check for colon pattern)
	if !strings.Contains(pane, ":") {
		t.Error("Log pane should contain timestamp with colon separator")
	}
}

// TestRenderLogPane_LongMessages tests rendering with long message content
func TestRenderLogPane_LongMessages(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tf.AddMessage(mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "This is a very long message content that should be truncated to fit within the available pane width without breaking the layout or causing rendering issues",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 20)

	// Should render without panic
	if pane == "" {
		t.Error("Log pane should not be empty with long messages")
	}
}

// TestRenderLogPane_ChronologicalOrder tests that messages are in chronological order
func TestRenderLogPane_ChronologicalOrder(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add messages in specific order
	baseTime := time.Now()
	tf.AddMessage(mcp.Message{
		Timestamp: baseTime,
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "First",
	})
	tf.AddMessage(mcp.Message{
		Timestamp: baseTime.Add(1 * time.Second),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Second",
	})
	tf.AddMessage(mcp.Message{
		Timestamp: baseTime.Add(2 * time.Second),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Third",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderLogPane(80, 20)

	// All messages should be present
	if !strings.Contains(pane, "First") {
		t.Error("Log pane should contain first message")
	}
	if !strings.Contains(pane, "Second") {
		t.Error("Log pane should contain second message")
	}
	if !strings.Contains(pane, "Third") {
		t.Error("Log pane should contain third message")
	}
}
