package test

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/process"
)

// BenchmarkConfigLoad benchmarks configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := createTestConfig(b, tmpDir, 10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.Load(configPath)
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
	}
}

// BenchmarkConfigLoadLarge benchmarks loading large configuration
func BenchmarkConfigLoadLarge(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := createTestConfig(b, tmpDir, 50)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.Load(configPath)
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
	}
}

// TestMemoryUsageUnderLoad tests memory usage with many agents
func TestMemoryUsageUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	tests := []struct {
		name        string
		agents      int
		maxMemoryMB float64
	}{
		{"Small", 5, 10},
		{"Medium", 20, 25},
		{"Large", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createTestConfig(t, tmpDir, tt.agents)
			
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			// Load configuration multiple times
			for i := 0; i < 100; i++ {
				_, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
			}

			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			// Use TotalAlloc to get cumulative allocations
			allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
			t.Logf("Memory allocated: %.2f MB", allocMB)

			if allocMB > tt.maxMemoryMB {
				t.Errorf("Memory usage %.2f MB exceeds limit %.2f MB", allocMB, tt.maxMemoryMB)
			}
		})
	}
}

// TestStartupTime tests application startup time
func TestStartupTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping startup time test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := createTestConfig(t, tmpDir, 10)

	start := time.Now()
	
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Simulate initialization
	_, err = process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
	if err != nil {
		t.Fatalf("Failed to create process manager: %v", err)
	}
	
	elapsed := time.Since(start)
	t.Logf("Startup time: %v", elapsed)

	maxStartupTime := 500 * time.Millisecond
	if elapsed > maxStartupTime {
		t.Errorf("Startup time %v exceeds limit %v", elapsed, maxStartupTime)
	}

	if len(cfg.Agents) != 10 {
		t.Errorf("Expected 10 agents, got %d", len(cfg.Agents))
	}
}

// TestShutdownTime tests graceful shutdown time
func TestShutdownTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping shutdown time test in short mode")
	}

	tmpDir := t.TempDir()
	pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
	if err != nil {
		t.Fatalf("Failed to create process manager: %v", err)
	}

	// Start mock processes
	for i := 0; i < 5; i++ {
		_, err := pm.Start(
			fmt.Sprintf("agent-%d", i),
			"sleep",
			[]string{"10"},
			[]string{},
		)
		if err != nil {
			t.Fatalf("Failed to start process: %v", err)
		}
	}

	start := time.Now()
	err = pm.StopAll()
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("StopAll failed: %v", err)
	}

	t.Logf("Shutdown time: %v", elapsed)

	maxShutdownTime := 2 * time.Second
	if elapsed > maxShutdownTime {
		t.Errorf("Shutdown time %v exceeds limit %v", elapsed, maxShutdownTime)
	}
}

// BenchmarkProcessOperations benchmarks process management operations
func BenchmarkProcessOperations(b *testing.B) {
	tmpDir := b.TempDir()
	pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
	if err != nil {
		b.Fatalf("Failed to create process manager: %v", err)
	}
	
	b.Run("Start", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			name := fmt.Sprintf("bench-agent-%d", i)
			b.StartTimer()
			
			pid, err := pm.Start(name, "echo", []string{"test"}, []string{})
			if err != nil {
				b.Fatalf("Start failed: %v", err)
			}
			
			b.StopTimer()
			_ = pm.Stop(pid)
			b.StartTimer()
		}
	})
	
	b.Run("IsRunning", func(b *testing.B) {
		name := "bench-agent"
		pid, err := pm.Start(name, "sleep", []string{"60"}, []string{})
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		defer pm.Stop(pid)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pm.IsRunning(pid)
		}
	})
}

// TestConfigLoadPerformance tests configuration loading performance
func TestConfigLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping config load performance test in short mode")
	}

	tests := []struct {
		name      string
		agents    int
		maxTimeMs int64
	}{
		{"10 agents", 10, 10},
		{"50 agents", 50, 50},
		{"100 agents", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createTestConfig(t, tmpDir, tt.agents)
			
			start := time.Now()
			_, err := config.Load(configPath)
			elapsed := time.Since(start)
			
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}
			
			t.Logf("Load time: %v", elapsed)
			
			if elapsed.Milliseconds() > tt.maxTimeMs {
				t.Errorf("Load time %v exceeds limit %d ms", elapsed, tt.maxTimeMs)
			}
		})
	}
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping regression test in short mode")
	}

	// Baseline performance metrics
	baselines := map[string]time.Duration{
		"config_load":    10 * time.Millisecond,
		"tui_render":     50 * time.Millisecond,
		"process_start":  100 * time.Millisecond,
		"process_stop":   500 * time.Millisecond,
	}

	t.Run("ConfigLoad", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := createTestConfig(t, tmpDir, 10)
		
		start := time.Now()
		_, err := config.Load(configPath)
		elapsed := time.Since(start)
		
		if err != nil {
			t.Fatalf("Config load failed: %v", err)
		}
		
		if elapsed > baselines["config_load"] {
			t.Errorf("Config load time %v exceeds baseline %v", elapsed, baselines["config_load"])
		}
	})

	// TUI render test skipped - TUI model not yet fully implemented
	t.Run("TUIRender", func(t *testing.T) {
		t.Skip("TUI model not yet fully implemented for performance testing")
	})

	t.Run("ProcessStart", func(t *testing.T) {
		tmpDir := t.TempDir()
		pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
		if err != nil {
			t.Fatalf("Failed to create process manager: %v", err)
		}
		
		start := time.Now()
		_, err = pm.Start("test-agent", "echo", []string{"test"}, []string{})
		elapsed := time.Since(start)
		
		if err != nil {
			t.Fatalf("Process start failed: %v", err)
		}
		
		if elapsed > baselines["process_start"] {
			t.Errorf("Process start time %v exceeds baseline %v", elapsed, baselines["process_start"])
		}
		
		_ = pm.StopAll()
	})
}

// Helper functions

func createTestConfig(t testing.TB, dir string, numAgents int) string {
	t.Helper()
	
	configPath := dir + "/asc.toml"
	content := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

`
	
	for i := 0; i < numAgents; i++ {
		content += fmt.Sprintf(`
[agent.agent-%d]
command = "echo"
model = "claude"
phases = ["planning", "implementation"]
`, i)
	}
	
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	return configPath
}
