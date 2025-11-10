package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalStyle represents different modal styles
type ModalStyle int

const (
	ModalStyleGlass ModalStyle = iota
	ModalStyleNeon
	ModalStyleMinimal
)

// CreateGlassModal creates a modal with frosted glass effect
func CreateGlassModal(width, height int, title, content string, theme Theme) string {
	// Create glass effect with semi-transparent background simulation
	glassStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.BorderGlow).
		Background(theme.MidnightBlue).
		Foreground(theme.Foreground).
		Width(width - 4).
		Height(height - 4).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)
	
	// Create title bar with gradient
	titleGradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(title))
	gradientTitle := ApplyGradientToText(title, titleGradient)
	
	titleBar := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 8).
		MarginBottom(1).
		Render(gradientTitle)
	
	// Combine title and content
	modalContent := lipgloss.JoinVertical(
		lipgloss.Center,
		titleBar,
		content,
	)
	
	return glassStyle.Render(modalContent)
}

// CreateNeonModal creates a modal with neon styling
func CreateNeonModal(width, height int, title, content string, theme Theme) string {
	// Create neon border effect
	neonStyle := lipgloss.NewStyle().
		Border(DoubleBorder()).
		BorderForeground(theme.GlowPink).
		Background(theme.DeepPurple).
		Foreground(theme.Foreground).
		Width(width - 4).
		Height(height - 4).
		Padding(1, 2)
	
	// Create glowing title
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.NeonPink).
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 8).
		MarginBottom(1)
	
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		content,
	)
	
	return neonStyle.Render(modalContent)
}

// CreateMinimalModal creates a minimal modal
func CreateMinimalModal(width, height int, title, content string, theme Theme) string {
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Background(theme.Background).
		Foreground(theme.Foreground).
		Width(width - 4).
		Height(height - 4).
		Padding(1, 2)
	
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 8).
		MarginBottom(1)
	
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		content,
	)
	
	return modalStyle.Render(modalContent)
}

// CreateBackdropBlur simulates a backdrop blur effect
func CreateBackdropBlur(width, height int, theme Theme) string {
	// Simulate blur with semi-transparent overlay
	blurStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Foreground(theme.Muted)
	
	var lines []string
	for i := 0; i < height; i++ {
		lines = append(lines, blurStyle.Render(strings.Repeat("░", width)))
	}
	
	return strings.Join(lines, "\n")
}

// CreateModalShadow creates a shadow effect for modals
func CreateModalShadow(width, height int, theme Theme) string {
	shadowStyle := lipgloss.NewStyle().
		Foreground(theme.DeepPurple)
	
	var lines []string
	for i := 0; i < height; i++ {
		// Shadow gets lighter as it goes down
		opacity := float64(i) / float64(height)
		char := "▓"
		if opacity > 0.7 {
			char = "░"
		} else if opacity > 0.4 {
			char = "▒"
		}
		
		lines = append(lines, shadowStyle.Render(strings.Repeat(char, width)))
	}
	
	return strings.Join(lines, "\n")
}

// CreateCloseButton creates a styled close button
func CreateCloseButton(theme Theme, hover bool) string {
	var style lipgloss.Style
	
	if hover {
		style = lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Error).
			Bold(true).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Foreground(theme.Error).
			Background(theme.DeepPurple).
			Padding(0, 1)
	}
	
	return style.Render("✕")
}

// CreateModalButton creates a styled button for modals
func CreateModalButton(text string, theme Theme, primary bool, hover bool) string {
	var style lipgloss.Style
	
	if primary {
		if hover {
			style = lipgloss.NewStyle().
				Foreground(theme.Background).
				Background(theme.GlowPink).
				Bold(true).
				Padding(0, 2)
		} else {
			style = lipgloss.NewStyle().
				Foreground(theme.Background).
				Background(theme.NeonPink).
				Bold(true).
				Padding(0, 2)
		}
	} else {
		if hover {
			style = lipgloss.NewStyle().
				Foreground(theme.Foreground).
				Background(theme.Muted).
				Padding(0, 2)
		} else {
			style = lipgloss.NewStyle().
				Foreground(theme.Foreground).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Border).
				Padding(0, 2)
		}
	}
	
	return style.Render(text)
}

// CreateModalWithButtons creates a modal with action buttons
func CreateModalWithButtons(width, height int, title, content string, buttons []string, theme Theme) string {
	// Create modal content
	modal := CreateGlassModal(width, height-3, title, content, theme)
	
	// Create button row
	var buttonWidgets []string
	for i, btnText := range buttons {
		isPrimary := i == 0 // First button is primary
		button := CreateModalButton(btnText, theme, isPrimary, false)
		buttonWidgets = append(buttonWidgets, button)
	}
	
	buttonRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		buttonWidgets...,
	)
	
	buttonRowStyled := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width).
		MarginTop(1).
		Render(buttonRow)
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		modal,
		buttonRowStyled,
	)
}

// AnimateFadeIn creates a fade-in animation for modals
func AnimateFadeIn(content string, theme Theme, progress float64) string {
	if progress >= 1.0 {
		return content
	}
	
	// Simulate fade by adjusting color intensity
	// This is simplified - in reality we'd adjust all colors
	return content
}

// AnimateFadeOut creates a fade-out animation for modals
func AnimateFadeOut(content string, theme Theme, progress float64) string {
	if progress >= 1.0 {
		return ""
	}
	
	// Simulate fade by adjusting color intensity
	return content
}

// CreateFloatingModal creates a modal with floating animation
func CreateFloatingModal(width, height int, title, content string, theme Theme, phase float64) string {
	// Create base modal
	modal := CreateGlassModal(width, height, title, content, theme)
	
	// Add subtle floating effect (simulated with spacing)
	// In terminal, we can't actually move elements, but we can add visual cues
	return modal
}

// CreatePulsingModal creates a modal with pulsing glow
func CreatePulsingModal(width, height int, title, content string, theme Theme, phase float64) string {
	// Create pulsing border color
	pulsingColor := PulseColor(theme.BorderGlow, phase)
	
	modalStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(pulsingColor).
		Background(theme.MidnightBlue).
		Foreground(theme.Foreground).
		Width(width - 4).
		Height(height - 4).
		Padding(1, 2)
	
	// Create title
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 8).
		MarginBottom(1)
	
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		content,
	)
	
	return modalStyle.Render(modalContent)
}

// CreateShimmeringModal creates a modal with shimmering effect
func CreateShimmeringModal(width, height int, title, content string, theme Theme, frame int) string {
	// Create shimmering title
	shimmerTitle := ShimmerText(title, theme, frame)
	
	modalStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.BorderGlow).
		Background(theme.MidnightBlue).
		Foreground(theme.Foreground).
		Width(width - 4).
		Height(height - 4).
		Padding(1, 2)
	
	titleBar := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 8).
		MarginBottom(1).
		Render(shimmerTitle)
	
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		content,
	)
	
	return modalStyle.Render(modalContent)
}

// CreateConfirmDialog creates a confirmation dialog
func CreateConfirmDialog(width, height int, title, message string, theme Theme) string {
	// Create message content
	messageStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Align(lipgloss.Center).
		Width(width - 12)
	
	styledMessage := messageStyle.Render(message)
	
	// Create buttons
	confirmBtn := CreateModalButton("Confirm", theme, true, false)
	cancelBtn := CreateModalButton("Cancel", theme, false, false)
	
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		confirmBtn,
		"  ",
		cancelBtn,
	)
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styledMessage,
		"",
		buttons,
	)
	
	return CreateGlassModal(width, height, title, content, theme)
}

// CreateInputDialog creates an input dialog
func CreateInputDialog(width, height int, title, prompt, value string, theme Theme) string {
	// Create prompt
	promptStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		MarginBottom(1)
	
	// Create input field
	inputStyle := lipgloss.NewStyle().
		Foreground(theme.Cyan).
		Background(theme.DeepPurple).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width - 16).
		Padding(0, 1)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		promptStyle.Render(prompt),
		inputStyle.Render(value+"│"), // Cursor simulation
	)
	
	return CreateGlassModal(width, height, title, content, theme)
}

// CreateLoadingModal creates a loading modal with spinner
func CreateLoadingModal(width, height int, title, message string, theme Theme, frame int) string {
	// Create spinner
	spinner := RenderLoadingSpinner(theme, frame)
	
	// Create message
	messageStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Align(lipgloss.Center)
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		spinner,
		"",
		messageStyle.Render(message),
	)
	
	return CreateGlassModal(width, height, title, content, theme)
}

// CreateErrorModal creates an error modal
func CreateErrorModal(width, height int, title, errorMsg string, theme Theme) string {
	// Create error icon
	errorIcon := lipgloss.NewStyle().
		Foreground(theme.Error).
		Bold(true).
		Align(lipgloss.Center).
		Render("✗")
	
	// Create error message
	messageStyle := lipgloss.NewStyle().
		Foreground(theme.Error).
		Align(lipgloss.Center).
		Width(width - 12)
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		errorIcon,
		"",
		messageStyle.Render(errorMsg),
	)
	
	return CreateGlassModal(width, height, title, content, theme)
}

// CreateSuccessModal creates a success modal
func CreateSuccessModal(width, height int, title, message string, theme Theme) string {
	// Create success icon
	successIcon := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true).
		Align(lipgloss.Center).
		Render("✓")
	
	// Create message
	messageStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Align(lipgloss.Center).
		Width(width - 12)
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		successIcon,
		"",
		messageStyle.Render(message),
	)
	
	return CreateGlassModal(width, height, title, content, theme)
}

// CenterModal centers a modal in the terminal
func CenterModal(modal string, termWidth, termHeight int) string {
	modalLines := strings.Split(modal, "\n")
	modalHeight := len(modalLines)
	modalWidth := 0
	
	// Find max width
	for _, line := range modalLines {
		if lipgloss.Width(line) > modalWidth {
			modalWidth = lipgloss.Width(line)
		}
	}
	
	// Calculate padding
	topPadding := (termHeight - modalHeight) / 2
	leftPadding := (termWidth - modalWidth) / 2
	
	if topPadding < 0 {
		topPadding = 0
	}
	if leftPadding < 0 {
		leftPadding = 0
	}
	
	// Add padding
	var result []string
	
	// Top padding
	for i := 0; i < topPadding; i++ {
		result = append(result, "")
	}
	
	// Modal with left padding
	leftPad := strings.Repeat(" ", leftPadding)
	for _, line := range modalLines {
		result = append(result, leftPad+line)
	}
	
	return strings.Join(result, "\n")
}
