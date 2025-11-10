package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderVaporwaveHeader renders a header with gradient and holographic effects
func RenderVaporwaveHeader(width int, title string, theme Theme, frame int) string {
	// Create gradient header bar
	headerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Create holographic title
	holographicTitle := RenderHolographicText(title, theme, frame)
	
	// Add decorative elements
	leftDecor := GeometricOrnament("triangle", theme)
	rightDecor := GeometricOrnament("triangle", theme)
	
	// Calculate spacing
	titleWidth := len(title) + 4 // Account for ornaments
	leftPad := (width - titleWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	
	headerContent := strings.Repeat(" ", leftPad) +
		leftDecor + " " +
		holographicTitle +
		" " + rightDecor
	
	return headerStyle.Render(headerContent)
}

// RenderGradientHeader renders a header with animated color shift
func RenderGradientHeader(width int, title string, theme Theme, frame int) string {
	// Create shifting gradient background
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], width)
	
	// Shift gradient based on frame
	shiftedGradient := make([]lipgloss.Color, width)
	for i := 0; i < width; i++ {
		shiftedGradient[i] = gradient[(i+frame)%len(gradient)]
	}
	
	// Create header with gradient background (simplified - use middle color)
	bgColor := shiftedGradient[width/2]
	
	headerStyle := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(theme.Foreground).
		Bold(true).
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Center)
	
	return headerStyle.Render(title)
}

// RenderScanlineHeader renders a header with scan-line animation
func RenderScanlineHeader(width int, title string, theme Theme, frame int) string {
	headerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Foreground(theme.Accent).
		Bold(true).
		Width(width).
		Padding(0, 1)
	
	// Add scan-line effect
	scanlinePos := frame % width
	
	// Create title with scan-line highlight
	titleRunes := []rune(title)
	result := ""
	
	titleStart := (width - len(title)) / 2
	
	for i := 0; i < width; i++ {
		if i >= titleStart && i < titleStart+len(title) {
			charIndex := i - titleStart
			char := string(titleRunes[charIndex])
			
			// Highlight character at scan-line position
			if i == scanlinePos || i == scanlinePos-1 {
				style := lipgloss.NewStyle().
					Foreground(theme.GlowCyan).
					Bold(true)
				result += style.Render(char)
			} else {
				result += char
			}
		} else {
			result += " "
		}
	}
	
	return headerStyle.Render(result)
}

// RenderVaporwaveFooter renders a footer with keybindings and neon highlights
func RenderVaporwaveFooter(width int, keybindings map[string]string, theme Theme) string {
	footerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Foreground(theme.Foreground).
		Width(width).
		Padding(0, 1)
	
	// Build keybindings with neon highlights
	var bindings []string
	for key, desc := range keybindings {
		binding := RenderKeyBinding(key, desc, theme)
		bindings = append(bindings, binding)
	}
	
	keybindingsStr := strings.Join(bindings, " | ")
	
	return footerStyle.Render(keybindingsStr)
}

// RenderFooterWithStatus renders a footer with keybindings and connection status
func RenderFooterWithStatus(width int, keybindings map[string]string, connections map[string]bool, theme Theme, frame int) string {
	footerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Build keybindings section
	var bindings []string
	for key, desc := range keybindings {
		binding := RenderKeyBinding(key, desc, theme)
		bindings = append(bindings, binding)
	}
	keybindingsStr := strings.Join(bindings, " | ")
	
	// Build connection status section
	var statuses []string
	for name, connected := range connections {
		status := RenderConnectionStatus(connected, theme, frame)
		statuses = append(statuses, name+": "+status)
	}
	statusStr := strings.Join(statuses, " | ")
	
	// Calculate spacing
	keybindingsWidth := lipgloss.Width(keybindingsStr)
	statusWidth := lipgloss.Width(statusStr)
	spacerWidth := width - keybindingsWidth - statusWidth - 4
	
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	
	spacer := strings.Repeat(" ", spacerWidth)
	
	footerContent := keybindingsStr + spacer + statusStr
	
	return footerStyle.Render(footerContent)
}

// RenderTimestamp renders a timestamp with elegant formatting
func RenderTimestamp(t time.Time, theme Theme) string {
	timestamp := t.Format("15:04:05")
	
	style := lipgloss.NewStyle().
		Foreground(theme.Muted)
	
	return style.Render(timestamp)
}

// RenderNotificationBadgeInFooter renders a notification badge in footer
func RenderNotificationBadgeInFooter(count int, theme Theme) string {
	if count <= 0 {
		return ""
	}
	
	badge := RenderNotificationBadge(count, theme)
	
	label := lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Render("Notifications: ")
	
	return label + badge
}

// RenderHolographicFooter renders a footer with holographic shimmer
func RenderHolographicFooter(width int, content string, theme Theme, frame int) string {
	footerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Apply holographic effect to content
	holographicContent := RenderHolographicText(content, theme, frame)
	
	return footerStyle.Render(holographicContent)
}

// RenderGlowingFooter renders a footer with glowing elements
func RenderGlowingFooter(width int, content string, theme Theme, phase float64) string {
	footerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Apply glow effect
	glowColor := PulseColor(theme.Accent, phase)
	
	contentStyle := lipgloss.NewStyle().
		Foreground(glowColor).
		Bold(true)
	
	return footerStyle.Render(contentStyle.Render(content))
}

// RenderProgressFooter renders a footer with progress indicator
func RenderProgressFooter(width int, label string, progress float64, theme Theme) string {
	footerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Create label
	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)
	
	// Create progress bar
	barWidth := width - len(label) - 10
	if barWidth < 10 {
		barWidth = 10
	}
	
	progressBar := RenderProgressBar(progress, barWidth, theme)
	
	// Create percentage
	percentage := fmt.Sprintf("%.0f%%", progress*100)
	percentStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)
	
	content := labelStyle.Render(label) + " " +
		progressBar + " " +
		percentStyle.Render(percentage)
	
	return footerStyle.Render(content)
}

// RenderAnimatedBorder renders an animated border for header/footer
func RenderAnimatedBorder(width int, theme Theme, frame int) string {
	// Create animated border using gradient
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], width)
	
	var border strings.Builder
	for i := 0; i < width; i++ {
		colorIndex := (i + frame) % len(gradient)
		style := lipgloss.NewStyle().Foreground(gradient[colorIndex])
		border.WriteString(style.Render("─"))
	}
	
	return border.String()
}

// RenderHeaderWithSubtitle renders a header with title and subtitle
func RenderHeaderWithSubtitle(width int, title, subtitle string, theme Theme) string {
	headerStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(1, 2)
	
	// Create title with gradient
	titleGradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(title))
	gradientTitle := ApplyGradientToText(title, titleGradient)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(width - 4)
	
	// Create subtitle
	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.Muted).
		Align(lipgloss.Center).
		Width(width - 4)
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(gradientTitle),
		subtitleStyle.Render(subtitle),
	)
	
	return headerStyle.Render(content)
}

// RenderStatusBar renders a status bar with multiple sections
func RenderStatusBar(width int, sections []StatusSection, theme Theme) string {
	statusStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Calculate section widths
	sectionWidth := width / len(sections)
	
	var renderedSections []string
	for _, section := range sections {
		sectionStyle := lipgloss.NewStyle().
			Width(sectionWidth).
			Align(lipgloss.Center)
		
		// Apply color based on section type
		var color lipgloss.Color
		switch section.Type {
		case "success":
			color = theme.Success
		case "warning":
			color = theme.Warning
		case "error":
			color = theme.Error
		case "info":
			color = theme.Info
		default:
			color = theme.Foreground
		}
		
		contentStyle := lipgloss.NewStyle().Foreground(color)
		content := contentStyle.Render(section.Content)
		
		renderedSections = append(renderedSections, sectionStyle.Render(content))
	}
	
	statusContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedSections...)
	
	return statusStyle.Render(statusContent)
}

// StatusSection represents a section in the status bar
type StatusSection struct {
	Type    string
	Content string
}

// RenderBreadcrumb renders a breadcrumb navigation
func RenderBreadcrumb(items []string, theme Theme) string {
	var parts []string
	
	for i, item := range items {
		if i > 0 {
			separator := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Render(" › ")
			parts = append(parts, separator)
		}
		
		var style lipgloss.Style
		if i == len(items)-1 {
			// Current item - highlight
			style = lipgloss.NewStyle().
				Foreground(theme.Accent).
				Bold(true)
		} else {
			// Previous items - muted
			style = lipgloss.NewStyle().
				Foreground(theme.Muted)
		}
		
		parts = append(parts, style.Render(item))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// RenderTabBar renders a tab bar
func RenderTabBar(width int, tabs []string, activeTab int, theme Theme) string {
	tabBarStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	var renderedTabs []string
	
	for i, tab := range tabs {
		var style lipgloss.Style
		
		if i == activeTab {
			// Active tab
			style = lipgloss.NewStyle().
				Foreground(theme.Accent).
				Background(theme.MidnightBlue).
				Bold(true).
				Padding(0, 2)
		} else {
			// Inactive tab
			style = lipgloss.NewStyle().
				Foreground(theme.Muted).
				Padding(0, 2)
		}
		
		renderedTabs = append(renderedTabs, style.Render(tab))
	}
	
	tabsContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	return tabBarStyle.Render(tabsContent)
}

// RenderToolbar renders a toolbar with buttons
func RenderToolbar(width int, buttons []ToolbarButton, theme Theme) string {
	toolbarStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	var renderedButtons []string
	
	for _, button := range buttons {
		var style lipgloss.Style
		
		if button.Active {
			style = lipgloss.NewStyle().
				Foreground(theme.Background).
				Background(theme.Accent).
				Bold(true).
				Padding(0, 1)
		} else if button.Disabled {
			style = lipgloss.NewStyle().
				Foreground(theme.Muted).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(theme.Foreground).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Border).
				Padding(0, 1)
		}
		
		renderedButtons = append(renderedButtons, style.Render(button.Label))
	}
	
	buttonsContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedButtons...)
	
	return toolbarStyle.Render(buttonsContent)
}

// ToolbarButton represents a button in the toolbar
type ToolbarButton struct {
	Label    string
	Active   bool
	Disabled bool
}

// RenderSearchBar renders a search bar
func RenderSearchBar(width int, query string, focused bool, theme Theme) string {
	searchStyle := lipgloss.NewStyle().
		Background(theme.DeepPurple).
		Width(width).
		Padding(0, 1)
	
	// Create search icon
	iconStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	icon := iconStyle.Render("🔍 ")
	
	// Create input field
	var inputStyle lipgloss.Style
	if focused {
		inputStyle = lipgloss.NewStyle().
			Foreground(theme.Cyan).
			Background(theme.MidnightBlue).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Accent).
			Width(width - 8).
			Padding(0, 1)
	} else {
		inputStyle = lipgloss.NewStyle().
			Foreground(theme.Foreground).
			Background(theme.MidnightBlue).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Width(width - 8).
			Padding(0, 1)
	}
	
	cursor := ""
	if focused {
		cursor = "│"
	}
	
	input := inputStyle.Render(query + cursor)
	
	content := icon + input
	
	return searchStyle.Render(content)
}
