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
	procManager process.ProcessManager

	// State
	agents   []mcp.AgentStatus
	tasks    []beads.Task
	messages []mcp.Message

	// UI state
	width       int
	height      int
	lastRefresh time.Time

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
		config:      cfg,
		beadsClient: beadsClient,
		mcpClient:   mcpClient,
		procManager: procManager,
		agents:      []mcp.AgentStatus{},
		tasks:       []beads.Task{},
		messages:    []mcp.Message{},
		lastRefresh: time.Now(),
	}
}

// tickMsg is sent periodically to trigger data refresh
type tickMsg time.Time

// Init initializes the TUI model and starts the ticker
func (m Model) Init() tea.Cmd {
	// Start periodic refresh ticker
	return tea.Batch(
		tickCmd(),
		refreshDataCmd(m),
	)
}

// tickCmd returns a command that sends a tick message after a delay
func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// GetError returns the current error state of the model
func (m Model) GetError() error {
	return m.err
}
