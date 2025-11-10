package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// DemoVaporwaveDesign demonstrates the vaporwave design system
func DemoVaporwaveDesign(width, height int) string {
	theme := VaporwaveTheme()
	
	// Create header
	header := RenderVaporwaveHeader(width, "AGENT STACK CONTROLLER", theme, 0)
	
	// Create demo sections
	sections := []string{
		demoColorPalette(theme),
		demoTypography(theme),
		demoBorders(theme),
		demoIndicators(theme),
		demoAnimations(theme, 0),
	}
	
	// Stack sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	
	// Create footer
	keybindings := map[string]string{
		"q": "quit",
		"r": "refresh",
		"t": "test",
	}
	footer := RenderVaporwaveFooter(width, keybindings, theme)
	
	// Combine all
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// demoColorPalette demonstrates the color palette
func demoColorPalette(theme Theme) string {
	title := RenderH2("Color Palette", theme)
	
	swatchStyle := lipgloss.NewStyle().
		Width(15).
		Height(3).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true)
	
	swatches := []string{
		swatchStyle.Background(theme.NeonPink).Foreground(theme.Background).Render("Neon Pink"),
		swatchStyle.Background(theme.ElectricBlue).Foreground(theme.Background).Render("Electric Blue"),
		swatchStyle.Background(theme.Purple).Foreground(theme.Background).Render("Purple"),
		swatchStyle.Background(theme.Cyan).Foreground(theme.Background).Render("Cyan"),
		swatchStyle.Background(theme.SunsetOrange).Foreground(theme.Background).Render("Sunset Orange"),
	}
	
	swatchRow := lipgloss.JoinHorizontal(lipgloss.Top, swatches...)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		swatchRow,
		"",
	)
}

// demoTypography demonstrates typography styles
func demoTypography(theme Theme) string {
	title := RenderH2("Typography", theme)
	
	examples := []string{
		RenderH1("H1 Header with Gradient", theme),
		RenderH2("H2 Header with Bold", theme),
		RenderBody("Body text for regular content", theme),
		RenderCaption("Caption text for secondary information", theme),
		RenderCode("code_snippet_123", theme),
		RenderNeonText("Neon glowing text", theme),
	}
	
	content := lipgloss.JoinVertical(lipgloss.Left, examples...)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		"",
	)
}

// demoBorders demonstrates border styles
func demoBorders(theme Theme) string {
	title := RenderH2("Border Styles", theme)
	
	// Create examples of different border styles
	roundedStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.Border).
		Width(20).
		Height(3).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)
	
	doubleStyle := lipgloss.NewStyle().
		Border(DoubleBorder()).
		BorderForeground(theme.Purple).
		Width(20).
		Height(3).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)
	
	thickStyle := lipgloss.NewStyle().
		Border(ThickBorder()).
		BorderForeground(theme.Cyan).
		Width(20).
		Height(3).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)
	
	borders := lipgloss.JoinHorizontal(
		lipgloss.Top,
		roundedStyle.Render("Rounded"),
		" ",
		doubleStyle.Render("Double"),
		" ",
		thickStyle.Render("Thick"),
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		borders,
		"",
	)
}

// demoIndicators demonstrates status indicators
func demoIndicators(theme Theme) string {
	title := RenderH2("Status Indicators", theme)
	
	phase := 0.0
	
	indicators := lipgloss.JoinHorizontal(
		lipgloss.Top,
		RenderGlowingOrb(IndicatorIdle, theme, phase)+" Idle   ",
		RenderGlowingOrb(IndicatorWorking, theme, phase)+" Working   ",
		RenderGlowingOrb(IndicatorError, theme, phase)+" Error   ",
		RenderGlowingOrb(IndicatorOffline, theme, phase)+" Offline",
	)
	
	// Progress bar
	progressLabel := RenderBody("Progress: ", theme)
	progressBar := RenderProgressBar(0.65, 40, theme)
	progress := progressLabel + progressBar
	
	// Task badges
	badges := lipgloss.JoinHorizontal(
		lipgloss.Top,
		RenderTaskStatusBadge("open", theme)+" ",
		RenderTaskStatusBadge("in_progress", theme)+" ",
		RenderTaskStatusBadge("completed", theme)+" ",
		RenderTaskStatusBadge("blocked", theme),
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		indicators,
		"",
		progress,
		"",
		badges,
		"",
	)
}

// demoAnimations demonstrates animation capabilities
func demoAnimations(theme Theme, frame int) string {
	title := RenderH2("Animations", theme)
	
	// Shimmer text
	shimmer := ShimmerText("Shimmering Text Effect", theme, frame)
	
	// Holographic text
	holographic := RenderHolographicText("Holographic Rainbow", theme, frame)
	
	// Loading spinner
	spinner := RenderLoadingSpinner(theme, frame) + " Loading..."
	
	// Pulsing dot
	pulse := RenderPulsingDot(theme, float64(frame)/10.0) + " Pulsing indicator"
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		shimmer,
		holographic,
		spinner,
		pulse,
		"",
	)
}

// DemoThemeComparison shows all themes side by side
func DemoThemeComparison(width int) string {
	themes := []Theme{
		VaporwaveTheme(),
		CyberpunkTheme(),
		MinimalTheme(),
	}
	
	var previews []string
	previewWidth := width / len(themes)
	
	for _, theme := range themes {
		preview := RenderThemePreview(theme, previewWidth-2)
		previews = append(previews, preview)
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, previews...)
}

// DemoResponsiveLayout demonstrates responsive layout
func DemoResponsiveLayout(width, height int) string {
	theme := VaporwaveTheme()
	breakpoint := GetBreakpoint(width)
	
	title := RenderH1(fmt.Sprintf("Responsive Layout - %s", breakpoint.Name), theme)
	
	info := fmt.Sprintf(
		"Width: %d | Height: %d | Breakpoint: %s (%d-%d)",
		width, height, breakpoint.Name, breakpoint.MinWidth, breakpoint.MaxWidth,
	)
	
	layout := GetResponsiveLayout(width, height)
	layoutInfo := fmt.Sprintf(
		"Columns: %d | Rows: %d | Spacing: %d",
		layout.Columns, layout.Rows, layout.Spacing,
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		RenderBody(info, theme),
		RenderBody(layoutInfo, theme),
	)
}

// DemoModalStyles demonstrates different modal styles
func DemoModalStyles(width, height int) string {
	theme := VaporwaveTheme()
	
	modalWidth := width / 3
	modalHeight := height / 2
	
	glassModal := CreateGlassModal(
		modalWidth,
		modalHeight,
		"Glass Modal",
		"This is a glass morphism modal with frosted effect.",
		theme,
	)
	
	neonModal := CreateNeonModal(
		modalWidth,
		modalHeight,
		"Neon Modal",
		"This is a neon modal with glowing borders.",
		theme,
	)
	
	minimalModal := CreateMinimalModal(
		modalWidth,
		modalHeight,
		"Minimal Modal",
		"This is a minimal, clean modal.",
		theme,
	)
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		glassModal,
		" ",
		neonModal,
		" ",
		minimalModal,
	)
}

// DemoPatterns demonstrates background patterns
func DemoPatterns(width, height int) string {
	theme := VaporwaveTheme()
	
	patternHeight := height / 3
	
	grid := RenderGridOverlay(width, patternHeight, theme)
	scanlines := RenderScanlines(width, patternHeight, theme, 0)
	starfield := RenderStarfield(width, patternHeight, theme, 0)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		RenderH2("Background Patterns", theme),
		grid,
		scanlines,
		starfield,
	)
}

// DemoPerformance demonstrates performance monitoring
func DemoPerformance() string {
	theme := VaporwaveTheme()
	
	monitor := NewPerformanceMonitor(60)
	monitor.StartFrame()
	// ... rendering work ...
	monitor.EndFrame()
	
	info := fmt.Sprintf(
		"FPS: %.1f | Frame Time: %v | Frame Count: %d",
		monitor.GetFPS(),
		monitor.GetFrameTime(),
		monitor.GetFrameCount(),
	)
	
	return RenderBody(info, theme)
}

// DemoFullInterface demonstrates a complete interface
func DemoFullInterface(width, height int, frame int) string {
	theme := VaporwaveTheme()
	
	// Header
	header := RenderVaporwaveHeader(width, "VAPORWAVE DESIGN SYSTEM", theme, frame)
	
	// Calculate layout
	layout := GetResponsiveLayout(width, height-6) // Reserve space for header/footer
	dims := CalculatePaneDimensions(width, height-6, layout)
	
	// Left pane - Agent status
	agentContent := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderGlowingOrb(IndicatorWorking, theme, float64(frame)/10.0)+" Agent-1: Working",
		RenderGlowingOrb(IndicatorIdle, theme, float64(frame)/10.0)+" Agent-2: Idle",
		RenderGlowingOrb(IndicatorError, theme, float64(frame)/10.0)+" Agent-3: Error",
	)
	leftPane := CreateGlowPaneStyle(dims.LeftWidth, dims.TopHeight+dims.BottomHeight, "Agents", theme).
		Render(agentContent)
	
	// Right top pane - Tasks
	taskContent := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTaskStatusBadge("in_progress", theme)+" Task #123: Implement feature",
		RenderTaskStatusBadge("open", theme)+" Task #124: Fix bug",
		RenderTaskStatusBadge("completed", theme)+" Task #125: Write tests",
	)
	rightTopPane := CreateGlowPaneStyle(dims.RightWidth, dims.TopHeight, "Tasks", theme).
		Render(taskContent)
	
	// Right bottom pane - Logs
	logContent := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderStatusText("info", "Agent-1 started task #123", theme),
		RenderStatusText("success", "Agent-2 completed task #125", theme),
		RenderStatusText("error", "Agent-3 encountered error", theme),
	)
	rightBottomPane := CreateGlowPaneStyle(dims.RightWidth, dims.BottomHeight, "Logs", theme).
		Render(logContent)
	
	// Combine right panes
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, rightTopPane, rightBottomPane)
	
	// Combine layout
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightColumn)
	
	// Footer
	keybindings := map[string]string{
		"q": "quit",
		"r": "refresh",
		"t": "test",
		"↑↓": "navigate",
	}
	connections := map[string]bool{
		"beads": true,
		"mcp":   true,
	}
	footer := RenderFooterWithStatus(width, keybindings, connections, theme, frame)
	
	// Combine all
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		footer,
	)
}
