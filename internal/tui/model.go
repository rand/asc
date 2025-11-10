package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/beads"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/mcp"
	"github.com/yourusername/asc/internal/process"
)

// Model represents the TUI application state
type Model struct {
	// Configuration
	config config.Config

	// Data sources
	beadsClient beads.BeadsClient
	mcpClient   mcp.MCPClient
	wsClient    *mcp.WebSocketClient // WebSocket client for real-time updates
	procManager process.ProcessManager

	// State
	agents   []mcp.AgentStatus
	tasks    []beads.Task
	messages []mcp.Message

	// UI state
	width         int
	height        int
	lastRefresh   time.Time
	wsConnected   bool // WebSocket connection status
	beadsConnected bool // Beads connection status

	// Error state
	err error
}

// NewModel creates a new TUI model with the given configuration and clients
func NewModel(
	cfg config.Config,
	beadsClient beads.BeadsClient,
	mcpClient mcp.MCPClient,
	procManager process.ProcessManager,
) Model {
	return Model{
		config:         cfg,
		beadsClient:    beadsClient,
		mcpClient:      mcpClient,
		wsClient:       nil, // Will be initialized in Init if WebSocket URL is available
		procManager:    procManager,
		agents:         []mcp.AgentStatus{},
		tasks:          []beads.Task{},
		messages:       []mcp.Message{},
		lastRefresh:    time.Now(),
		wsConnected:    false,
		beadsConnected: false,
	}
}

// tickMsg is sent periodically to trigger data refresh (for beads polling)
type tickMsg time.Time

// wsEventMsg wraps a WebSocket event for the TUI
type wsEventMsg mcp.Event

// Init initializes the TUI model and starts the ticker
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		refreshDataCmd(m), // Initial data load
	}

	// Try to initialize WebSocket connection for real-time MCP updates
	if m.config.Services.MCPAgentMail.URL != "" {
		// Convert HTTP URL to WebSocket URL
		wsURL := convertToWebSocketURL(m.config.Services.MCPAgentMail.URL)
		m.wsClient = mcp.NewWebSocketClient(wsURL)
		
		// Attempt to connect (non-blocking)
		cmds = append(cmds, connectWebSocketCmd(m.wsClient))
		
		// Start listening for WebSocket events
		cmds = append(cmds, waitForWSEventCmd(m.wsClient))
	}

	// Start periodic refresh ticker for beads (git-based, cannot be real-time)
	cmds = append(cmds, tickCmd())

	return tea.Batch(cmds...)
}

// tickCmd returns a command that sends a tick message after a delay
// This is used for polling beads, which is git-based and cannot be real-time
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// connectWebSocketCmd attempts to connect the WebSocket client
func connectWebSocketCmd(wsClient *mcp.WebSocketClient) tea.Cmd {
	return func() tea.Msg {
		if err := wsClient.Connect(); err != nil {
			// Connection failed, but we'll continue with polling fallback
			return wsEventMsg{
				Type:  mcp.EventError,
				Error: err.Error(),
			}
		}
		return nil
	}
}

// waitForWSEventCmd waits for the next WebSocket event
func waitForWSEventCmd(wsClient *mcp.WebSocketClient) tea.Cmd {
	return func() tea.Msg {
		event := <-wsClient.Events()
		return wsEventMsg(event)
	}
}

// convertToWebSocketURL converts an HTTP URL to a WebSocket URL
func convertToWebSocketURL(httpURL string) string {
	// Replace http:// with ws:// and https:// with wss://
	if len(httpURL) > 7 && httpURL[:7] == "http://" {
		return "ws://" + httpURL[7:] + "/ws"
	} else if len(httpURL) > 8 && httpURL[:8] == "https://" {
		return "wss://" + httpURL[8:] + "/ws"
	}
	return httpURL + "/ws"
}

// GetError returns the current error state of the model
func (m Model) GetError() error {
	return m.err
}

// Cleanup closes any open connections and performs cleanup
func (m *Model) Cleanup() {
	if m.wsClient != nil {
		m.wsClient.Close()
	}
}
