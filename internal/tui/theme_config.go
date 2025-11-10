package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

// ThemeConfig represents a theme configuration file
type ThemeConfig struct {
	Name   string            `json:"name"`
	Colors map[string]string `json:"colors"`
}

// ThemeManager manages theme loading, saving, and switching
type ThemeManager struct {
	currentTheme Theme
	themesDir    string
	configPath   string
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	homeDir, _ := os.UserHomeDir()
	themesDir := filepath.Join(homeDir, ".asc", "themes")
	configPath := filepath.Join(homeDir, ".asc", "theme.json")
	
	// Ensure themes directory exists
	os.MkdirAll(themesDir, 0755)
	
	return &ThemeManager{
		currentTheme: VaporwaveTheme(),
		themesDir:    themesDir,
		configPath:   configPath,
	}
}

// LoadTheme loads a theme by name
func (tm *ThemeManager) LoadTheme(name string) error {
	// Check built-in themes first
	switch name {
	case "vaporwave":
		tm.currentTheme = VaporwaveTheme()
		return nil
	case "cyberpunk":
		tm.currentTheme = CyberpunkTheme()
		return nil
	case "minimal":
		tm.currentTheme = MinimalTheme()
		return nil
	}
	
	// Try to load custom theme
	themePath := filepath.Join(tm.themesDir, name+".json")
	return tm.LoadThemeFromFile(themePath)
}

// LoadThemeFromFile loads a theme from a JSON file
func (tm *ThemeManager) LoadThemeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read theme file: %w", err)
	}
	
	var config ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse theme file: %w", err)
	}
	
	theme, err := tm.configToTheme(config)
	if err != nil {
		return fmt.Errorf("failed to convert theme config: %w", err)
	}
	
	tm.currentTheme = theme
	return nil
}

// SaveTheme saves the current theme to a file
func (tm *ThemeManager) SaveTheme(name string) error {
	config := tm.themeToConfig(tm.currentTheme, name)
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}
	
	themePath := filepath.Join(tm.themesDir, name+".json")
	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}
	
	return nil
}

// ExportTheme exports the current theme to a file
func (tm *ThemeManager) ExportTheme(path string) error {
	config := tm.themeToConfig(tm.currentTheme, tm.currentTheme.Name)
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}
	
	return nil
}

// ImportTheme imports a theme from a file
func (tm *ThemeManager) ImportTheme(path string) error {
	return tm.LoadThemeFromFile(path)
}

// ListThemes lists all available themes
func (tm *ThemeManager) ListThemes() ([]string, error) {
	themes := []string{"vaporwave", "cyberpunk", "minimal"}
	
	// Add custom themes
	entries, err := os.ReadDir(tm.themesDir)
	if err != nil {
		return themes, nil // Return built-in themes if directory doesn't exist
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			themes = append(themes, name)
		}
	}
	
	return themes, nil
}

// GetCurrentTheme returns the current theme
func (tm *ThemeManager) GetCurrentTheme() Theme {
	return tm.currentTheme
}

// SetCurrentTheme sets the current theme
func (tm *ThemeManager) SetCurrentTheme(theme Theme) {
	tm.currentTheme = theme
}

// SaveCurrentThemeConfig saves the current theme name to config
func (tm *ThemeManager) SaveCurrentThemeConfig() error {
	config := map[string]string{
		"theme": tm.currentTheme.Name,
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(tm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// LoadCurrentThemeConfig loads the saved theme name from config
func (tm *ThemeManager) LoadCurrentThemeConfig() error {
	data, err := os.ReadFile(tm.configPath)
	if err != nil {
		// Config doesn't exist, use default
		return nil
	}
	
	var config map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	if themeName, ok := config["theme"]; ok {
		return tm.LoadTheme(themeName)
	}
	
	return nil
}

// configToTheme converts a ThemeConfig to a Theme
func (tm *ThemeManager) configToTheme(config ThemeConfig) (Theme, error) {
	theme := Theme{
		Name: config.Name,
	}
	
	// Map colors from config
	if color, ok := config.Colors["neon_pink"]; ok {
		theme.NeonPink = lipgloss.Color(color)
	}
	if color, ok := config.Colors["electric_blue"]; ok {
		theme.ElectricBlue = lipgloss.Color(color)
	}
	if color, ok := config.Colors["purple"]; ok {
		theme.Purple = lipgloss.Color(color)
	}
	if color, ok := config.Colors["cyan"]; ok {
		theme.Cyan = lipgloss.Color(color)
	}
	if color, ok := config.Colors["sunset_orange"]; ok {
		theme.SunsetOrange = lipgloss.Color(color)
	}
	if color, ok := config.Colors["deep_purple"]; ok {
		theme.DeepPurple = lipgloss.Color(color)
	}
	if color, ok := config.Colors["midnight_blue"]; ok {
		theme.MidnightBlue = lipgloss.Color(color)
	}
	if color, ok := config.Colors["dark_teal"]; ok {
		theme.DarkTeal = lipgloss.Color(color)
	}
	if color, ok := config.Colors["glow_pink"]; ok {
		theme.GlowPink = lipgloss.Color(color)
	}
	if color, ok := config.Colors["glow_blue"]; ok {
		theme.GlowBlue = lipgloss.Color(color)
	}
	if color, ok := config.Colors["glow_purple"]; ok {
		theme.GlowPurple = lipgloss.Color(color)
	}
	if color, ok := config.Colors["glow_cyan"]; ok {
		theme.GlowCyan = lipgloss.Color(color)
	}
	if color, ok := config.Colors["background"]; ok {
		theme.Background = lipgloss.Color(color)
	}
	if color, ok := config.Colors["foreground"]; ok {
		theme.Foreground = lipgloss.Color(color)
	}
	if color, ok := config.Colors["border"]; ok {
		theme.Border = lipgloss.Color(color)
	}
	if color, ok := config.Colors["border_glow"]; ok {
		theme.BorderGlow = lipgloss.Color(color)
	}
	if color, ok := config.Colors["accent"]; ok {
		theme.Accent = lipgloss.Color(color)
	}
	if color, ok := config.Colors["muted"]; ok {
		theme.Muted = lipgloss.Color(color)
	}
	if color, ok := config.Colors["success"]; ok {
		theme.Success = lipgloss.Color(color)
	}
	if color, ok := config.Colors["warning"]; ok {
		theme.Warning = lipgloss.Color(color)
	}
	if color, ok := config.Colors["error"]; ok {
		theme.Error = lipgloss.Color(color)
	}
	if color, ok := config.Colors["info"]; ok {
		theme.Info = lipgloss.Color(color)
	}
	
	// Set gradients
	theme.GradientPrimary = []lipgloss.Color{theme.NeonPink, theme.Purple}
	theme.GradientSecondary = []lipgloss.Color{theme.ElectricBlue, theme.Cyan}
	theme.GradientAccent = []lipgloss.Color{theme.Purple, theme.ElectricBlue}
	
	return theme, nil
}

// themeToConfig converts a Theme to a ThemeConfig
func (tm *ThemeManager) themeToConfig(theme Theme, name string) ThemeConfig {
	return ThemeConfig{
		Name: name,
		Colors: map[string]string{
			"neon_pink":      string(theme.NeonPink),
			"electric_blue":  string(theme.ElectricBlue),
			"purple":         string(theme.Purple),
			"cyan":           string(theme.Cyan),
			"sunset_orange":  string(theme.SunsetOrange),
			"deep_purple":    string(theme.DeepPurple),
			"midnight_blue":  string(theme.MidnightBlue),
			"dark_teal":      string(theme.DarkTeal),
			"glow_pink":      string(theme.GlowPink),
			"glow_blue":      string(theme.GlowBlue),
			"glow_purple":    string(theme.GlowPurple),
			"glow_cyan":      string(theme.GlowCyan),
			"background":     string(theme.Background),
			"foreground":     string(theme.Foreground),
			"border":         string(theme.Border),
			"border_glow":    string(theme.BorderGlow),
			"accent":         string(theme.Accent),
			"muted":          string(theme.Muted),
			"success":        string(theme.Success),
			"warning":        string(theme.Warning),
			"error":          string(theme.Error),
			"info":           string(theme.Info),
		},
	}
}

// CreateAccessibilityTheme creates a high-contrast accessibility theme
func CreateAccessibilityTheme() Theme {
	return Theme{
		Name: "accessibility",
		
		// High contrast colors
		NeonPink:     lipgloss.Color("#FFFFFF"),
		ElectricBlue: lipgloss.Color("#FFFFFF"),
		Purple:       lipgloss.Color("#FFFFFF"),
		Cyan:         lipgloss.Color("#FFFFFF"),
		SunsetOrange: lipgloss.Color("#FFFFFF"),
		
		// Dark base colors
		DeepPurple:   lipgloss.Color("#000000"),
		MidnightBlue: lipgloss.Color("#000000"),
		DarkTeal:     lipgloss.Color("#000000"),
		
		// Glow colors (same as base for accessibility)
		GlowPink:     lipgloss.Color("#FFFFFF"),
		GlowBlue:     lipgloss.Color("#FFFFFF"),
		GlowPurple:   lipgloss.Color("#FFFFFF"),
		GlowCyan:     lipgloss.Color("#FFFFFF"),
		
		// UI element colors
		Background:   lipgloss.Color("#000000"),
		Foreground:   lipgloss.Color("#FFFFFF"),
		Border:       lipgloss.Color("#FFFFFF"),
		BorderGlow:   lipgloss.Color("#FFFFFF"),
		Accent:       lipgloss.Color("#FFFFFF"),
		Muted:        lipgloss.Color("#808080"),
		
		// Status colors (high contrast)
		Success:      lipgloss.Color("#00FF00"),
		Warning:      lipgloss.Color("#FFFF00"),
		Error:        lipgloss.Color("#FF0000"),
		Info:         lipgloss.Color("#00FFFF"),
		
		// Gradient definitions
		GradientPrimary:   []lipgloss.Color{lipgloss.Color("#FFFFFF"), lipgloss.Color("#FFFFFF")},
		GradientSecondary: []lipgloss.Color{lipgloss.Color("#FFFFFF"), lipgloss.Color("#FFFFFF")},
		GradientAccent:    []lipgloss.Color{lipgloss.Color("#FFFFFF"), lipgloss.Color("#FFFFFF")},
	}
}

// RenderThemePreview renders a preview of a theme
func RenderThemePreview(theme Theme, width int) string {
	// Create preview sections
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)
	
	title := titleStyle.Render(theme.Name)
	
	// Color swatches
	swatchStyle := lipgloss.NewStyle().
		Width(10).
		Height(2).
		Align(lipgloss.Center, lipgloss.Center)
	
	swatches := []string{
		swatchStyle.Background(theme.NeonPink).Render("Pink"),
		swatchStyle.Background(theme.ElectricBlue).Render("Blue"),
		swatchStyle.Background(theme.Purple).Render("Purple"),
		swatchStyle.Background(theme.Cyan).Render("Cyan"),
		swatchStyle.Background(theme.SunsetOrange).Render("Orange"),
	}
	
	swatchRow := lipgloss.JoinHorizontal(lipgloss.Top, swatches...)
	
	// Status indicators
	statusStyle := lipgloss.NewStyle().Padding(0, 1)
	
	statuses := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statusStyle.Foreground(theme.Success).Render("✓ Success"),
		statusStyle.Foreground(theme.Warning).Render("⚠ Warning"),
		statusStyle.Foreground(theme.Error).Render("✗ Error"),
		statusStyle.Foreground(theme.Info).Render("ℹ Info"),
	)
	
	// Border preview
	borderPreview := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.Border).
		Width(width - 4).
		Padding(1).
		Render("Border Preview")
	
	// Combine all elements
	preview := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		swatchRow,
		"",
		statuses,
		"",
		borderPreview,
	)
	
	return preview
}

// CreateThemeFromTemplate creates a new theme from a template
func CreateThemeFromTemplate(name string, baseTheme Theme, customColors map[string]lipgloss.Color) Theme {
	theme := baseTheme
	theme.Name = name
	
	// Apply custom colors
	for key, color := range customColors {
		switch key {
		case "neon_pink":
			theme.NeonPink = color
		case "electric_blue":
			theme.ElectricBlue = color
		case "purple":
			theme.Purple = color
		case "cyan":
			theme.Cyan = color
		case "sunset_orange":
			theme.SunsetOrange = color
		case "accent":
			theme.Accent = color
		case "background":
			theme.Background = color
		case "foreground":
			theme.Foreground = color
		}
	}
	
	return theme
}

// ValidateTheme validates a theme configuration
func ValidateTheme(theme Theme) error {
	// Check that all required colors are set
	if theme.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	
	// Validate color format (basic check)
	colors := []lipgloss.Color{
		theme.NeonPink,
		theme.ElectricBlue,
		theme.Purple,
		theme.Cyan,
		theme.Background,
		theme.Foreground,
	}
	
	for _, color := range colors {
		if string(color) == "" {
			return fmt.Errorf("invalid color in theme")
		}
	}
	
	return nil
}

// GetThemeDocumentation returns documentation for theme customization
func GetThemeDocumentation() string {
	doc := `
# Theme Customization Guide

## Theme Structure

A theme consists of the following color categories:

### Core Colors
- neon_pink: Primary accent color
- electric_blue: Secondary accent color
- purple: Tertiary accent color
- cyan: Quaternary accent color
- sunset_orange: Highlight color

### Base Colors
- deep_purple: Dark background shade
- midnight_blue: Darker background shade
- dark_teal: Alternative dark shade

### Glow Colors
- glow_pink, glow_blue, glow_purple, glow_cyan: Luminous effect colors

### UI Colors
- background: Main background color
- foreground: Main text color
- border: Border color
- border_glow: Border glow effect color
- accent: Primary accent for highlights
- muted: Muted/secondary text color

### Status Colors
- success: Success state color (green)
- warning: Warning state color (yellow)
- error: Error state color (red)
- info: Info state color (blue)

## Creating a Custom Theme

1. Create a JSON file in ~/.asc/themes/
2. Define your colors using hex format (#RRGGBB)
3. Load the theme using the theme manager

Example theme.json:
{
  "name": "my-theme",
  "colors": {
    "neon_pink": "#FF71CE",
    "electric_blue": "#01CDFE",
    ...
  }
}

## Built-in Themes

- vaporwave: Default vaporwave aesthetic
- cyberpunk: High-contrast cyberpunk style
- minimal: Clean, minimal design
- accessibility: High-contrast for accessibility
`
	
	return doc
}
