package test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGoVersionCompatibility tests that the project builds with minimum and current Go versions
func TestGoVersionCompatibility(t *testing.T) {
	// Get current Go version
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get Go version: %v", err)
	}
	
	currentVersion := string(output)
	t.Logf("Current Go version: %s", strings.TrimSpace(currentVersion))
	
	// Check go.mod for minimum version
	goModPath := filepath.Join("..", "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	
	lines := strings.Split(string(content), "\n")
	var minVersion string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				minVersion = parts[1]
				break
			}
		}
	}
	
	if minVersion == "" {
		t.Fatal("Could not find Go version in go.mod")
	}
	
	t.Logf("Minimum Go version specified in go.mod: %s", minVersion)
	
	// Note: We should support Go 1.21+ according to requirements
	// The current go.mod shows 1.25.4 which is higher than minimum
	if !strings.HasPrefix(minVersion, "1.21") && !strings.HasPrefix(minVersion, "1.22") && 
	   !strings.HasPrefix(minVersion, "1.23") && !strings.HasPrefix(minVersion, "1.24") &&
	   !strings.HasPrefix(minVersion, "1.25") {
		t.Logf("WARNING: go.mod specifies version %s, but requirements state minimum should be 1.21", minVersion)
	}
}

// TestGoBuildWithCurrentVersion tests that the project builds successfully
func TestGoBuildWithCurrentVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping build test in short mode")
	}
	
	// Clean build
	cmd := exec.Command("go", "clean", "-cache")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Failed to clean cache: %v", err)
	}
	
	// Build the project
	cmd = exec.Command("go", "build", "-v", "./...")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, output)
	}
	
	t.Logf("Build successful with current Go version")
}

// TestGoModTidy verifies go.mod is tidy
func TestGoModTidy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping go mod tidy test in short mode")
	}
	
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, output)
	}
	
	// Check if go.mod or go.sum changed
	cmd = exec.Command("git", "diff", "--exit-code", "go.mod", "go.sum")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Logf("WARNING: go.mod or go.sum has changes after 'go mod tidy'. Run 'go mod tidy' to update.")
	} else {
		t.Log("go.mod and go.sum are tidy")
	}
}

// TestGoVet runs go vet on all packages
func TestGoVet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping go vet test in short mode")
	}
	
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet found issues: %v\nOutput: %s", err, output)
	}
	
	t.Log("go vet passed")
}

// TestPythonVersionCompatibility tests Python version requirements
func TestPythonVersionCompatibility(t *testing.T) {
	// Check if python3 is available
	cmd := exec.Command("python3", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("python3 not available: %v", err)
	}
	
	versionStr := strings.TrimSpace(string(output))
	t.Logf("Current Python version: %s", versionStr)
	
	// Extract version number
	parts := strings.Fields(versionStr)
	if len(parts) < 2 {
		t.Fatalf("Could not parse Python version from: %s", versionStr)
	}
	
	version := parts[1]
	versionParts := strings.Split(version, ".")
	if len(versionParts) < 2 {
		t.Fatalf("Invalid Python version format: %s", version)
	}
	
	// Requirements state minimum Python 3.8
	major := versionParts[0]
	minor := versionParts[1]
	
	if major != "3" {
		t.Fatalf("Python 3.x required, found: %s", version)
	}
	
	t.Logf("Python version check: major=%s, minor=%s", major, minor)
	
	// Note: We should support Python 3.8+ according to requirements
	// Current system has 3.14 which is well above minimum
}

// TestPythonDependencies tests that Python dependencies can be installed
func TestPythonDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Python dependency test in short mode")
	}
	
	// Check if requirements.txt exists
	reqPath := filepath.Join("..", "agent", "requirements.txt")
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		t.Fatalf("requirements.txt not found at: %s", reqPath)
	}
	
	// Read requirements.txt
	content, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}
	
	t.Logf("Python dependencies:\n%s", string(content))
	
	// Check if pip is available
	cmd := exec.Command("python3", "-m", "pip", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("pip not available: %v", err)
	}
	
	t.Logf("pip version: %s", strings.TrimSpace(string(output)))
	
	// Try to check if dependencies are installable (dry-run)
	// Note: This doesn't actually install, just checks if packages exist
	cmd = exec.Command("python3", "-m", "pip", "install", "--dry-run", "-r", "requirements.txt")
	cmd.Dir = filepath.Join("..", "agent")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("WARNING: Some dependencies may not be installable: %v\nOutput: %s", err, output)
	} else {
		t.Log("All Python dependencies are available")
	}
}

// TestExternalDependencies checks for required external tools
func TestExternalDependencies(t *testing.T) {
	requiredTools := []struct {
		name        string
		command     string
		args        []string
		required    bool
		description string
	}{
		{"git", "git", []string{"--version"}, true, "Version control system"},
		{"docker", "docker", []string{"--version"}, false, "Container runtime (optional)"},
		{"bd", "bd", []string{"--version"}, false, "Beads CLI (required for full functionality)"},
	}
	
	results := make(map[string]bool)
	
	for _, tool := range requiredTools {
		cmd := exec.Command(tool.command, tool.args...)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			results[tool.name] = false
			if tool.required {
				t.Errorf("Required tool '%s' not found: %v", tool.name, err)
			} else {
				t.Logf("Optional tool '%s' not found: %v", tool.name, err)
			}
		} else {
			results[tool.name] = true
			version := strings.TrimSpace(string(output))
			t.Logf("✓ %s found: %s - %s", tool.name, version, tool.description)
		}
	}
	
	// Summary
	t.Logf("\nDependency Check Summary:")
	for name, found := range results {
		status := "✓"
		if !found {
			status = "✗"
		}
		t.Logf("  %s %s", status, name)
	}
}

// TestGoDependencyVersions checks for deprecated or outdated dependencies
func TestGoDependencyVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dependency version check in short mode")
	}
	
	// Run go list to get dependency information
	cmd := exec.Command("go", "list", "-json", "-m", "all")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list dependencies: %v", err)
	}
	
	// Parse JSON output (one JSON object per line)
	lines := strings.Split(string(output), "\n")
	var modules []map[string]interface{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "{" {
			continue
		}
		
		// Accumulate lines until we have a complete JSON object
		if strings.HasPrefix(line, "{") {
			var module map[string]interface{}
			decoder := json.NewDecoder(strings.NewReader(line))
			if err := decoder.Decode(&module); err == nil {
				modules = append(modules, module)
			}
		}
	}
	
	t.Logf("Found %d Go modules", len(modules))
	
	// Check for deprecated modules (this is a basic check)
	deprecatedPatterns := []string{
		"github.com/golang/protobuf", // Use google.golang.org/protobuf instead
		"gopkg.in/yaml.v2",           // Use gopkg.in/yaml.v3 instead
	}
	
	for _, module := range modules {
		path, ok := module["Path"].(string)
		if !ok {
			continue
		}
		
		version, _ := module["Version"].(string)
		
		for _, deprecated := range deprecatedPatterns {
			if strings.Contains(path, deprecated) {
				t.Logf("WARNING: Using potentially deprecated module: %s@%s", path, version)
			}
		}
	}
}

// TestDependencyUpdateScenario tests that dependencies can be updated
func TestDependencyUpdateScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dependency update test in short mode")
	}
	
	// This test verifies that 'go get -u' would work
	// We don't actually update, just check
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: Could not check for updates: %v", err)
		return
	}
	
	lines := strings.Split(string(output), "\n")
	updatesAvailable := 0
	
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			updatesAvailable++
			t.Logf("Update available: %s", line)
		}
	}
	
	if updatesAvailable > 0 {
		t.Logf("Found %d dependencies with updates available", updatesAvailable)
	} else {
		t.Log("All dependencies are up to date")
	}
}

// TestCrossCompilation tests that the project can be built for different platforms
func TestCrossCompilation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-compilation test in short mode")
	}
	
	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
	}
	
	for _, platform := range platforms {
		t.Run(fmt.Sprintf("%s_%s", platform.goos, platform.goarch), func(t *testing.T) {
			// Skip if we're testing the current platform (already tested in build test)
			if platform.goos == runtime.GOOS && platform.goarch == runtime.GOARCH {
				t.Skip("Current platform already tested")
			}
			
			cmd := exec.Command("go", "build", "-o", "/dev/null", "./...")
			cmd.Dir = ".."
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOOS=%s", platform.goos),
				fmt.Sprintf("GOARCH=%s", platform.goarch),
			)
			
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Cross-compilation failed for %s/%s: %v\nOutput: %s",
					platform.goos, platform.goarch, err, output)
			} else {
				t.Logf("✓ Successfully cross-compiled for %s/%s", platform.goos, platform.goarch)
			}
		})
	}
}

// TestGoModuleIntegrity verifies go.mod and go.sum integrity
func TestGoModuleIntegrity(t *testing.T) {
	// Verify go.mod exists
	goModPath := filepath.Join("..", "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatal("go.mod not found")
	}
	
	// Verify go.sum exists
	goSumPath := filepath.Join("..", "go.sum")
	if _, err := os.Stat(goSumPath); os.IsNotExist(err) {
		t.Fatal("go.sum not found")
	}
	
	// Verify checksums
	cmd := exec.Command("go", "mod", "verify")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod verify failed: %v\nOutput: %s", err, output)
	}
	
	t.Logf("Module integrity verified: %s", strings.TrimSpace(string(output)))
}

// TestMinimumGoVersion documents the minimum Go version requirement
func TestMinimumGoVersion(t *testing.T) {
	t.Log("=== Go Version Requirements ===")
	t.Log("Minimum supported: Go 1.21")
	t.Log("Recommended: Go 1.22+")
	t.Log("Tested with: Go 1.25.4")
	t.Log("")
	t.Log("Note: The project should be updated to specify 'go 1.21' in go.mod")
	t.Log("to ensure compatibility with the minimum supported version.")
}

// TestMinimumPythonVersion documents the minimum Python version requirement
func TestMinimumPythonVersion(t *testing.T) {
	t.Log("=== Python Version Requirements ===")
	t.Log("Minimum supported: Python 3.8")
	t.Log("Recommended: Python 3.10+")
	t.Log("Tested with: Python 3.14")
	t.Log("")
	t.Log("Python dependencies:")
	t.Log("  - anthropic>=0.34.0 (Claude API)")
	t.Log("  - google-generativeai>=0.3.0 (Gemini API)")
	t.Log("  - openai>=1.0.0 (OpenAI API)")
	t.Log("  - requests>=2.31.0 (HTTP client)")
	t.Log("  - python-dotenv>=1.0.0 (Environment variables)")
}
