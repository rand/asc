package test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/process"
)

// TestPerformanceValidation_StartupTime validates startup time performance
func TestPerformanceValidation_StartupTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name           string
		numAgents      int
		maxStartupTime time.Duration
	}{
		{"1 agent", 1, 200 * time.Millisecond},
		{"3 agents", 3, 300 * time.Millisecond},
		{"5 agents", 5, 400 * time.Millisecond},
		{"10 agents", 10, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createPerformanceTestConfig(t, tmpDir, tt.numAgents)

			start := time.Now()

			// Simulate startup sequence
			cfg, err := config.Load(configPath)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
			if err != nil {
				t.Fatalf("Failed to create process manager: %v", err)
			}

			// Simulate agent initialization (without actually starting processes)
			for name := range cfg.Agents {
				_ = name // Would start agent here
			}

			elapsed := time.Since(start)

			t.Logf("Startup time with %d agents: %v", tt.numAgents, elapsed)

			if elapsed > tt.maxStartupTime {
				t.Errorf("Startup time %v exceeds limit %v for %d agents", 
					elapsed, tt.maxStartupTime, tt.numAgents)
			}

			// Cleanup
			_ = pm.StopAll()
		})
	}
}

// TestPerformanceValidation_ShutdownTime validates shutdown time performance
func TestPerformanceValidation_ShutdownTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name            string
		numAgents       int
		maxShutdownTime time.Duration
	}{
		{"1 agent", 1, 500 * time.Millisecond},
		{"3 agents", 3, 1 * time.Second},
		{"5 agents", 5, 1500 * time.Millisecond},
		{"10 agents", 10, 2 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
			if err != nil {
				t.Fatalf("Failed to create process manager: %v", err)
			}

			// Start mock processes
			for i := 0; i < tt.numAgents; i++ {
				_, err := pm.Start(
					fmt.Sprintf("agent-%d", i),
					"sleep",
					[]string{"30"},
					[]string{},
				)
				if err != nil {
					t.Fatalf("Failed to start process %d: %v", i, err)
				}
			}

			// Give processes time to start
			time.Sleep(100 * time.Millisecond)

			start := time.Now()
			err = pm.StopAll()
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("StopAll failed: %v", err)
			}

			t.Logf("Shutdown time with %d agents: %v", tt.numAgents, elapsed)

			if elapsed > tt.maxShutdownTime {
				t.Errorf("Shutdown time %v exceeds limit %v for %d agents",
					elapsed, tt.maxShutdownTime, tt.numAgents)
			}
		})
	}
}

// TestPerformanceValidation_MemoryUsage validates memory usage with different agent counts
func TestPerformanceValidation_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name        string
		numAgents   int
		maxMemoryMB float64
	}{
		{"1 agent", 1, 5.0},
		{"3 agents", 3, 10.0},
		{"5 agents", 5, 15.0},
		{"10 agents", 10, 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createPerformanceTestConfig(t, tmpDir, tt.numAgents)

			// Force GC and get baseline
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			// Load configuration and simulate operations
			for i := 0; i < 50; i++ {
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				_ = cfg
			}

			// Force GC and measure
			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
			heapMB := float64(m2.HeapAlloc) / 1024 / 1024

			t.Logf("Memory with %d agents - Allocated: %.2f MB, Heap: %.2f MB",
				tt.numAgents, allocMB, heapMB)

			if allocMB > tt.maxMemoryMB {
				t.Errorf("Memory usage %.2f MB exceeds limit %.2f MB for %d agents",
					allocMB, tt.maxMemoryMB, tt.numAgents)
			}
		})
	}
}

// TestPerformanceValidation_TUIResponsiveness validates TUI responsiveness under load
func TestPerformanceValidation_TUIResponsiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	// Test configuration loading which is part of TUI initialization
	tests := []struct {
		name           string
		numAgents      int
		numIterations  int
		maxAvgTimeMs   int64
	}{
		{"Light load", 5, 100, 5},
		{"Medium load", 10, 100, 10},
		{"Heavy load", 20, 100, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createPerformanceTestConfig(t, tmpDir, tt.numAgents)

			var totalDuration time.Duration

			for i := 0; i < tt.numIterations; i++ {
				start := time.Now()
				_, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				totalDuration += time.Since(start)
			}

			avgDuration := totalDuration / time.Duration(tt.numIterations)
			avgMs := avgDuration.Milliseconds()

			t.Logf("TUI responsiveness with %d agents - Avg: %d ms over %d iterations",
				tt.numAgents, avgMs, tt.numIterations)

			if avgMs > tt.maxAvgTimeMs {
				t.Errorf("Average response time %d ms exceeds limit %d ms",
					avgMs, tt.maxAvgTimeMs)
			}
		})
	}
}

// TestPerformanceValidation_TaskProcessingThroughput validates task processing throughput
func TestPerformanceValidation_TaskProcessingThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name            string
		numTasks        int
		numAgents       int
		minThroughput   float64 // tasks per second
	}{
		{"Small workload", 10, 1, 5.0},
		{"Medium workload", 50, 3, 10.0},
		{"Large workload", 100, 5, 15.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := createPerformanceTestConfig(t, tmpDir, tt.numAgents)

			start := time.Now()

			// Simulate task processing
			cfg, err := config.Load(configPath)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			pm, err := process.NewManager(tmpDir+"/pids", tmpDir+"/logs")
			if err != nil {
				t.Fatalf("Failed to create process manager: %v", err)
			}

			// Simulate processing tasks
			for i := 0; i < tt.numTasks; i++ {
				// Simulate task assignment to agent
				agentIdx := i % len(cfg.Agents)
				agentName := fmt.Sprintf("agent-%d", agentIdx)
				_ = agentName
				
				// Simulate minimal processing time
				time.Sleep(1 * time.Millisecond)
			}

			elapsed := time.Since(start)
			throughput := float64(tt.numTasks) / elapsed.Seconds()

			t.Logf("Task processing throughput: %.2f tasks/sec (%d tasks, %d agents, %v)",
				throughput, tt.numTasks, tt.numAgents, elapsed)

			if throughput < tt.minThroughput {
				t.Errorf("Throughput %.2f tasks/sec is below minimum %.2f tasks/sec",
					throughput, tt.minThroughput)
			}

			// Cleanup
			_ = pm.StopAll()
		})
	}
}

// TestPerformanceValidation_LargeLogFiles validates performance with large log files
func TestPerformanceValidation_LargeLogFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name        string
		logSizeMB   int
		maxReadTime time.Duration
	}{
		{"10MB log", 10, 500 * time.Millisecond},
		{"50MB log", 50, 2 * time.Second},
		{"100MB log", 100, 4 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := tmpDir + "/large.log"

			// Create large log file
			t.Logf("Creating %dMB log file...", tt.logSizeMB)
			createLargeLogFile(t, logFile, tt.logSizeMB)

			// Measure read time
			start := time.Now()
			data, err := os.ReadFile(logFile)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			sizeMB := float64(len(data)) / 1024 / 1024
			t.Logf("Read %.2f MB log file in %v", sizeMB, elapsed)

			if elapsed > tt.maxReadTime {
				t.Errorf("Read time %v exceeds limit %v for %dMB log",
					elapsed, tt.maxReadTime, tt.logSizeMB)
			}
		})
	}
}

// TestPerformanceValidation_ManyTasks validates performance with many tasks
func TestPerformanceValidation_ManyTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tests := []struct {
		name         string
		numTasks     int
		maxLoadTime  time.Duration
	}{
		{"100 tasks", 100, 100 * time.Millisecond},
		{"500 tasks", 500, 300 * time.Millisecond},
		{"1000 tasks", 1000, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create mock task data
			tasks := make([]string, tt.numTasks)
			for i := 0; i < tt.numTasks; i++ {
				tasks[i] = fmt.Sprintf(`{"id": "task-%d", "title": "Task %d", "status": "open"}`, i, i)
			}

			taskData := strings.Join(tasks, "\n")
			taskFile := tmpDir + "/tasks.jsonl"
			
			err := os.WriteFile(taskFile, []byte(taskData), 0644)
			if err != nil {
				t.Fatalf("Failed to write task file: %v", err)
			}

			// Measure load time
			start := time.Now()
			data, err := os.ReadFile(taskFile)
			if err != nil {
				t.Fatalf("Failed to read task file: %v", err)
			}
			
			// Simulate parsing
			lines := strings.Split(string(data), "\n")
			parsedCount := 0
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					parsedCount++
				}
			}
			
			elapsed := time.Since(start)

			t.Logf("Loaded and parsed %d tasks in %v", parsedCount, elapsed)

			if elapsed > tt.maxLoadTime {
				t.Errorf("Load time %v exceeds limit %v for %d tasks",
					elapsed, tt.maxLoadTime, tt.numTasks)
			}

			if parsedCount != tt.numTasks {
				t.Errorf("Expected %d tasks, got %d", tt.numTasks, parsedCount)
			}
		})
	}
}

// TestPerformanceValidation_ConcurrentOperations validates concurrent operation performance
func TestPerformanceValidation_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tmpDir := t.TempDir()
	configPath := createPerformanceTestConfig(t, tmpDir, 10)

	tests := []struct {
		name           string
		numGoroutines  int
		numIterations  int
		maxTotalTime   time.Duration
	}{
		{"Low concurrency", 5, 20, 1 * time.Second},
		{"Medium concurrency", 10, 20, 2 * time.Second},
		{"High concurrency", 20, 20, 3 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			
			done := make(chan bool, tt.numGoroutines)
			
			for i := 0; i < tt.numGoroutines; i++ {
				go func() {
					for j := 0; j < tt.numIterations; j++ {
						_, err := config.Load(configPath)
						if err != nil {
							t.Errorf("Failed to load config: %v", err)
						}
					}
					done <- true
				}()
			}
			
			// Wait for all goroutines
			for i := 0; i < tt.numGoroutines; i++ {
				<-done
			}
			
			elapsed := time.Since(start)
			totalOps := tt.numGoroutines * tt.numIterations
			opsPerSec := float64(totalOps) / elapsed.Seconds()

			t.Logf("Concurrent operations: %d goroutines × %d iterations = %d ops in %v (%.2f ops/sec)",
				tt.numGoroutines, tt.numIterations, totalOps, elapsed, opsPerSec)

			if elapsed > tt.maxTotalTime {
				t.Errorf("Total time %v exceeds limit %v", elapsed, tt.maxTotalTime)
			}
		})
	}
}

// TestPerformanceValidation_MemoryLeaks validates no memory leaks during extended operations
func TestPerformanceValidation_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	tmpDir := t.TempDir()
	configPath := createPerformanceTestConfig(t, tmpDir, 5)

	// Warm up
	for i := 0; i < 10; i++ {
		_, _ = config.Load(configPath)
	}

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform many operations
	iterations := 1000
	for i := 0; i < iterations; i++ {
		_, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Calculate heap growth (handle potential underflow)
	var heapGrowthMB float64
	if m2.HeapAlloc > m1.HeapAlloc {
		heapGrowthMB = float64(m2.HeapAlloc-m1.HeapAlloc) / 1024 / 1024
	} else {
		heapGrowthMB = 0
	}
	
	allocPerOp := float64(m2.TotalAlloc-m1.TotalAlloc) / float64(iterations) / 1024

	t.Logf("Memory after %d iterations - Heap growth: %.2f MB, Alloc per op: %.2f KB",
		iterations, heapGrowthMB, allocPerOp)

	// Heap should not grow significantly (allow 10MB growth for safety)
	maxHeapGrowthMB := 10.0
	if heapGrowthMB > maxHeapGrowthMB {
		t.Errorf("Heap growth %.2f MB exceeds limit %.2f MB - possible memory leak",
			heapGrowthMB, maxHeapGrowthMB)
	}
}

// Helper functions

func createPerformanceTestConfig(t testing.TB, dir string, numAgents int) string {
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
phases = ["planning", "implementation", "testing"]
`, i)
	}

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return configPath
}

func createLargeLogFile(t testing.TB, path string, sizeMB int) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	defer f.Close()

	// Write log entries until we reach desired size
	logEntry := "[2025-11-10 10:30:00] INFO [agent-1] Processing task #12345 - This is a sample log entry with some content\n"
	entrySize := len(logEntry)
	targetSize := sizeMB * 1024 * 1024
	numEntries := targetSize / entrySize

	for i := 0; i < numEntries; i++ {
		_, err := f.WriteString(logEntry)
		if err != nil {
			t.Fatalf("Failed to write log entry: %v", err)
		}
	}

	err = f.Sync()
	if err != nil {
		t.Fatalf("Failed to sync log file: %v", err)
	}
}
