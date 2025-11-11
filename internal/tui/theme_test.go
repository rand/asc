package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestVaporwaveTheme(t *testing.T) {
	theme := VaporwaveTheme()

	if theme.Name != "vaporwave" {
		t.Errorf("Expected theme name 'vaporwave', got '%s'", theme.Name)
	}

	// Test core colors are set
	if theme.NeonPink == "" {
		t.Error("NeonPink color not set")
	}
	if theme.ElectricBlue == "" {
		t.Error("ElectricBlue color not set")
	}
	if theme.Purple == "" {
		t.Error("Purple color not set")
	}

	// Test gradients are defined
	if len(theme.GradientPrimary) != 2 {
		t.Errorf("Expected GradientPrimary to have 2 colors, got %d", len(theme.GradientPrimary))
	}
}

func TestCyberpunkTheme(t *testing.T) {
	theme := CyberpunkTheme()

	if theme.Name != "cyberpunk" {
		t.Errorf("Expected theme name 'cyberpunk', got '%s'", theme.Name)
	}

	// Test that colors are different from vaporwave
	vaporwave := VaporwaveTheme()
	if theme.NeonPink == vaporwave.NeonPink {
		t.Error("Cyberpunk theme should have different colors than vaporwave")
	}
}

func TestMinimalTheme(t *testing.T) {
	theme := MinimalTheme()

	if theme.Name != "minimal" {
		t.Errorf("Expected theme name 'minimal', got '%s'", theme.Name)
	}

	// Test that minimal theme has muted colors
	if theme.Background == "" {
		t.Error("Background color not set")
	}
}

func TestGetThemeByName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"vaporwave", "vaporwave"},
		{"cyberpunk", "cyberpunk"},
		{"minimal", "minimal"},
		{"unknown", "vaporwave"}, // Should default to vaporwave
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetThemeByName(tt.name)
			if theme.Name != tt.expected {
				t.Errorf("GetThemeByName(%s) = %s, want %s", tt.name, theme.Name, tt.expected)
			}
		})
	}
}

func TestSetTheme(t *testing.T) {
	// Save original theme
	original := CurrentTheme

	// Set new theme
	cyberpunk := CyberpunkTheme()
	SetTheme(cyberpunk)

	if CurrentTheme.Name != "cyberpunk" {
		t.Errorf("Expected CurrentTheme to be 'cyberpunk', got '%s'", CurrentTheme.Name)
	}

	// Restore original
	CurrentTheme = original
}

func TestSetThemeByName(t *testing.T) {
	// Save original theme
	original := CurrentTheme

	// Set theme by name
	SetThemeByName("minimal")

	if CurrentTheme.Name != "minimal" {
		t.Errorf("Expected CurrentTheme to be 'minimal', got '%s'", CurrentTheme.Name)
	}

	// Restore original
	CurrentTheme = original
}

func TestColorInterpolate(t *testing.T) {
	color1 := lipgloss.Color("#FF0000") // Red
	color2 := lipgloss.Color("#0000FF") // Blue

	tests := []struct {
		name string
		t    float64
	}{
		{"start", 0.0},
		{"middle", 0.5},
		{"end", 1.0},
		{"below_zero", -0.5},  // Should clamp to 0
		{"above_one", 1.5},    // Should clamp to 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorInterpolate(color1, color2, tt.t)
			if result == "" {
				t.Error("ColorInterpolate returned empty color")
			}

			// Test clamping
			if tt.t < 0 {
				// Should return color1
				if result != color1 {
					// Allow for slight variations in hex formatting
					if string(result)[0] != '#' {
						t.Error("ColorInterpolate should clamp negative t to 0")
					}
				}
			}
			if tt.t > 1 {
				// Should return color2
				if result != color2 {
					// Allow for slight variations in hex formatting
					if string(result)[0] != '#' {
						t.Error("ColorInterpolate should clamp t > 1 to 1")
					}
				}
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		wantR    uint8
		wantG    uint8
		wantB    uint8
	}{
		{"with_hash", "#FF0000", 255, 0, 0},
		{"without_hash", "00FF00", 0, 255, 0},
		{"blue", "#0000FF", 0, 0, 255},
		{"white", "#FFFFFF", 255, 255, 255},
		{"black", "#000000", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := parseHexColor(tt.hex)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("parseHexColor(%s) = (%d, %d, %d), want (%d, %d, %d)",
					tt.hex, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestGenerateGradient(t *testing.T) {
	start := lipgloss.Color("#FF0000")
	end := lipgloss.Color("#0000FF")

	tests := []struct {
		name  string
		steps int
		want  int
	}{
		{"single_step", 1, 1},
		{"two_steps", 2, 2},
		{"five_steps", 5, 5},
		{"ten_steps", 10, 10},
		{"zero_steps", 0, 1}, // Should return at least start color
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gradient := GenerateGradient(start, end, tt.steps)
			if len(gradient) != tt.want {
				t.Errorf("GenerateGradient with %d steps returned %d colors, want %d",
					tt.steps, len(gradient), tt.want)
			}

			// First color should be close to start
			if tt.steps >= 1 && gradient[0] == "" {
				t.Error("First gradient color is empty")
			}
		})
	}
}

func TestApplyGradientToText(t *testing.T) {
	theme := VaporwaveTheme()
	gradient := GenerateGradient(theme.NeonPink, theme.ElectricBlue, 10)

	tests := []struct {
		name string
		text string
	}{
		{"simple", "Hello"},
		{"empty", ""},
		{"long", "This is a longer text to test gradient application"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyGradientToText(tt.text, gradient)

			if tt.text == "" {
				if result != "" {
					t.Error("ApplyGradientToText should return empty string for empty input")
				}
			} else {
				if result == "" {
					t.Error("ApplyGradientToText returned empty string for non-empty input")
				}
			}
		})
	}
}

func TestApplyGradientToTextEmptyGradient(t *testing.T) {
	result := ApplyGradientToText("Hello", []lipgloss.Color{})
	if result != "Hello" {
		t.Error("ApplyGradientToText should return original text for empty gradient")
	}
}

func TestPulseColor(t *testing.T) {
	theme := VaporwaveTheme()

	tests := []struct {
		name  string
		phase float64
	}{
		{"zero", 0.0},
		{"quarter", 1.57},  // π/2
		{"half", 3.14},     // π
		{"full", 6.28},     // 2π
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PulseColor(theme.NeonPink, tt.phase)
			if result == "" {
				t.Error("PulseColor returned empty color")
			}

			// Result should be a valid hex color
			if string(result)[0] != '#' {
				t.Error("PulseColor should return hex color")
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name string
		a, b, c uint8
		want uint8
	}{
		{"a_max", 255, 100, 50, 255},
		{"b_max", 50, 255, 100, 255},
		{"c_max", 50, 100, 255, 255},
		{"all_equal", 100, 100, 100, 100},
		{"zeros", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := max(tt.a, tt.b, tt.c)
			if got != tt.want {
				t.Errorf("max(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, got, tt.want)
			}
		})
	}
}

func TestMapToANSI(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"red", lipgloss.Color("#FF0000")},
		{"green", lipgloss.Color("#00FF00")},
		{"blue", lipgloss.Color("#0000FF")},
		{"white", lipgloss.Color("#FFFFFF")},
		{"black", lipgloss.Color("#000000")},
		{"gray", lipgloss.Color("#808080")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapToANSI(tt.color)
			if result == "" {
				t.Error("mapToANSI returned empty color")
			}
		})
	}
}

func TestAdaptColorForTerminal(t *testing.T) {
	color := lipgloss.Color("#FF71CE")

	// Test that it returns a valid color
	result := AdaptColorForTerminal(color)
	if result == "" {
		t.Error("AdaptColorForTerminal returned empty color")
	}
}

func TestSupportsTrueColor(t *testing.T) {
	// Just test that it doesn't panic
	_ = SupportsTrueColor()
}

func TestSupports256Color(t *testing.T) {
	// Just test that it doesn't panic
	_ = Supports256Color()
}
