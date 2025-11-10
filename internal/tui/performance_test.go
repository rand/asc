package tui

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkPerformanceMonitor benchmarks performance monitoring operations
func BenchmarkPerformanceMonitor(b *testing.B) {
	pm := NewPerformanceMonitor(60)
	
	b.Run("StartEndFrame", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pm.StartFrame()
			pm.EndFrame()
		}
	})
	
	b.Run("GetFPS", func(b *testing.B) {
		pm.StartFrame()
		time.Sleep(time.Millisecond)
		pm.EndFrame()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pm.GetFPS()
		}
	})
	
	b.Run("ShouldSkipFrame", func(b *testing.B) {
		pm.StartFrame()
		time.Sleep(time.Millisecond)
		pm.EndFrame()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pm.ShouldSkipFrame()
		}
	})
}

// BenchmarkRenderCache benchmarks render cache operations
func BenchmarkRenderCache(b *testing.B) {
	cache := NewRenderCache(1000)
	content := "This is sample rendered content that would be cached"
	
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key-%d", i), content, time.Minute)
		}
	})
	
	b.Run("Get-Hit", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("key-%d", i), content, time.Minute)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(fmt.Sprintf("key-%d", i%100))
		}
	})
	
	b.Run("Get-Miss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(fmt.Sprintf("missing-%d", i))
		}
	})
	
	b.Run("Invalidate", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("key-%d", i), content, time.Minute)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Invalidate(fmt.Sprintf("key-%d", i%100))
		}
	})
	
	b.Run("Clear", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Pre-populate cache
			for j := 0; j < 100; j++ {
				cache.Set(fmt.Sprintf("key-%d", j), content, time.Minute)
			}
			b.StartTimer()
			
			cache.Clear()
		}
	})
}

// BenchmarkRenderCacheEviction benchmarks cache eviction
func BenchmarkRenderCacheEviction(b *testing.B) {
	cache := NewRenderCache(100)
	content := "cached content"
	
	// Fill cache to capacity
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), content, time.Minute)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This should trigger eviction
		cache.Set(fmt.Sprintf("new-key-%d", i), content, time.Minute)
	}
}

// BenchmarkDirtyTracker benchmarks dirty tracking operations
func BenchmarkDirtyTracker(b *testing.B) {
	tracker := NewDirtyTracker()
	panes := []string{"agents", "tasks", "logs", "footer", "header"}
	
	b.Run("MarkDirty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tracker.MarkDirty(panes[i%len(panes)])
		}
	})
	
	b.Run("IsDirty", func(b *testing.B) {
		for _, pane := range panes {
			tracker.MarkDirty(pane)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tracker.IsDirty(panes[i%len(panes)])
		}
	})
	
	b.Run("ClearDirty", func(b *testing.B) {
		for _, pane := range panes {
			tracker.MarkDirty(pane)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tracker.ClearDirty(panes[i%len(panes)])
		}
	})
	
	b.Run("GetDirtyPanes", func(b *testing.B) {
		for _, pane := range panes {
			tracker.MarkDirty(pane)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tracker.GetDirtyPanes()
		}
	})
	
	b.Run("ClearAll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			for _, pane := range panes {
				tracker.MarkDirty(pane)
			}
			b.StartTimer()
			
			tracker.ClearAll()
		}
	})
}

// BenchmarkBatchUpdate benchmarks batch update operations
func BenchmarkBatchUpdate(b *testing.B) {
	bu := NewBatchUpdate()
	counter := 0
	update := func() { counter++ }
	
	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bu.Add(update)
		}
	})
	
	b.Run("Execute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bu = NewBatchUpdate()
			for j := 0; j < 10; j++ {
				bu.Add(update)
			}
			b.StartTimer()
			
			bu.Execute()
		}
	})
	
	b.Run("HasPending", func(b *testing.B) {
		bu.Add(update)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bu.HasPending()
		}
	})
}

// BenchmarkThrottle benchmarks throttle operations
func BenchmarkThrottle(b *testing.B) {
	throttle := NewThrottle(10 * time.Millisecond)
	counter := 0
	fn := func() { counter++ }
	
	b.Run("ShouldCall", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = throttle.ShouldCall()
		}
	})
	
	b.Run("Call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			throttle.Call(fn)
		}
	})
}

// BenchmarkDebounce benchmarks debounce operations
func BenchmarkDebounce(b *testing.B) {
	debounce := NewDebounce(10 * time.Millisecond)
	counter := 0
	fn := func() { counter++ }
	
	b.Run("Call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			debounce.Call(fn)
		}
	})
	
	b.Run("Cancel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			debounce.Call(fn)
			b.StartTimer()
			
			debounce.Cancel()
		}
	})
}

// BenchmarkMicroInteraction benchmarks micro-interaction operations
func BenchmarkMicroInteraction(b *testing.B) {
	mi := NewMicroInteraction("test", 100*time.Millisecond)
	
	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mi.Update()
		}
	})
	
	b.Run("IsComplete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mi.IsComplete()
		}
	})
	
	b.Run("GetEasedProgress", func(b *testing.B) {
		mi.Progress = 0.5
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mi.GetEasedProgress()
		}
	})
}

// BenchmarkOptimizeRendering benchmarks rendering optimization
func BenchmarkOptimizeRendering(b *testing.B) {
	oldContent := "This is the old content that was previously rendered"
	newContent := "This is the new content that needs to be rendered"
	
	b.Run("Changed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = OptimizeRendering(oldContent, newContent)
		}
	})
	
	b.Run("Unchanged", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = OptimizeRendering(oldContent, oldContent)
		}
	})
}

// TestPerformanceMonitorAccuracy tests performance monitor accuracy
func TestPerformanceMonitorAccuracy(t *testing.T) {
	pm := NewPerformanceMonitor(60)
	
	// Simulate frames at ~60 FPS
	for i := 0; i < 10; i++ {
		pm.StartFrame()
		time.Sleep(16 * time.Millisecond) // ~60 FPS
		pm.EndFrame()
	}
	
	fps := pm.GetFPS()
	if fps < 50 || fps > 70 {
		t.Errorf("Expected FPS around 60, got %.2f", fps)
	}
	
	frameTime := pm.GetFrameTime()
	if frameTime < 15*time.Millisecond || frameTime > 20*time.Millisecond {
		t.Errorf("Expected frame time around 16ms, got %v", frameTime)
	}
}

// TestRenderCacheExpiration tests cache expiration
func TestRenderCacheExpiration(t *testing.T) {
	cache := NewRenderCache(10)
	
	cache.Set("key1", "content1", 50*time.Millisecond)
	
	// Should be in cache
	content, ok := cache.Get("key1")
	if !ok || content != "content1" {
		t.Error("Expected content to be in cache")
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Should be expired
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expected content to be expired")
	}
}

// TestRenderCacheEviction tests cache eviction policy
func TestRenderCacheEviction(t *testing.T) {
	cache := NewRenderCache(3)
	
	cache.Set("key1", "content1", time.Minute)
	time.Sleep(10 * time.Millisecond)
	cache.Set("key2", "content2", time.Minute)
	time.Sleep(10 * time.Millisecond)
	cache.Set("key3", "content3", time.Minute)
	
	// Cache is full, adding one more should evict oldest (key1)
	cache.Set("key4", "content4", time.Minute)
	
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be evicted")
	}
	
	_, ok = cache.Get("key2")
	if !ok {
		t.Error("Expected key2 to still be in cache")
	}
}

// TestThrottleRateLimit tests throttle rate limiting
func TestThrottleRateLimit(t *testing.T) {
	throttle := NewThrottle(50 * time.Millisecond)
	counter := 0
	fn := func() { counter++ }
	
	// First call should execute
	throttle.Call(fn)
	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
	
	// Immediate second call should be throttled
	throttle.Call(fn)
	if counter != 1 {
		t.Errorf("Expected counter to still be 1, got %d", counter)
	}
	
	// After interval, should execute again
	time.Sleep(60 * time.Millisecond)
	throttle.Call(fn)
	if counter != 2 {
		t.Errorf("Expected counter to be 2, got %d", counter)
	}
}

// TestDebounceDelay tests debounce delay
func TestDebounceDelay(t *testing.T) {
	debounce := NewDebounce(50 * time.Millisecond)
	counter := 0
	fn := func() { counter++ }
	
	// Multiple rapid calls
	for i := 0; i < 5; i++ {
		debounce.Call(fn)
		time.Sleep(10 * time.Millisecond)
	}
	
	// Should not have executed yet
	if counter != 0 {
		t.Errorf("Expected counter to be 0, got %d", counter)
	}
	
	// Wait for debounce
	time.Sleep(60 * time.Millisecond)
	
	// Should have executed once
	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
}

// TestMicroInteractionProgress tests micro-interaction progress
func TestMicroInteractionProgress(t *testing.T) {
	mi := NewMicroInteraction("test", 100*time.Millisecond)
	
	if mi.IsComplete() {
		t.Error("Expected interaction to not be complete initially")
	}
	
	time.Sleep(50 * time.Millisecond)
	mi.Update()
	
	if mi.Progress < 0.4 || mi.Progress > 0.6 {
		t.Errorf("Expected progress around 0.5, got %.2f", mi.Progress)
	}
	
	time.Sleep(60 * time.Millisecond)
	mi.Update()
	
	if !mi.IsComplete() {
		t.Error("Expected interaction to be complete")
	}
	
	if mi.Progress != 1.0 {
		t.Errorf("Expected progress to be 1.0, got %.2f", mi.Progress)
	}
}

// BenchmarkConcurrentCacheAccess benchmarks concurrent cache access
func BenchmarkConcurrentCacheAccess(b *testing.B) {
	cache := NewRenderCache(1000)
	content := "test content"
	
	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), content, time.Minute)
	}
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%100)
			if i%2 == 0 {
				cache.Set(key, content, time.Minute)
			} else {
				_, _ = cache.Get(key)
			}
			i++
		}
	})
}

// BenchmarkConcurrentDirtyTracking benchmarks concurrent dirty tracking
func BenchmarkConcurrentDirtyTracking(b *testing.B) {
	tracker := NewDirtyTracker()
	panes := []string{"agents", "tasks", "logs", "footer", "header"}
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pane := panes[i%len(panes)]
			if i%3 == 0 {
				tracker.MarkDirty(pane)
			} else if i%3 == 1 {
				_ = tracker.IsDirty(pane)
			} else {
				tracker.ClearDirty(pane)
			}
			i++
		}
	})
}
