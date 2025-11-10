package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BorderStyle represents different border styles
type BorderStyle int

const (
	BorderStyleRounded BorderStyle = iota
	BorderStyleDouble
	BorderStyleThick
	BorderStyleGlow
	BorderStyleNeon
)

// CustomBorder defines a custom border with ornaments
type CustomBorder struct {
	lipgloss.Border
	TopLeftOrnament     string
	TopRightOrnament    string
	BottomLeftOrnament  string
	BottomRightOrnament string
}

// VaporwaveBorder returns a vaporwave-styled border with glow effect
func VaporwaveBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:          "─",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "╭",
		TopRight:     "╮",
		BottomLeft:   "╰",
		BottomRight:  "╯",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		Middle:       "┼",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
	}
}

// DoubleBorder returns a double-line border
func DoubleBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:          "═",
		Bottom:       "═",
		Left:         "║",
		Right:        "║",
		TopLeft:      "╔",
		TopRight:     "╗",
		BottomLeft:   "╚",
		BottomRight:  "╝",
		MiddleLeft:   "╠",
		MiddleRight:  "╣",
		Middle:       "╬",
		MiddleTop:    "╦",
		MiddleBottom: "╩",
	}
}

// ThickBorder returns a thick border
func ThickBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:          "━",
		Bottom:       "━",
		Left:         "┃",
		Right:        "┃",
		TopLeft:      "┏",
		TopRight:     "┓",
		BottomLeft:   "┗",
		BottomRight:  "┛",
		MiddleLeft:   "┣",
		MiddleRight:  "┫",
		Middle:       "╋",
		MiddleTop:    "┳",
		MiddleBottom: "┻",
	}
}

// NeonBorder returns a neon-styled border with decorative elements
func NeonBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:          "▔",
		Bottom:       "▁",
		Left:         "▏",
		Right:        "▕",
		TopLeft:      "▛",
		TopRight:     "▜",
		BottomLeft:   "▙",
		BottomRight:  "▟",
		MiddleLeft:   "▌",
		MiddleRight:  "▐",
		Middle:       "▞",
		MiddleTop:    "▀",
		MiddleBottom: "▄",
	}
}

// GetBorderByStyle returns a border based on the style
func GetBorderByStyle(style BorderStyle) lipgloss.Border {
	switch style {
	case BorderStyleRounded:
		return VaporwaveBorder()
	case BorderStyleDouble:
		return DoubleBorder()
	case BorderStyleThick:
		return ThickBorder()
	case BorderStyleNeon:
		return NeonBorder()
	case BorderStyleGlow:
		return VaporwaveBorder() // Use rounded as base for glow
	default:
		return VaporwaveBorder()
	}
}

// CreateGlowBorder creates a border with a glow effect using gradient colors
func CreateGlowBorder(width, height int, title string, theme Theme) lipgloss.Style {
	// Create gradient for border
	gradient := GenerateGradient(theme.Border, theme.BorderGlow, 10)
	
	// Use middle color from gradient for consistent glow
	glowColor := gradient[len(gradient)/2]
	
	return lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(glowColor).
		Width(width).
		Height(height).
		Padding(0, 1)
}

// CreateGradientBorder creates a border with gradient colors
// Note: lipgloss doesn't support per-character border colors, so we simulate with a single color
func CreateGradientBorder(width, height int, title string, colors []lipgloss.Color) lipgloss.Style {
	// Use the middle color from the gradient
	borderColor := colors[len(colors)/2]
	
	return lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height).
		Padding(0, 1)
}

// CreatePaneStyle creates a styled pane with border and title
func CreatePaneStyle(width, height int, title string, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.Border).
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1)
}

// CreateGlowPaneStyle creates a pane with glow effect
func CreateGlowPaneStyle(width, height int, title string, theme Theme) lipgloss.Style {
	return CreateGlowBorder(width-2, height-2, title, theme)
}

// RenderTitleBar renders a title bar with decorative elements
func RenderTitleBar(title string, width int, theme Theme) string {
	// Create gradient for title
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(title))
	gradientTitle := ApplyGradientToText(title, gradient)
	
	// Add decorative elements
	leftOrnament := "◢"
	rightOrnament := "◣"
	
	ornamentStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	
	// Calculate spacing
	titleWidth := lipgloss.Width(title) + 4 // +4 for ornaments and spaces
	padding := (width - titleWidth) / 2
	if padding < 0 {
		padding = 0
	}
	
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-titleWidth-padding)
	
	return leftPad +
		ornamentStyle.Render(leftOrnament) +
		" " +
		gradientTitle +
		" " +
		ornamentStyle.Render(rightOrnament) +
		rightPad
}

// RenderTitleBarSimple renders a simple title bar with bold text
func RenderTitleBarSimple(title string, theme Theme) string {
	style := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)
	
	return style.Render(title)
}

// AddCornerOrnaments adds decorative corner ornaments to content
func AddCornerOrnaments(content string, width, height int, theme Theme) string {
	lines := strings.Split(content, "\n")
	
	// Ensure we have enough lines
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	
	// Add ornaments to corners
	ornamentStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	
	// Top corners
	if len(lines) > 0 {
		lines[0] = ornamentStyle.Render("◢") + lines[0][1:len(lines[0])-1] + ornamentStyle.Render("◣")
	}
	
	// Bottom corners
	if len(lines) > 1 {
		lastIdx := len(lines) - 1
		lines[lastIdx] = ornamentStyle.Render("◥") + lines[lastIdx][1:len(lines[lastIdx])-1] + ornamentStyle.Render("◤")
	}
	
	return strings.Join(lines, "\n")
}

// CreateLayeredBorder creates a border with depth using layered borders
func CreateLayeredBorder(width, height int, title string, theme Theme) string {
	// Create outer border (glow)
	outerStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.BorderGlow).
		Width(width).
		Height(height).
		Padding(0)
	
	// Create inner border (main)
	innerStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.Border).
		Width(width - 4).
		Height(height - 4).
		Padding(0, 1)
	
	// Render title
	titleBar := RenderTitleBarSimple(title, theme)
	
	// Compose layers
	innerContent := innerStyle.Render(titleBar)
	return outerStyle.Render(innerContent)
}

// AnimatedBorderColor returns a color for animated borders based on frame
func AnimatedBorderColor(theme Theme, frame int) lipgloss.Color {
	// Cycle through gradient colors
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], 60)
	
	index := frame % len(gradient)
	return gradient[index]
}

// PulsingBorderStyle creates a border style with pulsing glow effect
func PulsingBorderStyle(width, height int, theme Theme, phase float64) lipgloss.Style {
	// Create pulsing color
	pulsingColor := PulseColor(theme.Border, phase)
	
	return lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(pulsingColor).
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1)
}

// CreateShadowEffect creates a shadow effect for borders
// Note: This is simulated using Unicode characters
func CreateShadowEffect(content string, theme Theme) string {
	lines := strings.Split(content, "\n")
	
	shadowStyle := lipgloss.NewStyle().Foreground(theme.DeepPurple)
	
	// Add shadow to right and bottom
	var result []string
	for _, line := range lines {
		result = append(result, line+shadowStyle.Render("▌"))
	}
	
	// Add bottom shadow
	if len(result) > 0 {
		shadowLine := shadowStyle.Render(strings.Repeat("▀", lipgloss.Width(result[0])))
		result = append(result, shadowLine)
	}
	
	return strings.Join(result, "\n")
}

// GeometricOrnament returns a geometric ornament pattern
func GeometricOrnament(pattern string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.Accent)
	
	switch pattern {
	case "triangle":
		return style.Render("◢◣")
	case "diamond":
		return style.Render("◆")
	case "hexagon":
		return style.Render("⬡")
	case "star":
		return style.Render("✦")
	case "circle":
		return style.Render("◉")
	default:
		return style.Render("◆")
	}
}

// CreateDecorativeBorder creates a border with decorative geometric patterns
func CreateDecorativeBorder(width, height int, title string, theme Theme) string {
	// Create base border
	baseStyle := lipgloss.NewStyle().
		Border(DoubleBorder()).
		BorderForeground(theme.Border).
		Width(width - 2).
		Height(height - 2).
		Padding(0, 1)
	
	// Create title with ornaments
	leftOrnament := GeometricOrnament("triangle", theme)
	rightOrnament := GeometricOrnament("triangle", theme)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)
	
	decoratedTitle := leftOrnament + " " + titleStyle.Render(title) + " " + rightOrnament
	
	return baseStyle.Render(decoratedTitle)
}
