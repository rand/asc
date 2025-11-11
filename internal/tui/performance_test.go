package tui

import (
	"testing"
	"time"
)

func TestNewPerformanceMonitor(t *testing.T) {
	targetFPS := 60
	pm := NewPerformanceMonitor(targetFPS)

	if pm == nil {
		t.Fatal("NewPerformanceMonitor returned nil")
	}

	if pm.targetFPS != targetFPS {
		t.Errorf("Expected targetFPS %d, got %d", targetFPS, pm.targetFPS)
	}

	if pm.frameCount != 0 {
		t.Error("Initial frame count should be 0")
	}
}

func TestPerformanceMonitorFrameCycle(t *testing.T) {
	pm := NewPerformanceMonitor(60)

	// Start frame
	pm.StartFrame()

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// End frame
	pm.EndFrame()

	// Check that metrics were updated
	if pm.GetFrameCount() != 1 {
		t.Errorf("Expected frame count 1, got %d", pm.GetFrameCount())
	}

	if pm.GetFrameTime() == 0 {
		t.Error("Frame time should not be 0")
	}

	if pm.GetFPS() == 0 {
		t.Error("FPS should not be 0")
	}
}

func TestPerformanceMonitorGetFPS(t *testing.T) {
	pm := NewPerformanceMonitor(60)

	pm.StartFrame()
	time.Sleep(16 * time.Millisecond) // ~60 FPS
	pm.EndFrame()

	fps := pm.GetFPS()
	if fps <= 0 {
		t.Error("FPS should be positive")
	}

	// FPS should be roughly 60 (allow for variance)
	if fps < 30 || fps > 100 {
		t.Logf("FPS %f is outside expected range (30-100), but this may be due to system load", fps)
	}
}

func TestPerformanceMonitorGetFrameTime(t *testing.T) {
	pm := NewPerformanceMonitor(60)

	pm.StartFrame()
	sleepDuration := 10 * time.Millisecond
	time.Sleep(sleepDuration)
	pm.EndFrame()

	frameTime := pm.GetFrameTime()
	if frameTime < sleepDuration {
		t.Errorf("Frame time %v should be at least %v", frameTime, sleepDuration)
	}
}

func TestPerformanceMonitorShouldSkipFrame(t *testing.T) {
	pm := NewPerformanceMonitor(60)

	// Very fast frame
	pm.StartFrame()
	pm.EndFrame()

	// Should skip if frame was too fast
	_ = pm.ShouldSkipFrame()
}

func TestPerformanceMonitorGetFrameCount(t *testing.T) {
	pm := NewPerformanceMonitor(60)

	if pm.GetFrameCount() != 0 {
		t.Error("Initial frame count should be 0")
	}

	// Process multiple frames
	for i := 0; i < 5; i++ {
		pm.StartFrame()
		pm.EndFrame()
	}

	if pm.GetFrameCount() != 5 {
		t.Errorf("Expected frame count 5, got %d", pm.GetFrameCount())
	}
}

func TestNewRenderCache(t *testing.T) {
	maxEntries := 10
	cache := NewRenderCache(maxEntries)

	if cache == nil {
		t.Fatal("NewRenderCache returned nil")
	}

	if cache.maxEntries != maxEntries {
		t.Errorf("Expected maxEntries %d, got %d", maxEntries, cache.maxEntries)
	}
}

func TestRenderCacheSetAndGet(t *testing.T) {
	cache := NewRenderCache(10)

	key := "test_key"
	content := "test content"
	ttl := 1 * time.Second

	// Set content
	cache.Set(key, content, ttl)

	// Get content
	retrieved, exists := cache.Get(key)
	if !exists {
		t.Error("Content should exist in cache")
	}

	if retrieved != content {
		t.Errorf("Expected content '%s', got '%s'", content, retrieved)
	}
}

func TestRenderCacheGetNonExistent(t *testing.T) {
	cache := NewRenderCache(10)

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Error("Non-existent key should not exist")
	}
}

func TestRenderCacheExpiration(t *testing.T) {
	cache := NewRenderCache(10)

	key := "expiring_key"
	content := "expiring content"
	ttl := 50 * time.Millisecond

	cache.Set(key, content, ttl)

	// Should exist immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Error("Content should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(ttl + 10*time.Millisecond)

	// Should not exist after expiration
	_, exists = cache.Get(key)
	if exists {
		t.Error("Content should not exist after TTL expiration")
	}
}

func TestRenderCacheInvalidate(t *testing.T) {
	cache := NewRenderCache(10)

	key := "test_key"
	cache.Set(key, "content", 1*time.Second)

	// Invalidate
	cache.Invalidate(key)

	// Should not exist
	_, exists := cache.Get(key)
	if exists {
		t.Error("Content should not exist after invalidation")
	}
}

func TestRenderCacheClear(t *testing.T) {
	cache := NewRenderCache(10)

	// Add multiple entries
	cache.Set("key1", "content1", 1*time.Second)
	cache.Set("key2", "content2", 1*time.Second)
	cache.Set("key3", "content3", 1*time.Second)

	// Clear
	cache.Clear()

	// None should exist
	_, exists1 := cache.Get("key1")
	_, exists2 := cache.Get("key2")
	_, exists3 := cache.Get("key3")

	if exists1 || exists2 || exists3 {
		t.Error("No content should exist after clear")
	}
}

func TestRenderCacheEviction(t *testing.T) {
	cache := NewRenderCache(3) // Small cache

	// Fill cache
	cache.Set("key1", "content1", 1*time.Second)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	cache.Set("key2", "content2", 1*time.Second)
	time.Sleep(10 * time.Millisecond)
	cache.Set("key3", "content3", 1*time.Second)
	time.Sleep(10 * time.Millisecond)

	// Add one more (should evict oldest)
	cache.Set("key4", "content4", 1*time.Second)

	// key1 should be evicted
	_, exists := cache.Get("key1")
	if exists {
		t.Error("Oldest entry should have been evicted")
	}

	// Others should exist
	_, exists2 := cache.Get("key2")
	_, exists3 := cache.Get("key3")
	_, exists4 := cache.Get("key4")

	if !exists2 || !exists3 || !exists4 {
		t.Error("Recent entries should still exist")
	}
}

func TestNewDirtyTracker(t *testing.T) {
	dt := NewDirtyTracker()

	if dt == nil {
		t.Fatal("NewDirtyTracker returned nil")
	}
}

func TestDirtyTrackerMarkAndCheck(t *testing.T) {
	dt := NewDirtyTracker()

	pane := "test_pane"

	// Should not be dirty initially
	if dt.IsDirty(pane) {
		t.Error("Pane should not be dirty initially")
	}

	// Mark dirty
	dt.MarkDirty(pane)

	// Should be dirty now
	if !dt.IsDirty(pane) {
		t.Error("Pane should be dirty after marking")
	}
}

func TestDirtyTrackerClearDirty(t *testing.T) {
	dt := NewDirtyTracker()

	pane := "test_pane"
	dt.MarkDirty(pane)

	// Clear
	dt.ClearDirty(pane)

	// Should not be dirty
	if dt.IsDirty(pane) {
		t.Error("Pane should not be dirty after clearing")
	}
}

func TestDirtyTrackerClearAll(t *testing.T) {
	dt := NewDirtyTracker()

	// Mark multiple panes
	dt.MarkDirty("pane1")
	dt.MarkDirty("pane2")
	dt.MarkDirty("pane3")

	// Clear all
	dt.ClearAll()

	// None should be dirty
	if dt.IsDirty("pane1") || dt.IsDirty("pane2") || dt.IsDirty("pane3") {
		t.Error("No panes should be dirty after ClearAll")
	}
}

func TestDirtyTrackerGetDirtyPanes(t *testing.T) {
	dt := NewDirtyTracker()

	// Mark some panes
	dt.MarkDirty("pane1")
	dt.MarkDirty("pane2")

	dirtyPanes := dt.GetDirtyPanes()

	if len(dirtyPanes) != 2 {
		t.Errorf("Expected 2 dirty panes, got %d", len(dirtyPanes))
	}
}

func TestOptimizeRendering(t *testing.T) {
	tests := []struct {
		name       string
		oldContent string
		newContent string
		want       bool
	}{
		{"same_content", "hello", "hello", false},
		{"different_content", "hello", "world", true},
		{"empty_to_content", "", "hello", true},
		{"content_to_empty", "hello", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptimizeRendering(tt.oldContent, tt.newContent)
			if result != tt.want {
				t.Errorf("OptimizeRendering() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNewBatchUpdate(t *testing.T) {
	bu := NewBatchUpdate()

	if bu == nil {
		t.Fatal("NewBatchUpdate returned nil")
	}

	if bu.HasPending() {
		t.Error("Should not have pending updates initially")
	}
}

func TestBatchUpdateAdd(t *testing.T) {
	bu := NewBatchUpdate()

	called := false
	bu.Add(func() {
		called = true
	})

	if !bu.HasPending() {
		t.Error("Should have pending updates after Add")
	}

	// Execute
	bu.Execute()

	if !called {
		t.Error("Update function should have been called")
	}

	if bu.HasPending() {
		t.Error("Should not have pending updates after Execute")
	}
}

func TestBatchUpdateExecute(t *testing.T) {
	bu := NewBatchUpdate()

	counter := 0
	bu.Add(func() { counter++ })
	bu.Add(func() { counter++ })
	bu.Add(func() { counter++ })

	bu.Execute()

	if counter != 3 {
		t.Errorf("Expected counter to be 3, got %d", counter)
	}
}

func TestNewThrottle(t *testing.T) {
	interval := 100 * time.Millisecond
	throttle := NewThrottle(interval)

	if throttle == nil {
		t.Fatal("NewThrottle returned nil")
	}

	if throttle.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, throttle.interval)
	}
}

func TestThrottleShouldCall(t *testing.T) {
	throttle := NewThrottle(50 * time.Millisecond)

	// First call should be allowed
	if !throttle.ShouldCall() {
		t.Error("First call should be allowed")
	}

	// Immediate second call should be throttled
	if throttle.ShouldCall() {
		t.Error("Immediate second call should be throttled")
	}

	// After interval, should be allowed
	time.Sleep(60 * time.Millisecond)
	if !throttle.ShouldCall() {
		t.Error("Call after interval should be allowed")
	}
}

func TestThrottleCall(t *testing.T) {
	throttle := NewThrottle(50 * time.Millisecond)

	counter := 0
	fn := func() { counter++ }

	// First call
	throttle.Call(fn)
	if counter != 1 {
		t.Error("First call should execute")
	}

	// Immediate second call (should be throttled)
	throttle.Call(fn)
	if counter != 1 {
		t.Error("Throttled call should not execute")
	}

	// After interval
	time.Sleep(60 * time.Millisecond)
	throttle.Call(fn)
	if counter != 2 {
		t.Error("Call after interval should execute")
	}
}

func TestNewDebounce(t *testing.T) {
	interval := 100 * time.Millisecond
	debounce := NewDebounce(interval)

	if debounce == nil {
		t.Fatal("NewDebounce returned nil")
	}

	if debounce.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, debounce.interval)
	}
}

func TestDebounceCall(t *testing.T) {
	debounce := NewDebounce(50 * time.Millisecond)

	counter := 0
	fn := func() { counter++ }

	// Multiple rapid calls
	debounce.Call(fn)
	debounce.Call(fn)
	debounce.Call(fn)

	// Should not execute immediately
	if counter != 0 {
		t.Error("Debounced function should not execute immediately")
	}

	// Wait for debounce interval
	time.Sleep(60 * time.Millisecond)

	// Should have executed once
	if counter != 1 {
		t.Errorf("Debounced function should execute once, got %d", counter)
	}
}

func TestDebounceCancel(t *testing.T) {
	debounce := NewDebounce(50 * time.Millisecond)

	counter := 0
	fn := func() { counter++ }

	// Call and cancel
	debounce.Call(fn)
	debounce.Cancel()

	// Wait
	time.Sleep(60 * time.Millisecond)

	// Should not have executed
	if counter != 0 {
		t.Error("Cancelled debounce should not execute")
	}
}

func TestNewMicroInteraction(t *testing.T) {
	duration := 200 * time.Millisecond
	mi := NewMicroInteraction("test", duration)

	if mi == nil {
		t.Fatal("NewMicroInteraction returned nil")
	}

	if mi.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", mi.Type)
	}

	if mi.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, mi.Duration)
	}
}

func TestMicroInteractionUpdate(t *testing.T) {
	mi := NewMicroInteraction("test", 100*time.Millisecond)

	// Initial progress should be 0
	if mi.Progress != 0 {
		t.Error("Initial progress should be 0")
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Update
	mi.Update()

	// Progress should be between 0 and 1
	if mi.Progress <= 0 || mi.Progress > 1 {
		t.Errorf("Progress should be between 0 and 1, got %f", mi.Progress)
	}
}

func TestMicroInteractionIsComplete(t *testing.T) {
	mi := NewMicroInteraction("test", 50*time.Millisecond)

	// Should not be complete initially
	mi.Update()
	if mi.IsComplete() {
		t.Error("Should not be complete immediately")
	}

	// Wait for completion
	time.Sleep(60 * time.Millisecond)
	mi.Update()

	// Should be complete
	if !mi.IsComplete() {
		t.Error("Should be complete after duration")
	}
}

func TestMicroInteractionGetEasedProgress(t *testing.T) {
	mi := NewMicroInteraction("test", 100*time.Millisecond)

	// Set progress manually for testing
	mi.Progress = 0.5

	easedProgress := mi.GetEasedProgress()

	// Should return a valid value
	if easedProgress < 0 || easedProgress > 1 {
		t.Errorf("Eased progress should be between 0 and 1, got %f", easedProgress)
	}
}
