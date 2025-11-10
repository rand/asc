package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusIndicator represents different status indicator types
type StatusIndicator int

const (
	IndicatorIdle StatusIndicator = iota
	IndicatorWorking
	IndicatorError
	IndicatorOffline
	IndicatorSuccess
	IndicatorWarning
)

// RenderGlowingOrb renders a glowing orb for agent status
func RenderGlowingOrb(status StatusIndicator, theme Theme, phase float64) string {
	var baseColor lipgloss.Color
	var icon string
	
	switch status {
	case IndicatorIdle:
		baseColor = theme.Success
		_ = theme.GlowCyan // glowColor
		icon = "●"
	case IndicatorWorking:
		baseColor = theme.Info
		_ = theme.GlowBlue // glowColor
		icon = "◉"
	case IndicatorError:
		baseColor = theme.Error
		_ = theme.GlowPink // glowColor
		icon = "⬤"
	case IndicatorOffline:
		baseColor = theme.Muted
		_ = theme.Muted // glowColor
		icon = "○"
	case IndicatorSuccess:
		baseColor = theme.Success
		_ = theme.GlowCyan // glowColor
		icon = "✓"
	case IndicatorWarning:
		baseColor = theme.Warning
		_ = theme.SunsetOrange // glowColor
		icon = "⚠"
	}
	
	// Apply pulsing effect
	pulsingColor := PulseColor(baseColor, phase)
	
	style := lipgloss.NewStyle().
		Foreground(pulsingColor).
		Bold(true)
	
	return style.Render(icon)
}

// RenderProgressBar renders a progress bar with gradient fill
func RenderProgressBar(progress float64, width int, theme Theme) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	
	// Calculate filled width
	filledWidth := int(float64(width) * progress)
	emptyWidth := width - filledWidth
	
	// Create gradient for filled portion
	gradient := GenerateGradient(theme.GradientSecondary[0], theme.GradientSecondary[1], width)
	
	// Build progress bar
	var bar strings.Builder
	
	// Filled portion with gradient
	for i := 0; i < filledWidth; i++ {
		style := lipgloss.NewStyle().Foreground(gradient[i])
		bar.WriteString(style.Render("█"))
	}
	
	// Empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	bar.WriteString(emptyStyle.Render(strings.Repeat("░", emptyWidth)))
	
	return bar.String()
}

// RenderProgressBarWithShine renders a progress bar with shine effect
func RenderProgressBarWithShine(progress float64, width int, theme Theme, frame int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	
	filledWidth := int(float64(width) * progress)
	emptyWidth := width - filledWidth
	
	var bar strings.Builder
	
	// Filled portion with moving shine
	for i := 0; i < filledWidth; i++ {
		// Create shine effect that moves across
		shinePos := frame % width
		
		var color lipgloss.Color
		if i == shinePos || i == shinePos-1 {
			// Shine position - use bright color
			color = theme.GlowCyan
		} else {
			// Normal fill color
			color = theme.Cyan
		}
		
		style := lipgloss.NewStyle().Foreground(color)
		bar.WriteString(style.Render("█"))
	}
	
	// Empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	bar.WriteString(emptyStyle.Render(strings.Repeat("░", emptyWidth)))
	
	return bar.String()
}

// RenderTaskStatusBadge renders a task status badge with rounded corners
func RenderTaskStatusBadge(status string, theme Theme) string {
	var color lipgloss.Color
	var text string
	
	switch status {
	case "open":
		color = theme.Muted
		text = "OPEN"
	case "in_progress":
		color = theme.Warning
		text = "IN PROGRESS"
	case "completed":
		color = theme.Success
		text = "COMPLETED"
	case "blocked":
		color = theme.Error
		text = "BLOCKED"
	default:
		color = theme.Foreground
		text = strings.ToUpper(status)
	}
	
	style := lipgloss.NewStyle().
		Foreground(color).
		Background(theme.DeepPurple).
		Bold(true).
		Padding(0, 1)
	
	return style.Render(text)
}

// RenderGlowingBadge renders a badge with glow effect
func RenderGlowingBadge(text string, theme Theme, phase float64) string {
	pulsingColor := PulseColor(theme.Accent, phase)
	
	style := lipgloss.NewStyle().
		Foreground(pulsingColor).
		Background(theme.DeepPurple).
		Bold(true).
		Padding(0, 1)
	
	return style.Render(text)
}

// RenderConnectionStatus renders a connection status indicator with signal waves
func RenderConnectionStatus(connected bool, theme Theme, frame int) string {
	if !connected {
		style := lipgloss.NewStyle().Foreground(theme.Error)
		return style.Render("○ offline")
	}
	
	// Animate signal waves
	wavePhase := frame % 3
	var waves string
	
	switch wavePhase {
	case 0:
		waves = "◜◝"
	case 1:
		waves = "◞◟"
	case 2:
		waves = "◠◡"
	}
	
	style := lipgloss.NewStyle().Foreground(theme.Success)
	return style.Render("● " + waves)
}

// RenderHealthMeter renders a health meter with gradient fill
func RenderHealthMeter(health float64, width int, theme Theme) string {
	if health < 0 {
		health = 0
	}
	if health > 1 {
		health = 1
	}
	
	filledWidth := int(float64(width) * health)
	emptyWidth := width - filledWidth
	
	// Choose color based on health level
	var startColor, endColor lipgloss.Color
	if health > 0.7 {
		startColor = theme.Success
		endColor = theme.GlowCyan
	} else if health > 0.3 {
		startColor = theme.Warning
		endColor = theme.SunsetOrange
	} else {
		startColor = theme.Error
		endColor = theme.GlowPink
	}
	
	gradient := GenerateGradient(startColor, endColor, width)
	
	var meter strings.Builder
	
	// Filled portion
	for i := 0; i < filledWidth; i++ {
		style := lipgloss.NewStyle().Foreground(gradient[i])
		meter.WriteString(style.Render("█"))
	}
	
	// Empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(theme.DeepPurple)
	meter.WriteString(emptyStyle.Render(strings.Repeat("░", emptyWidth)))
	
	return meter.String()
}

// RenderSparkle renders a sparkle/particle effect
func RenderSparkle(theme Theme, frame int) string {
	sparkles := []string{"✦", "✧", "⋆", "·"}
	
	index := frame % len(sparkles)
	sparkle := sparkles[index]
	
	// Cycle through colors
	colors := []lipgloss.Color{
		theme.NeonPink,
		theme.Cyan,
		theme.Purple,
		theme.SunsetOrange,
	}
	
	colorIndex := (frame / 2) % len(colors)
	
	style := lipgloss.NewStyle().Foreground(colors[colorIndex])
	return style.Render(sparkle)
}

// RenderPulsingDot renders a pulsing dot indicator
func RenderPulsingDot(theme Theme, phase float64) string {
	pulsingColor := PulseColor(theme.Accent, phase)
	
	style := lipgloss.NewStyle().
		Foreground(pulsingColor).
		Bold(true)
	
	return style.Render("●")
}

// RenderLoadingSpinner renders a loading spinner with vaporwave styling
func RenderLoadingSpinner(theme Theme, frame int) string {
	spinners := []string{"◐", "◓", "◑", "◒"}
	
	index := frame % len(spinners)
	spinner := spinners[index]
	
	// Create gradient color cycling
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(spinners))
	color := gradient[index]
	
	style := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
	
	return style.Render(spinner)
}

// RenderAnimatedDots renders animated loading dots
func RenderAnimatedDots(theme Theme, frame int) string {
	numDots := (frame % 4)
	dots := strings.Repeat(".", numDots)
	
	style := lipgloss.NewStyle().Foreground(theme.Accent)
	return style.Render(dots)
}

// RenderStateTransition renders a smooth color transition between states
func RenderStateTransition(fromState, toState StatusIndicator, theme Theme, progress float64) string {
	// Get colors for each state
	fromColor := getStatusColor(fromState, theme)
	toColor := getStatusColor(toState, theme)
	
	// Interpolate color
	transitionColor := ColorInterpolate(fromColor, toColor, progress)
	
	style := lipgloss.NewStyle().
		Foreground(transitionColor).
		Bold(true)
	
	return style.Render("●")
}

// getStatusColor returns the color for a status indicator
func getStatusColor(status StatusIndicator, theme Theme) lipgloss.Color {
	switch status {
	case IndicatorIdle:
		return theme.Success
	case IndicatorWorking:
		return theme.Info
	case IndicatorError:
		return theme.Error
	case IndicatorOffline:
		return theme.Muted
	case IndicatorSuccess:
		return theme.Success
	case IndicatorWarning:
		return theme.Warning
	default:
		return theme.Foreground
	}
}

// RenderActivityIndicator renders an activity indicator with animation
func RenderActivityIndicator(active bool, theme Theme, frame int) string {
	if !active {
		style := lipgloss.NewStyle().Foreground(theme.Muted)
		return style.Render("○")
	}
	
	// Rotating animation
	rotations := []string{"◜", "◝", "◞", "◟"}
	index := frame % len(rotations)
	
	style := lipgloss.NewStyle().
		Foreground(theme.Info).
		Bold(true)
	
	return style.Render(rotations[index])
}

// RenderNotificationBadge renders a notification badge with count
func RenderNotificationBadge(count int, theme Theme) string {
	if count <= 0 {
		return ""
	}
	
	text := fmt.Sprintf("%d", count)
	if count > 99 {
		text = "99+"
	}
	
	style := lipgloss.NewStyle().
		Foreground(theme.Background).
		Background(theme.Error).
		Bold(true).
		Padding(0, 1)
	
	return style.Render(text)
}

// RenderSignalStrength renders a signal strength indicator
func RenderSignalStrength(strength int, theme Theme) string {
	// strength: 0-4 bars
	if strength < 0 {
		strength = 0
	}
	if strength > 4 {
		strength = 4
	}
	
	bars := []string{"▁", "▂", "▄", "▆", "█"}
	
	var result strings.Builder
	for i := 0; i < 4; i++ {
		var style lipgloss.Style
		if i < strength {
			// Active bar with gradient color
			colors := GenerateGradient(theme.Success, theme.GlowCyan, 4)
			style = lipgloss.NewStyle().Foreground(colors[i])
		} else {
			// Inactive bar
			style = lipgloss.NewStyle().Foreground(theme.Muted)
		}
		
		result.WriteString(style.Render(bars[i]))
	}
	
	return result.String()
}

// RenderBatteryIndicator renders a battery level indicator
func RenderBatteryIndicator(level float64, theme Theme) string {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	
	// Choose icon based on level
	var icon string
	if level > 0.75 {
		icon = "█████"
	} else if level > 0.5 {
		icon = "████░"
	} else if level > 0.25 {
		icon = "███░░"
	} else if level > 0.1 {
		icon = "██░░░"
	} else {
		icon = "█░░░░"
	}
	
	// Choose color based on level
	var color lipgloss.Color
	if level > 0.5 {
		color = theme.Success
	} else if level > 0.2 {
		color = theme.Warning
	} else {
		color = theme.Error
	}
	
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render("⚡" + icon)
}

// RenderWaveAnimation renders a wave animation effect
func RenderWaveAnimation(theme Theme, frame int, width int) string {
	var result strings.Builder
	
	for i := 0; i < width; i++ {
		// Calculate wave height using sine
		phase := float64(i+frame) / 5.0
		height := math.Sin(phase)
		
		// Map height to character
		var char string
		if height > 0.5 {
			char = "▀"
		} else if height > 0 {
			char = "▄"
		} else if height > -0.5 {
			char = "▁"
		} else {
			char = "▂"
		}
		
		// Color based on position
		gradient := GenerateGradient(theme.GradientSecondary[0], theme.GradientSecondary[1], width)
		style := lipgloss.NewStyle().Foreground(gradient[i])
		
		result.WriteString(style.Render(char))
	}
	
	return result.String()
}

// RenderPulseRing renders a pulsing ring effect
func RenderPulseRing(theme Theme, phase float64) string {
	// Create pulsing effect with different ring characters
	intensity := 0.5 + 0.5*math.Sin(phase)
	
	var ring string
	if intensity > 0.75 {
		ring = "◉"
	} else if intensity > 0.5 {
		ring = "◎"
	} else if intensity > 0.25 {
		ring = "○"
	} else {
		ring = "◌"
	}
	
	color := PulseColor(theme.Accent, phase)
	style := lipgloss.NewStyle().Foreground(color)
	
	return style.Render(ring)
}
