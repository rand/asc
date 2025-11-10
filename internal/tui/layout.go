package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Golden ratio constant for proportions
const GoldenRatio = 1.618

// Grid system base unit (8px equivalent in terminal)
const GridUnit = 1

// LayoutConfig represents layout configuration
type LayoutConfig struct {
	Width  int
	Height int
	Theme  Theme
}

// Breakpoint represents a responsive breakpoint
type Breakpoint struct {
	MinWidth int
	MaxWidth int
	Name     string
}

// Common breakpoints
var (
	BreakpointSmall  = Breakpoint{MinWidth: 0, MaxWidth: 80, Name: "small"}
	BreakpointMedium = Breakpoint{MinWidth: 81, MaxWidth: 120, Name: "medium"}
	BreakpointLarge  = Breakpoint{MinWidth: 121, MaxWidth: 160, Name: "large"}
	BreakpointXLarge = Breakpoint{MinWidth: 161, MaxWidth: 999999, Name: "xlarge"}
)

// GetBreakpoint returns the current breakpoint based on width
func GetBreakpoint(width int) Breakpoint {
	if width <= BreakpointSmall.MaxWidth {
		return BreakpointSmall
	} else if width <= BreakpointMedium.MaxWidth {
		return BreakpointMedium
	} else if width <= BreakpointLarge.MaxWidth {
		return BreakpointLarge
	}
	return BreakpointXLarge
}

// CalculateGoldenRatio calculates dimensions using golden ratio
func CalculateGoldenRatio(total int, larger bool) int {
	if larger {
		return int(float64(total) * GoldenRatio / (GoldenRatio + 1))
	}
	return int(float64(total) / (GoldenRatio + 1))
}

// ApplyGridSpacing applies grid-based spacing
func ApplyGridSpacing(value int) int {
	return value * GridUnit
}

// CalculatePadding calculates padding based on grid system
func CalculatePadding(level int) (vertical, horizontal int) {
	// Level 0: no padding
	// Level 1: 1 grid unit (1 char)
	// Level 2: 2 grid units (2 chars)
	// etc.
	return ApplyGridSpacing(level), ApplyGridSpacing(level * 2)
}

// CalculateMargin calculates margin based on grid system
func CalculateMargin(level int) int {
	return ApplyGridSpacing(level)
}

// ResponsiveLayout represents a responsive layout configuration
type ResponsiveLayout struct {
	Breakpoint Breakpoint
	Columns    int
	Rows       int
	Spacing    int
}

// GetResponsiveLayout returns layout configuration for current breakpoint
func GetResponsiveLayout(width, height int) ResponsiveLayout {
	breakpoint := GetBreakpoint(width)
	
	switch breakpoint.Name {
	case "small":
		// Single column layout for small screens
		return ResponsiveLayout{
			Breakpoint: breakpoint,
			Columns:    1,
			Rows:       3,
			Spacing:    1,
		}
	case "medium":
		// Two column layout for medium screens
		return ResponsiveLayout{
			Breakpoint: breakpoint,
			Columns:    2,
			Rows:       2,
			Spacing:    2,
		}
	case "large", "xlarge":
		// Three pane layout for large screens
		return ResponsiveLayout{
			Breakpoint: breakpoint,
			Columns:    3,
			Rows:       2,
			Spacing:    2,
		}
	default:
		return ResponsiveLayout{
			Breakpoint: breakpoint,
			Columns:    2,
			Rows:       2,
			Spacing:    2,
		}
	}
}

// CalculatePaneDimensions calculates pane dimensions with golden ratio
func CalculatePaneDimensions(totalWidth, totalHeight int, layout ResponsiveLayout) PaneDimensions {
	spacing := ApplyGridSpacing(layout.Spacing)
	
	// Calculate available space after spacing
	availableWidth := totalWidth - (spacing * (layout.Columns - 1))
	availableHeight := totalHeight - (spacing * (layout.Rows - 1))
	
	// Use golden ratio for proportions
	leftWidth := CalculateGoldenRatio(availableWidth, false)
	rightWidth := availableWidth - leftWidth
	
	topHeight := availableHeight / 2
	bottomHeight := availableHeight - topHeight
	
	return PaneDimensions{
		LeftWidth:     leftWidth,
		RightWidth:    rightWidth,
		TopHeight:     topHeight,
		BottomHeight:  bottomHeight,
		Spacing:       spacing,
	}
}

// PaneDimensions represents calculated pane dimensions
type PaneDimensions struct {
	LeftWidth    int
	RightWidth   int
	TopHeight    int
	BottomHeight int
	Spacing      int
}

// CreateResponsivePane creates a pane with responsive dimensions
func CreateResponsivePane(width, height int, content string, theme Theme) string {
	// Calculate padding based on available space
	paddingLevel := 1
	if width > 100 {
		paddingLevel = 2
	}
	
	vPad, hPad := CalculatePadding(paddingLevel)
	
	paneStyle := lipgloss.NewStyle().
		Border(VaporwaveBorder()).
		BorderForeground(theme.Border).
		Width(width - 2).
		Height(height - 2).
		Padding(vPad, hPad)
	
	return paneStyle.Render(content)
}

// HandleOverflow handles content overflow with fade-out effect
func HandleOverflow(content string, maxHeight int, theme Theme) string {
	lines := lipgloss.Height(content)
	
	if lines <= maxHeight {
		return content
	}
	
	// Truncate and add fade effect
	contentLines := strings.Split(content, "\n")
	visibleLines := contentLines[:maxHeight-1]
	
	// Add fade indicator
	fadeStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)
	
	visibleLines = append(visibleLines, fadeStyle.Render("..."))
	
	return lipgloss.JoinVertical(lipgloss.Left, visibleLines...)
}

// AutoScale scales text based on available space
func AutoScale(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	
	// Truncate with ellipsis
	return TruncateText(text, maxWidth)
}

// AddBreathingRoom adds whitespace for better readability
func AddBreathingRoom(content string, theme Theme) string {
	// Add margin around content
	marginStyle := lipgloss.NewStyle().
		Margin(1, 2)
	
	return marginStyle.Render(content)
}

// CreateFlexLayout creates a flexible layout that adapts to content
func CreateFlexLayout(items []string, width int, direction string) string {
	if direction == "horizontal" {
		return lipgloss.JoinHorizontal(lipgloss.Top, items...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

// DistributeSpace distributes space evenly among items
func DistributeSpace(totalSpace int, numItems int, spacing int) []int {
	if numItems == 0 {
		return []int{}
	}
	
	// Calculate space per item
	totalSpacing := spacing * (numItems - 1)
	availableSpace := totalSpace - totalSpacing
	
	itemSpace := availableSpace / numItems
	remainder := availableSpace % numItems
	
	spaces := make([]int, numItems)
	for i := 0; i < numItems; i++ {
		spaces[i] = itemSpace
		if i < remainder {
			spaces[i]++
		}
	}
	
	return spaces
}

// CreateGridLayout creates a grid layout
func CreateGridLayout(items []string, columns int, width int, theme Theme) string {
	if len(items) == 0 {
		return ""
	}
	
	// Calculate column width
	spacing := ApplyGridSpacing(2)
	columnWidth := (width - (spacing * (columns - 1))) / columns
	
	var rows []string
	var currentRow []string
	
	for i, item := range items {
		// Create cell with fixed width
		cellStyle := lipgloss.NewStyle().
			Width(columnWidth).
			Padding(0, 1)
		
		cell := cellStyle.Render(item)
		currentRow = append(currentRow, cell)
		
		// Start new row when we reach column limit
		if (i+1)%columns == 0 || i == len(items)-1 {
			row := lipgloss.JoinHorizontal(lipgloss.Top, currentRow...)
			rows = append(rows, row)
			currentRow = []string{}
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// AnimateLayoutTransition animates a layout transition
func AnimateLayoutTransition(oldLayout, newLayout string, progress float64) string {
	// Simple transition - in a real implementation, we'd blend the layouts
	if progress < 0.5 {
		return oldLayout
	}
	return newLayout
}

// CalculateAspectRatio calculates aspect ratio
func CalculateAspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// MaintainAspectRatio maintains aspect ratio when resizing
func MaintainAspectRatio(originalWidth, originalHeight, newWidth int) int {
	aspectRatio := CalculateAspectRatio(originalWidth, originalHeight)
	return int(float64(newWidth) / aspectRatio)
}

// CreateSplitLayout creates a split layout (horizontal or vertical)
func CreateSplitLayout(left, right string, width int, ratio float64, direction string, theme Theme) string {
	if direction == "horizontal" {
		// Horizontal split
		leftWidth := int(float64(width) * ratio)
		rightWidth := width - leftWidth
		
		leftStyle := lipgloss.NewStyle().Width(leftWidth)
		rightStyle := lipgloss.NewStyle().Width(rightWidth)
		
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render(left),
			rightStyle.Render(right),
		)
	}
	
	// Vertical split
	return lipgloss.JoinVertical(
		lipgloss.Left,
		left,
		right,
	)
}

// CreateStackLayout creates a stacked layout
func CreateStackLayout(items []string, spacing int) string {
	if len(items) == 0 {
		return ""
	}
	
	// Add spacing between items
	var spacedItems []string
	for i, item := range items {
		spacedItems = append(spacedItems, item)
		if i < len(items)-1 {
			// Add spacing
			spacedItems = append(spacedItems, strings.Repeat("\n", spacing))
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, spacedItems...)
}

// CreateCenteredLayout creates a centered layout
func CreateCenteredLayout(content string, width, height int) string {
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)
	
	// Calculate centering
	leftPad := (width - contentWidth) / 2
	topPad := (height - contentHeight) / 2
	
	if leftPad < 0 {
		leftPad = 0
	}
	if topPad < 0 {
		topPad = 0
	}
	
	centeredStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
	
	return centeredStyle.Render(content)
}

// CreateScrollablePane creates a pane with scrolling support
func CreateScrollablePane(content string, width, height int, scrollOffset int, theme Theme) string {
	lines := strings.Split(content, "\n")
	
	// Calculate visible range
	startLine := scrollOffset
	endLine := scrollOffset + height
	
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	visibleLines := lines[startLine:endLine]
	
	// Add scroll indicators
	var indicators []string
	if startLine > 0 {
		upIndicator := lipgloss.NewStyle().
			Foreground(theme.Accent).
			Align(lipgloss.Center).
			Width(width).
			Render("▲")
		indicators = append(indicators, upIndicator)
	}
	
	indicators = append(indicators, visibleLines...)
	
	if endLine < len(lines) {
		downIndicator := lipgloss.NewStyle().
			Foreground(theme.Accent).
			Align(lipgloss.Center).
			Width(width).
			Render("▼")
		indicators = append(indicators, downIndicator)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, indicators...)
}

// CalculateOptimalColumns calculates optimal number of columns for content
func CalculateOptimalColumns(itemCount, width int) int {
	// Use square root as a heuristic
	optimalCols := int(math.Sqrt(float64(itemCount)))
	
	// Ensure at least 1 column
	if optimalCols < 1 {
		optimalCols = 1
	}
	
	// Ensure columns fit in width (assuming min 20 chars per column)
	maxCols := width / 20
	if optimalCols > maxCols {
		optimalCols = maxCols
	}
	
	return optimalCols
}

// CreateMasonryLayout creates a masonry-style layout
func CreateMasonryLayout(items []string, columns int, width int, theme Theme) string {
	if len(items) == 0 || columns == 0 {
		return ""
	}
	
	// Distribute items across columns
	columnItems := make([][]string, columns)
	for i, item := range items {
		col := i % columns
		columnItems[col] = append(columnItems[col], item)
	}
	
	// Create columns
	var renderedColumns []string
	columnWidth := width / columns
	
	for _, items := range columnItems {
		column := lipgloss.JoinVertical(lipgloss.Left, items...)
		columnStyle := lipgloss.NewStyle().Width(columnWidth)
		renderedColumns = append(renderedColumns, columnStyle.Render(column))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
}

// CreateAdaptiveLayout creates a layout that adapts to content size
func CreateAdaptiveLayout(items []string, width, height int, theme Theme) string {
	breakpoint := GetBreakpoint(width)
	
	switch breakpoint.Name {
	case "small":
		// Stack vertically for small screens
		return CreateStackLayout(items, 1)
	case "medium":
		// Two columns for medium screens
		return CreateGridLayout(items, 2, width, theme)
	case "large", "xlarge":
		// Three columns for large screens
		return CreateGridLayout(items, 3, width, theme)
	default:
		return CreateStackLayout(items, 1)
	}
}

// Helper function to split lines
func splitLines(content string) []string {
	return strings.Split(content, "\n")
}
