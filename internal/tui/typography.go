package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TextStyle represents different text hierarchy levels
type TextStyle int

const (
	TextStyleH1 TextStyle = iota
	TextStyleH2
	TextStyleBody
	TextStyleCaption
	TextStyleCode
	TextStyleEmphasis
)

// CreateTextStyle creates a styled text based on hierarchy
func CreateTextStyle(style TextStyle, theme Theme) lipgloss.Style {
	switch style {
	case TextStyleH1:
		return lipgloss.NewStyle().
			Foreground(theme.Accent).
			Bold(true).
			MarginBottom(1)
	case TextStyleH2:
		return lipgloss.NewStyle().
			Foreground(theme.Purple).
			Bold(true)
	case TextStyleBody:
		return lipgloss.NewStyle().
			Foreground(theme.Foreground)
	case TextStyleCaption:
		return lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true)
	case TextStyleCode:
		return lipgloss.NewStyle().
			Foreground(theme.Cyan).
			Background(theme.DeepPurple).
			Padding(0, 1)
	case TextStyleEmphasis:
		return lipgloss.NewStyle().
			Foreground(theme.NeonPink).
			Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(theme.Foreground)
	}
}

// RenderH1 renders a large header with gradient
func RenderH1(text string, theme Theme) string {
	// Add letter spacing simulation
	spacedText := AddLetterSpacing(text, 1)
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(spacedText))
	
	return ApplyGradientToText(spacedText, gradient)
}

// RenderH2 renders a medium header with bold styling
func RenderH2(text string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.Purple).
		Bold(true)
	
	return style.Render(text)
}

// RenderBody renders body text
func RenderBody(text string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.Foreground)
	
	return style.Render(text)
}

// RenderCaption renders caption text (muted, smaller)
func RenderCaption(text string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.Muted)
	
	return style.Render(text)
}

// RenderCode renders code/monospace text with neon accent
func RenderCode(text string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.Cyan).
		Background(theme.DeepPurple).
		Padding(0, 1)
	
	return style.Render(text)
}

// RenderID renders an ID with neon styling
func RenderID(id string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.ElectricBlue).
		Bold(true)
	
	return style.Render("#" + id)
}

// AddLetterSpacing adds spacing between letters
func AddLetterSpacing(text string, spacing int) string {
	if spacing <= 0 {
		return text
	}
	
	runes := []rune(text)
	var result strings.Builder
	
	for i, r := range runes {
		result.WriteRune(r)
		if i < len(runes)-1 {
			result.WriteString(strings.Repeat(" ", spacing))
		}
	}
	
	return result.String()
}

// AddTextShadow simulates text shadow using Unicode characters
func AddTextShadow(text string, theme Theme) string {
	// Create shadow style
	_ = lipgloss.NewStyle().Foreground(theme.DeepPurple) // shadowStyle
	
	// Create main text style
	mainStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Bold(true)
	
	// Render main text (shadow simulation is limited in terminal)
	main := mainStyle.Render(text)
	
	return main
}

// AddTextOutline adds an outline effect to text
func AddTextOutline(text string, theme Theme) string {
	// Use bold and specific color for outline effect
	style := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Background(theme.DeepPurple)
	
	return style.Render(text)
}

// RenderWithIcon renders text with an icon/emoji
func RenderWithIcon(icon, text string, theme Theme) string {
	iconStyle := lipgloss.NewStyle().
		Foreground(theme.Accent)
	
	textStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)
	
	return iconStyle.Render(icon) + " " + textStyle.Render(text)
}

// ShimmerText creates a shimmering text effect (for animation)
func ShimmerText(text string, theme Theme, frame int) string {
	// Create a moving gradient effect
	runes := []rune(text)
	result := ""
	
	for i, r := range runes {
		// Calculate color based on position and frame
		offset := (i + frame) % 20
		brightness := float64(offset) / 20.0
		
		var color lipgloss.Color
		if brightness < 0.5 {
			color = ColorInterpolate(theme.Muted, theme.Accent, brightness*2)
		} else {
			color = ColorInterpolate(theme.Accent, theme.Muted, (brightness-0.5)*2)
		}
		
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(r))
	}
	
	return result
}

// WaveText creates a wave animation effect
func WaveText(text string, theme Theme, frame int) string {
	runes := []rune(text)
	result := ""
	
	for i, r := range runes {
		// Vary color intensity based on position
		_ = float64(i+frame) / 5.0 // phase
		
		// Vary color intensity
		color := theme.Accent
		if (i+frame)%3 == 0 {
			color = theme.GlowPink
		}
		
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(r))
	}
	
	return result
}

// FadeInText creates a fade-in effect
func FadeInText(text string, theme Theme, progress float64) string {
	// progress should be 0.0 to 1.0
	if progress >= 1.0 {
		return RenderBody(text, theme)
	}
	
	// Show partial text based on progress
	runes := []rune(text)
	visibleCount := int(float64(len(runes)) * progress)
	
	if visibleCount > len(runes) {
		visibleCount = len(runes)
	}
	
	visible := string(runes[:visibleCount])
	return RenderBody(visible, theme)
}

// RenderGradientText renders text with a gradient color
func RenderGradientText(text string, startColor, endColor lipgloss.Color) string {
	gradient := GenerateGradient(startColor, endColor, len(text))
	return ApplyGradientToText(text, gradient)
}

// RenderHolographicText creates a holographic rainbow shimmer effect
func RenderHolographicText(text string, theme Theme, frame int) string {
	// Create rainbow gradient
	colors := []lipgloss.Color{
		theme.NeonPink,
		theme.Purple,
		theme.ElectricBlue,
		theme.Cyan,
		theme.SunsetOrange,
		theme.NeonPink, // Loop back
	}
	
	runes := []rune(text)
	result := ""
	
	for i, r := range runes {
		// Calculate color index with animation
		colorIndex := (i + frame) % len(colors)
		nextColorIndex := (colorIndex + 1) % len(colors)
		
		// Interpolate between colors
		t := float64((i+frame)%10) / 10.0
		color := ColorInterpolate(colors[colorIndex], colors[nextColorIndex], t)
		
		style := lipgloss.NewStyle().Foreground(color).Bold(true)
		result += style.Render(string(r))
	}
	
	return result
}

// RenderNeonText renders text with neon glow effect
func RenderNeonText(text string, theme Theme) string {
	// Create glow by using bright color and bold
	style := lipgloss.NewStyle().
		Foreground(theme.GlowPink).
		Bold(true)
	
	return style.Render(text)
}

// RenderStatusText renders status text with appropriate styling
func RenderStatusText(status, text string, theme Theme) string {
	var color lipgloss.Color
	var icon string
	
	switch status {
	case "success":
		color = theme.Success
		icon = "✓"
	case "warning":
		color = theme.Warning
		icon = "⚠"
	case "error":
		color = theme.Error
		icon = "✗"
	case "info":
		color = theme.Info
		icon = "ℹ"
	default:
		color = theme.Foreground
		icon = "•"
	}
	
	iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(theme.Foreground)
	
	return iconStyle.Render(icon) + " " + textStyle.Render(text)
}

// TruncateText truncates text to fit width with ellipsis
func TruncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	
	return text[:maxWidth-3] + "..."
}

// PadText pads text to a specific width
func PadText(text string, width int, align string) string {
	textWidth := lipgloss.Width(text)
	
	if textWidth >= width {
		return text
	}
	
	padding := width - textWidth
	
	switch align {
	case "left":
		return text + strings.Repeat(" ", padding)
	case "right":
		return strings.Repeat(" ", padding) + text
	case "center":
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default:
		return text
	}
}

// RenderKeyBinding renders a keybinding hint
func RenderKeyBinding(key, description string, theme Theme) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)
	
	descStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)
	
	return keyStyle.Render("("+key+")") + " " + descStyle.Render(description)
}

// RenderMultilineText renders multiline text with proper styling
func RenderMultilineText(lines []string, style lipgloss.Style) string {
	var styledLines []string
	for _, line := range lines {
		styledLines = append(styledLines, style.Render(line))
	}
	return strings.Join(styledLines, "\n")
}
