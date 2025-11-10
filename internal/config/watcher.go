package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a configuration file for changes and triggers reload callbacks
type Watcher struct {
	configPath string
	watcher    *fsnotify.Watcher
	callbacks  []ReloadCallback
	mu         sync.RWMutex
	stopCh     chan struct{}
	eventCh    chan *Config // Channel for sending reload events
	running    bool
}

// ReloadCallback is called when the configuration file changes
// It receives the new configuration and returns an error if reload fails
type ReloadCallback func(newConfig *Config) error

// NewWatcher creates a new configuration file watcher
func NewWatcher(configPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	w := &Watcher{
		configPath: configPath,
		watcher:    fsWatcher,
		callbacks:  []ReloadCallback{},
		stopCh:     make(chan struct{}),
		eventCh:    make(chan *Config, 10), // Buffered channel for reload events
		running:    false,
	}

	return w, nil
}

// Events returns the channel for receiving reload events
func (w *Watcher) Events() <-chan *Config {
	return w.eventCh
}

// OnReload registers a callback to be called when the configuration changes
func (w *Watcher) OnReload(callback ReloadCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Start begins watching the configuration file for changes
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	w.mu.Unlock()

	// Add the config file to the watcher
	if err := w.watcher.Add(w.configPath); err != nil {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// Start the watch loop in a goroutine
	go w.watchLoop()

	return nil
}

// Stop stops watching the configuration file
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopCh)
	w.watcher.Close()
	w.running = false
}

// watchLoop is the main event loop that processes file system events
func (w *Watcher) watchLoop() {
	// Debounce timer to avoid multiple reloads for rapid file changes
	var debounceTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// We're interested in Write and Create events
			// Some editors create a new file and rename it, so we watch for both
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Debounce: reset timer on each event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(debounceDuration, func() {
					w.handleConfigChange()
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			fmt.Printf("watcher error: %v\n", err)

		case <-w.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return
		}
	}
}

// handleConfigChange loads the new configuration and calls all registered callbacks
func (w *Watcher) handleConfigChange() {
	// Load the new configuration
	newConfig, err := Load(w.configPath)
	if err != nil {
		// Log error but don't stop watching
		fmt.Printf("failed to reload config: %v\n", err)
		return
	}

	// Send event to channel (non-blocking)
	select {
	case w.eventCh <- newConfig:
	default:
		// Channel full, skip this event
	}

	// Call all registered callbacks
	w.mu.RLock()
	callbacks := make([]ReloadCallback, len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(newConfig); err != nil {
			// Log error but continue with other callbacks
			fmt.Printf("reload callback error: %v\n", err)
		}
	}
}
