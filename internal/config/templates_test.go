package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		wantErr      bool
		wantName     string
	}{
		{
			name:         "solo template",
			templateType: TemplateSolo,
			wantErr:      false,
			wantName:     "solo",
		},
		{
			name:         "team template",
			templateType: TemplateTeam,
			wantErr:      false,
			wantName:     "team",
		},
		{
			name:         "swarm template",
			templateType: TemplateSwarm,
			wantErr:      false,
			wantName:     "swarm",
		},
		{
			name:         "invalid template",
			templateType: "invalid",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tmpl.Name != tt.wantName {
					t.Errorf("GetTemplate() name = %v, want %v", tmpl.Name, tt.wantName)
				}
				if tmpl.Content == "" {
					t.Errorf("GetTemplate() content is empty")
				}
				if tmpl.Description == "" {
					t.Errorf("GetTemplate() description is empty")
				}
			}
		})
	}
}

func TestListTemplates(t *testing.T) {
	templates := ListTemplates()
	
	if len(templates) != 3 {
		t.Errorf("ListTemplates() returned %d templates, want 3", len(templates))
	}
	
	// Check that all expected templates are present
	expectedNames := map[string]bool{
		"solo":  false,
		"team":  false,
		"swarm": false,
	}
	
	for _, tmpl := range templates {
		if _, exists := expectedNames[tmpl.Name]; exists {
			expectedNames[tmpl.Name] = true
		}
	}
	
	for name, found := range expectedNames {
		if !found {
			t.Errorf("ListTemplates() missing template: %s", name)
		}
	}
}

func TestSaveAndLoadCustomTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a test template
	testTemplate := &Template{
		Name:        "test-template",
		Description: "Test template",
		Content: `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["testing"]
`,
	}
	
	// Save template
	templatePath := filepath.Join(tmpDir, "test-template.toml")
	err := SaveTemplate(testTemplate, templatePath)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("SaveTemplate() did not create file")
	}
	
	// Load template
	loadedTemplate, err := LoadCustomTemplate(templatePath)
	if err != nil {
		t.Fatalf("LoadCustomTemplate() error = %v", err)
	}
	
	// Verify content matches
	if loadedTemplate.Content != testTemplate.Content {
		t.Errorf("LoadCustomTemplate() content mismatch")
	}
}

func TestSaveCustomTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a temporary config file
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

[agent.test-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["testing"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Save custom template
	err = SaveCustomTemplate(configPath, "my-custom-template")
	if err != nil {
		t.Fatalf("SaveCustomTemplate() error = %v", err)
	}
	
	// Verify template was saved
	templatePath := filepath.Join(tmpDir, ".asc", "templates", "my-custom-template.toml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("SaveCustomTemplate() did not create template file")
	}
	
	// Verify content matches
	savedContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read saved template: %v", err)
	}
	
	if string(savedContent) != configContent {
		t.Errorf("SaveCustomTemplate() content mismatch")
	}
}

func TestListCustomTemplates(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Test with no templates directory
	templates, err := ListCustomTemplates()
	if err != nil {
		t.Fatalf("ListCustomTemplates() error = %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("ListCustomTemplates() returned %d templates, want 0", len(templates))
	}
	
	// Create templates directory and add some templates
	templatesDir := filepath.Join(tmpDir, ".asc", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Add test templates
	testTemplates := []string{"template1", "template2", "template3"}
	for _, name := range testTemplates {
		templatePath := filepath.Join(templatesDir, name+".toml")
		err := os.WriteFile(templatePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}
	}
	
	// List templates
	templates, err = ListCustomTemplates()
	if err != nil {
		t.Fatalf("ListCustomTemplates() error = %v", err)
	}
	
	if len(templates) != len(testTemplates) {
		t.Errorf("ListCustomTemplates() returned %d templates, want %d", len(templates), len(testTemplates))
	}
	
	// Verify all templates are present
	foundTemplates := make(map[string]bool)
	for _, tmpl := range templates {
		foundTemplates[tmpl.Name] = true
	}
	
	for _, name := range testTemplates {
		if !foundTemplates[name] {
			t.Errorf("ListCustomTemplates() missing template: %s", name)
		}
	}
}

func TestLoadCustomTemplateByName(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create templates directory
	templatesDir := filepath.Join(tmpDir, ".asc", "templates")
	err := os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Create a test template
	testContent := `[core]
beads_db_path = "./test-repo"
`
	templatePath := filepath.Join(templatesDir, "test-template.toml")
	err = os.WriteFile(templatePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	
	// Load template by name
	tmpl, err := LoadCustomTemplateByName("test-template")
	if err != nil {
		t.Fatalf("LoadCustomTemplateByName() error = %v", err)
	}
	
	if tmpl.Name != "test-template" {
		t.Errorf("LoadCustomTemplateByName() name = %v, want test-template", tmpl.Name)
	}
	
	if tmpl.Content != testContent {
		t.Errorf("LoadCustomTemplateByName() content mismatch")
	}
	
	// Test loading non-existent template
	_, err = LoadCustomTemplateByName("non-existent")
	if err == nil {
		t.Errorf("LoadCustomTemplateByName() should return error for non-existent template")
	}
}

func TestSaveTemplate_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a test template
	testTemplate := &Template{
		Name:        "test-template",
		Description: "Test template",
		Content:     "[core]\nbeads_db_path = \"./test-repo\"\n",
	}
	
	// Save to a nested directory that doesn't exist
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "template.toml")
	err := SaveTemplate(testTemplate, nestedPath)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Errorf("SaveTemplate() did not create file in nested directory")
	}
	
	// Verify content
	content, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to read saved template: %v", err)
	}
	
	if string(content) != testTemplate.Content {
		t.Errorf("SaveTemplate() content mismatch")
	}
}

func TestSaveTemplate_WriteError(t *testing.T) {
	// Try to save to a read-only directory (simulate write error)
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err := os.MkdirAll(readOnlyDir, 0444) // Read-only
	if err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	
	testTemplate := &Template{
		Name:        "test-template",
		Description: "Test template",
		Content:     "[core]\nbeads_db_path = \"./test-repo\"\n",
	}
	
	templatePath := filepath.Join(readOnlyDir, "template.toml")
	err = SaveTemplate(testTemplate, templatePath)
	if err == nil {
		t.Errorf("SaveTemplate() should return error for read-only directory")
	}
}

func TestSaveCustomTemplate_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Try to save from non-existent config
	err := SaveCustomTemplate("/nonexistent/config.toml", "test-template")
	if err == nil {
		t.Errorf("SaveCustomTemplate() should return error for non-existent config")
	}
}

func TestSaveCustomTemplate_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create a temporary config file
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// Save custom template (templates directory doesn't exist yet)
	err = SaveCustomTemplate(configPath, "my-template")
	if err != nil {
		t.Fatalf("SaveCustomTemplate() error = %v", err)
	}
	
	// Verify templates directory was created
	templatesDir := filepath.Join(tmpDir, ".asc", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Errorf("SaveCustomTemplate() did not create templates directory")
	}
	
	// Verify template file exists
	templatePath := filepath.Join(templatesDir, "my-template.toml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("SaveCustomTemplate() did not create template file")
	}
}

func TestLoadCustomTemplate_NonExistent(t *testing.T) {
	_, err := LoadCustomTemplate("/nonexistent/template.toml")
	if err == nil {
		t.Errorf("LoadCustomTemplate() should return error for non-existent file")
	}
}

func TestListCustomTemplates_WithNonTomlFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create templates directory
	templatesDir := filepath.Join(tmpDir, ".asc", "templates")
	err := os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Add TOML templates
	tomlFiles := []string{"template1.toml", "template2.toml"}
	for _, name := range tomlFiles {
		templatePath := filepath.Join(templatesDir, name)
		err := os.WriteFile(templatePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}
	}
	
	// Add non-TOML files (should be ignored)
	nonTomlFiles := []string{"readme.txt", "config.json", "script.sh"}
	for _, name := range nonTomlFiles {
		filePath := filepath.Join(templatesDir, name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	// Add a subdirectory (should be ignored)
	subDir := filepath.Join(templatesDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// List templates
	templates, err := ListCustomTemplates()
	if err != nil {
		t.Fatalf("ListCustomTemplates() error = %v", err)
	}
	
	// Should only return TOML files
	if len(templates) != len(tomlFiles) {
		t.Errorf("ListCustomTemplates() returned %d templates, want %d", len(templates), len(tomlFiles))
	}
	
	// Verify template names don't include .toml extension
	for _, tmpl := range templates {
		if contains(tmpl.Name, ".toml") {
			t.Errorf("Template name should not include .toml extension: %s", tmpl.Name)
		}
	}
}
