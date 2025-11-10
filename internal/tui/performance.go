package tui

import (
	"time"
)

// PerformanceMonitor monitors rendering performance
type PerformanceMonitor struct {
	frameCount    int
	lastFrameTime time.Time
	fps           float64
	frameTime     time.Duration
	targetFPS     int
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(targetFPS int) *PerformanceMonitor {
	return &PerformanceMonitor{
		frameCount:    0,
		lastFrameTime: time.Now(),
		fps:           0,
		frameTime:     0,
		targetFPS:     targetFPS,
	}
}

// StartFrame marks the start of a frame
func (pm *PerformanceMonitor) StartFrame() {
	pm.lastFrameTime = time.Now()
}

// EndFrame marks the end of a frame and calculates metrics
func (pm *PerformanceMonitor) EndFrame() {
	now := time.Now()
	pm.frameTime = now.Sub(pm.lastFrameTime)
	pm.frameCount++
	
	// Calculate FPS
	if pm.frameTime > 0 {
		pm.fps = 1.0 / pm.frameTime.Seconds()
	}
}

// GetFPS returns the current FPS
func (pm *PerformanceMonitor) GetFPS() float64 {
	return pm.fps
}

// GetFrameTime returns the last frame time
func (pm *PerformanceMonitor) GetFrameTime() time.Duration {
	return pm.frameTime
}

// ShouldSkipFrame returns true if we should skip rendering to maintain target FPS
func (pm *PerformanceMonitor) ShouldSkipFrame() bool {
	targetFrameTime := time.Second / time.Duration(pm.targetFPS)
	return pm.frameTime < targetFrameTime
}

// GetFrameCount returns the total frame count
func (pm *PerformanceMonitor) GetFrameCount() int {
	return pm.frameCount
}

// RenderCache caches rendered content to avoid re-rendering
type RenderCache struct {
	cache      map[string]CachedContent
	maxEntries int
}

// CachedContent represents cached rendered content
type CachedContent struct {
	Content   string
	Timestamp time.Time
	TTL       time.Duration
}

// NewRenderCache creates a new render cache
func NewRenderCache(maxEntries int) *RenderCache {
	return &RenderCache{
		cache:      make(map[string]CachedContent),
		maxEntries: maxEntries,
	}
}

// Get retrieves content from cache
func (rc *RenderCache) Get(key string) (string, bool) {
	cached, exists := rc.cache[key]
	if !exists {
		return "", false
	}
	
	// Check if expired
	if time.Since(cached.Timestamp) > cached.TTL {
		delete(rc.cache, key)
		return "", false
	}
	
	return cached.Content, true
}

// Set stores content in cache
func (rc *RenderCache) Set(key, content string, ttl time.Duration) {
	// Evict oldest entry if cache is full
	if len(rc.cache) >= rc.maxEntries {
		rc.evictOldest()
	}
	
	rc.cache[key] = CachedContent{
		Content:   content,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// Invalidate removes an entry from cache
func (rc *RenderCache) Invalidate(key string) {
	delete(rc.cache, key)
}

// Clear clears the entire cache
func (rc *RenderCache) Clear() {
	rc.cache = make(map[string]CachedContent)
}

// evictOldest removes the oldest entry from cache
func (rc *RenderCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, cached := range rc.cache {
		if oldestKey == "" || cached.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.Timestamp
		}
	}
	
	if oldestKey != "" {
		delete(rc.cache, oldestKey)
	}
}

// DirtyTracker tracks which parts of the UI need re-rendering
type DirtyTracker struct {
	dirtyPanes map[string]bool
}

// NewDirtyTracker creates a new dirty tracker
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		dirtyPanes: make(map[string]bool),
	}
}

// MarkDirty marks a pane as needing re-render
func (dt *DirtyTracker) MarkDirty(pane string) {
	dt.dirtyPanes[pane] = true
}

// IsDirty checks if a pane needs re-rendering
func (dt *DirtyTracker) IsDirty(pane string) bool {
	return dt.dirtyPanes[pane]
}

// ClearDirty marks a pane as clean
func (dt *DirtyTracker) ClearDirty(pane string) {
	delete(dt.dirtyPanes, pane)
}

// ClearAll marks all panes as clean
func (dt *DirtyTracker) ClearAll() {
	dt.dirtyPanes = make(map[string]bool)
}

// GetDirtyPanes returns all dirty panes
func (dt *DirtyTracker) GetDirtyPanes() []string {
	var panes []string
	for pane := range dt.dirtyPanes {
		panes = append(panes, pane)
	}
	return panes
}

// OptimizeRendering optimizes rendering by skipping unchanged content
func OptimizeRendering(oldContent, newContent string) bool {
	// Return true if content has changed
	return oldContent != newContent
}

// BatchUpdate batches multiple updates together
type BatchUpdate struct {
	updates []func()
	pending bool
}

// NewBatchUpdate creates a new batch update
func NewBatchUpdate() *BatchUpdate {
	return &BatchUpdate{
		updates: []func(){},
		pending: false,
	}
}

// Add adds an update to the batch
func (bu *BatchUpdate) Add(update func()) {
	bu.updates = append(bu.updates, update)
	bu.pending = true
}

// Execute executes all batched updates
func (bu *BatchUpdate) Execute() {
	if !bu.pending {
		return
	}
	
	for _, update := range bu.updates {
		update()
	}
	
	bu.updates = []func(){}
	bu.pending = false
}

// HasPending returns true if there are pending updates
func (bu *BatchUpdate) HasPending() bool {
	return bu.pending
}

// Throttle throttles function calls to a maximum rate
type Throttle struct {
	lastCall time.Time
	interval time.Duration
}

// NewThrottle creates a new throttle
func NewThrottle(interval time.Duration) *Throttle {
	return &Throttle{
		lastCall: time.Time{},
		interval: interval,
	}
}

// ShouldCall returns true if enough time has passed since last call
func (t *Throttle) ShouldCall() bool {
	now := time.Now()
	if now.Sub(t.lastCall) >= t.interval {
		t.lastCall = now
		return true
	}
	return false
}

// Call calls the function if throttle allows
func (t *Throttle) Call(fn func()) {
	if t.ShouldCall() {
		fn()
	}
}

// Debounce debounces function calls
type Debounce struct {
	timer    *time.Timer
	interval time.Duration
}

// NewDebounce creates a new debounce
func NewDebounce(interval time.Duration) *Debounce {
	return &Debounce{
		timer:    nil,
		interval: interval,
	}
}

// Call debounces the function call
func (d *Debounce) Call(fn func()) {
	if d.timer != nil {
		d.timer.Stop()
	}
	
	d.timer = time.AfterFunc(d.interval, fn)
}

// Cancel cancels any pending debounced call
func (d *Debounce) Cancel() {
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

// MicroInteraction represents a subtle UI interaction
type MicroInteraction struct {
	Type      string
	StartTime time.Time
	Duration  time.Duration
	Progress  float64
}

// NewMicroInteraction creates a new micro-interaction
func NewMicroInteraction(interactionType string, duration time.Duration) *MicroInteraction {
	return &MicroInteraction{
		Type:      interactionType,
		StartTime: time.Now(),
		Duration:  duration,
		Progress:  0,
	}
}

// Update updates the micro-interaction progress
func (mi *MicroInteraction) Update() {
	elapsed := time.Since(mi.StartTime)
	mi.Progress = float64(elapsed) / float64(mi.Duration)
	
	if mi.Progress > 1.0 {
		mi.Progress = 1.0
	}
}

// IsComplete returns true if the interaction is complete
func (mi *MicroInteraction) IsComplete() bool {
	return mi.Progress >= 1.0
}

// GetEasedProgress returns progress with easing applied
func (mi *MicroInteraction) GetEasedProgress() float64 {
	return EaseInOutCubic(mi.Progress)
}
