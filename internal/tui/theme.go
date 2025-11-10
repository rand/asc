package tui

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete color theme for the TUI
type Theme struct {
	Name string
	
	// Core vaporwave colors
	NeonPink      lipgloss.Color
	ElectricBlue  lipgloss.Color
	Purple        lipgloss.Color
	Cyan          lipgloss.Color
	SunsetOrange  lipgloss.Color
	
	// Dark base colors
	DeepPurple    lipgloss.Color
	MidnightBlue  lipgloss.Color
	DarkTeal      lipgloss.Color
	
	// Glow/luminous effect colors (with alpha simulation)
	GlowPink      lipgloss.Color
	GlowBlue      lipgloss.Color
	GlowPurple    lipgloss.Color
	GlowCyan      lipgloss.Color
	
	// UI element colors
	Background    lipgloss.Color
	Foreground    lipgloss.Color
	Border        lipgloss.Color
	BorderGlow    lipgloss.Color
	Accent        lipgloss.Color
	Muted         lipgloss.Color
	
	// Status colors
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	Info          lipgloss.Color
	
	// Gradient definitions (start, end colors)
	GradientPrimary   []lipgloss.Color
	GradientSecondary []lipgloss.Color
	GradientAccent    []lipgloss.Color
}

// VaporwaveTheme returns the default vaporwave theme
func VaporwaveTheme() Theme {
	return Theme{
		Name: "vaporwave",
		
		// Core vaporwave colors
		NeonPink:     lipgloss.Color("#FF71CE"),
		ElectricBlue: lipgloss.Color("#01CDFE"),
		Purple:       lipgloss.Color("#B967FF"),
		Cyan:         lipgloss.Color("#05FFA1"),
		SunsetOrange: lipgloss.Color("#FFFB96"),
		
		// Dark base colors
		DeepPurple:   lipgloss.Color("#1A0933"),
		MidnightBlue: lipgloss.Color("#0D0221"),
		DarkTeal:     lipgloss.Color("#0F0E17"),
		
		// Glow colors (brighter versions for luminous effects)
		GlowPink:     lipgloss.Color("#FF9EE5"),
		GlowBlue:     lipgloss.Color("#4DE4FF"),
		GlowPurple:   lipgloss.Color("#D49FFF"),
		GlowCyan:     lipgloss.Color("#5FFFC4"),
		
		// UI element colors
		Background:   lipgloss.Color("#0D0221"),
		Foreground:   lipgloss.Color("#FFFFFF"),
		Border:       lipgloss.Color("#B967FF"),
		BorderGlow:   lipgloss.Color("#D49FFF"),
		Accent:       lipgloss.Color("#FF71CE"),
		Muted:        lipgloss.Color("#6B5B95"),
		
		// Status colors
		Success:      lipgloss.Color("#05FFA1"),
		Warning:      lipgloss.Color("#FFFB96"),
		Error:        lipgloss.Color("#FF71CE"),
		Info:         lipgloss.Color("#01CDFE"),
		
		// Gradient definitions
		GradientPrimary:   []lipgloss.Color{lipgloss.Color("#FF71CE"), lipgloss.Color("#B967FF")},
		GradientSecondary: []lipgloss.Color{lipgloss.Color("#01CDFE"), lipgloss.Color("#05FFA1")},
		GradientAccent:    []lipgloss.Color{lipgloss.Color("#B967FF"), lipgloss.Color("#01CDFE")},
	}
}

// CyberpunkTheme returns a cyberpunk-inspired theme
func CyberpunkTheme() Theme {
	return Theme{
		Name: "cyberpunk",
		
		// Core colors
		NeonPink:     lipgloss.Color("#FF0080"),
		ElectricBlue: lipgloss.Color("#00FFFF"),
		Purple:       lipgloss.Color("#8B00FF"),
		Cyan:         lipgloss.Color("#00FF00"),
		SunsetOrange: lipgloss.Color("#FFFF00"),
		
		// Dark base colors
		DeepPurple:   lipgloss.Color("#0A0A0A"),
		MidnightBlue: lipgloss.Color("#000000"),
		DarkTeal:     lipgloss.Color("#0D0D0D"),
		
		// Glow colors
		GlowPink:     lipgloss.Color("#FF33AA"),
		GlowBlue:     lipgloss.Color("#33FFFF"),
		GlowPurple:   lipgloss.Color("#AA33FF"),
		GlowCyan:     lipgloss.Color("#33FF33"),
		
		// UI element colors
		Background:   lipgloss.Color("#000000"),
		Foreground:   lipgloss.Color("#00FFFF"),
		Border:       lipgloss.Color("#FF0080"),
		BorderGlow:   lipgloss.Color("#FF33AA"),
		Accent:       lipgloss.Color("#00FFFF"),
		Muted:        lipgloss.Color("#404040"),
		
		// Status colors
		Success:      lipgloss.Color("#00FF00"),
		Warning:      lipgloss.Color("#FFFF00"),
		Error:        lipgloss.Color("#FF0000"),
		Info:         lipgloss.Color("#00FFFF"),
		
		// Gradient definitions
		GradientPrimary:   []lipgloss.Color{lipgloss.Color("#FF0080"), lipgloss.Color("#8B00FF")},
		GradientSecondary: []lipgloss.Color{lipgloss.Color("#00FFFF"), lipgloss.Color("#00FF00")},
		GradientAccent:    []lipgloss.Color{lipgloss.Color("#8B00FF"), lipgloss.Color("#00FFFF")},
	}
}

// MinimalTheme returns a minimal, clean theme
func MinimalTheme() Theme {
	return Theme{
		Name: "minimal",
		
		// Core colors (muted)
		NeonPink:     lipgloss.Color("#E0E0E0"),
		ElectricBlue: lipgloss.Color("#B0B0B0"),
		Purple:       lipgloss.Color("#C0C0C0"),
		Cyan:         lipgloss.Color("#D0D0D0"),
		SunsetOrange: lipgloss.Color("#F0F0F0"),
		
		// Dark base colors
		DeepPurple:   lipgloss.Color("#1A1A1A"),
		MidnightBlue: lipgloss.Color("#0A0A0A"),
		DarkTeal:     lipgloss.Color("#151515"),
		
		// Glow colors (subtle)
		GlowPink:     lipgloss.Color("#F0F0F0"),
		GlowBlue:     lipgloss.Color("#D0D0D0"),
		GlowPurple:   lipgloss.Color("#E0E0E0"),
		GlowCyan:     lipgloss.Color("#E5E5E5"),
		
		// UI element colors
		Background:   lipgloss.Color("#0A0A0A"),
		Foreground:   lipgloss.Color("#E0E0E0"),
		Border:       lipgloss.Color("#404040"),
		BorderGlow:   lipgloss.Color("#606060"),
		Accent:       lipgloss.Color("#FFFFFF"),
		Muted:        lipgloss.Color("#606060"),
		
		// Status colors
		Success:      lipgloss.Color("#90EE90"),
		Warning:      lipgloss.Color("#FFD700"),
		Error:        lipgloss.Color("#FF6B6B"),
		Info:         lipgloss.Color("#87CEEB"),
		
		// Gradient definitions
		GradientPrimary:   []lipgloss.Color{lipgloss.Color("#E0E0E0"), lipgloss.Color("#C0C0C0")},
		GradientSecondary: []lipgloss.Color{lipgloss.Color("#B0B0B0"), lipgloss.Color("#D0D0D0")},
		GradientAccent:    []lipgloss.Color{lipgloss.Color("#C0C0C0"), lipgloss.Color("#B0B0B0")},
	}
}

// ColorInterpolate interpolates between two hex colors
// t is a value between 0.0 and 1.0
func ColorInterpolate(color1, color2 lipgloss.Color, t float64) lipgloss.Color {
	// Clamp t to [0, 1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	
	// Parse hex colors
	r1, g1, b1 := parseHexColor(string(color1))
	r2, g2, b2 := parseHexColor(string(color2))
	
	// Interpolate
	r := uint8(float64(r1)*(1-t) + float64(r2)*t)
	g := uint8(float64(g1)*(1-t) + float64(g2)*t)
	b := uint8(float64(b1)*(1-t) + float64(b2)*t)
	
	// Return as hex color
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// parseHexColor parses a hex color string to RGB values
func parseHexColor(hex string) (uint8, uint8, uint8) {
	// Remove # if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	
	// Parse hex values
	var r, g, b uint8
	if len(hex) == 6 {
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	}
	
	return r, g, b
}

// GenerateGradient generates a gradient of colors between start and end
func GenerateGradient(start, end lipgloss.Color, steps int) []lipgloss.Color {
	if steps < 2 {
		return []lipgloss.Color{start}
	}
	
	gradient := make([]lipgloss.Color, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		gradient[i] = ColorInterpolate(start, end, t)
	}
	
	return gradient
}

// ApplyGradientToText applies a gradient to text (character by character)
func ApplyGradientToText(text string, gradient []lipgloss.Color) string {
	if len(gradient) == 0 {
		return text
	}
	
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	
	result := ""
	for i, r := range runes {
		// Map character position to gradient position
		gradientIndex := int(float64(i) / float64(len(runes)) * float64(len(gradient)))
		if gradientIndex >= len(gradient) {
			gradientIndex = len(gradient) - 1
		}
		
		color := gradient[gradientIndex]
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(r))
	}
	
	return result
}

// SupportsTrueColor checks if the terminal supports true color (24-bit)
func SupportsTrueColor() bool {
	// Check COLORTERM environment variable
	profile := lipgloss.ColorProfile()
	return profile == 3 // TrueColor profile value
}

// Supports256Color checks if the terminal supports 256 colors
func Supports256Color() bool {
	profile := lipgloss.ColorProfile()
	return profile == 3 || profile == 2 // TrueColor or ANSI256
}

// AdaptColorForTerminal adapts a color based on terminal capabilities
func AdaptColorForTerminal(color lipgloss.Color) lipgloss.Color {
	profile := lipgloss.ColorProfile()
	
	switch profile {
	case 3: // TrueColor
		// Terminal supports true color, return as-is
		return color
	case 2: // ANSI256
		// Terminal supports 256 colors, return as-is (lipgloss handles conversion)
		return color
	case 1: // ANSI
		// Terminal only supports 16 colors, map to closest ANSI color
		return mapToANSI(color)
	default:
		// Fallback to basic color
		return lipgloss.Color("15") // White
	}
}

// mapToANSI maps a hex color to the closest ANSI color
func mapToANSI(color lipgloss.Color) lipgloss.Color {
	r, g, b := parseHexColor(string(color))
	
	// Calculate luminance
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	
	// Map to ANSI colors based on RGB values and luminance
	if luminance < 64 {
		return lipgloss.Color("0") // Black
	} else if luminance > 192 {
		return lipgloss.Color("15") // Bright white
	}
	
	// Determine dominant color
	maxVal := max(r, g, b)
	
	if r == maxVal && r > 128 {
		if g > 100 && b < 100 {
			return lipgloss.Color("11") // Bright yellow
		}
		if b > 100 {
			return lipgloss.Color("13") // Bright magenta
		}
		return lipgloss.Color("9") // Bright red
	} else if g == maxVal && g > 128 {
		if b > 100 {
			return lipgloss.Color("14") // Bright cyan
		}
		return lipgloss.Color("10") // Bright green
	} else if b == maxVal && b > 128 {
		if r > 100 {
			return lipgloss.Color("13") // Bright magenta
		}
		return lipgloss.Color("12") // Bright blue
	}
	
	// Default to gray
	return lipgloss.Color("7")
}

// max returns the maximum of three uint8 values
func max(a, b, c uint8) uint8 {
	result := a
	if b > result {
		result = b
	}
	if c > result {
		result = c
	}
	return result
}

// PulseColor creates a pulsing effect by adjusting brightness
// phase should be a value that changes over time (e.g., frame counter)
func PulseColor(color lipgloss.Color, phase float64) lipgloss.Color {
	// Use sine wave for smooth pulsing (0.7 to 1.0 brightness)
	brightness := 0.85 + 0.15*math.Sin(phase)
	
	r, g, b := parseHexColor(string(color))
	
	// Apply brightness
	r = uint8(float64(r) * brightness)
	g = uint8(float64(g) * brightness)
	b = uint8(float64(b) * brightness)
	
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// GetThemeByName returns a theme by name
func GetThemeByName(name string) Theme {
	switch name {
	case "vaporwave":
		return VaporwaveTheme()
	case "cyberpunk":
		return CyberpunkTheme()
	case "minimal":
		return MinimalTheme()
	default:
		return VaporwaveTheme()
	}
}

// CurrentTheme holds the active theme (can be changed at runtime)
var CurrentTheme = VaporwaveTheme()

// SetTheme sets the active theme
func SetTheme(theme Theme) {
	CurrentTheme = theme
}

// SetThemeByName sets the active theme by name
func SetThemeByName(name string) {
	CurrentTheme = GetThemeByName(name)
}
