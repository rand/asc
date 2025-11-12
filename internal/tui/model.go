package tui

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/health"
	"github.com/rand/asc/internal/logger"
	"github.com/rand/asc/internal/mcp"
	"github.com/rand/asc/internal/process"
)

// Model represents the TUI application state
type Model struct {
	// Configuration
	config        config.Config
	configWatcher *config.Watcher       // Watches for config file changes
	reloadManager *config.ReloadManager // Manages config reload and agent lifecycle

	// Data sources
	beadsClient   beads.BeadsClient
	mcpClient     mcp.MCPClient
	wsClient      *mcp.WebSocketClient // WebSocket client for real-time updates
	procManager   process.ProcessManager
	healthMonitor *health.Monitor // Health monitoring system
	logAggregator *logger.LogAggregator // Log aggregation system

	// State
	agents       []mcp.AgentStatus
	tasks        []beads.Task
	messages     []mcp.Message
	healthIssues []health.HealthIssue

	// UI state
	width         int
	height        int
	lastRefresh   time.Time
	wsConnected   bool // WebSocket connection status
	beadsConnected bool // Beads connection status

	// Task interaction state
	selectedTaskIndex int    // Index of selected task in filtered list
	showTaskModal     bool   // Whether to show task detail modal
	showCreateModal   bool   // Whether to show create task modal
	createTaskInput   string // Input for new task title

	// Agent interaction state
	selectedAgentIndex int  // Index of selected agent (1-9)
	showConfirmModal   bool // Whether to show confirmation dialog
	confirmAction      string // Action to confirm (kill, restart)

	// Log filtering state
	searchMode      bool   // Whether in search mode
	searchInput     string // Search input text
	logFilterAgent  string // Filter logs by agent name
	logFilterType   string // Filter logs by message type

	// Reload notification state
	reloadNotification string    // Message to display for config reload
	reloadNotificationTime time.Time // When the notification was shown

	// Debug mode
	debugMode bool // Whether debug mode is enabled

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
	// Initialize log aggregator
	homeDir, _ := os.UserHomeDir()
	logsDir := filepath.Join(homeDir, ".asc", "logs")
	logAggregator := logger.NewLogAggregator(logsDir, 1000) // Keep last 1000 entries

	return Model{
		config:         cfg,
		configWatcher:  nil, // Will be initialized in Init
		reloadManager:  nil, // Will be initialized in Init
		beadsClient:    beadsClient,
		mcpClient:      mcpClient,
		wsClient:       nil, // Will be initialized in Init if WebSocket URL is available
		procManager:    procManager,
		healthMonitor:  nil, // Will be initialized in Init
		logAggregator:  logAggregator,
		agents:         []mcp.AgentStatus{},
		tasks:          []beads.Task{},
		messages:       []mcp.Message{},
		healthIssues:   []health.HealthIssue{},
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

	// Initialize health monitor
	if monitor, err := health.NewMonitor(m.mcpClient, m.procManager, m.config); err == nil {
		m.healthMonitor = monitor
		// Apply auto-recovery configuration from config file
		// If not specified (nil), it defaults to true (enabled)
		if m.config.Core.AutoRecovery != nil {
			m.healthMonitor.SetAutoRecovery(*m.config.Core.AutoRecovery)
		}
		// Otherwise, keep the default (true) from NewMonitor
		m.healthMonitor.Start()
	}

	// Initialize configuration hot-reload
	if watcher, err := config.NewWatcher(config.DefaultConfigPath()); err == nil {
		m.configWatcher = watcher
		
		// Create reload manager with environment variables
		envVars := m.getEnvVars()
		// Wrap the process manager to adapt the interface
		adaptedProcManager := newProcessManagerAdapter(m.procManager)
		m.reloadManager = config.NewReloadManager(&m.config, adaptedProcManager, envVars)
		
		// Register reload callback
		m.configWatcher.OnReload(func(newConfig *config.Config) error {
			// This will be called in a goroutine, so we need to send a message to the TUI
			// We'll handle the actual reload in the Update function
			return nil
		})
		
		// Start watching
		if err := m.configWatcher.Start(); err == nil {
			// Start listening for config reload events
			cmds = append(cmds, waitForConfigReloadCmd(m.configWatcher))
		}
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

// SetDebugMode enables or disables debug mode in the TUI
func (m *Model) SetDebugMode(debug bool) {
	m.debugMode = debug
}

// Cleanup closes any open connections and performs cleanup
func (m *Model) Cleanup() {
	if m.wsClient != nil {
		m.wsClient.Close()
	}
	if m.healthMonitor != nil {
		m.healthMonitor.Stop()
	}
	if m.configWatcher != nil {
		m.configWatcher.Stop()
	}
}

// getEnvVars returns environment variables needed for agents (API keys, etc.)
func (m Model) getEnvVars() map[string]string {
	envVars := make(map[string]string)
	
	// Get API keys from environment
	apiKeys := []string{
		"CLAUDE_API_KEY",
		"OPENAI_API_KEY",
		"GOOGLE_API_KEY",
	}
	
	for _, key := range apiKeys {
		if value := os.Getenv(key); value != "" {
			envVars[key] = value
		}
	}
	
	return envVars
}

// configReloadMsg is sent when the configuration file changes
type configReloadMsg struct {
	newConfig *config.Config
}

// waitForConfigReloadCmd waits for configuration reload events
func waitForConfigReloadCmd(watcher *config.Watcher) tea.Cmd {
	return func() tea.Msg {
		newConfig := <-watcher.Events()
		return configReloadMsg{newConfig: newConfig}
	}
}
