package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderGridOverlay renders a subtle grid pattern
func RenderGridOverlay(width, height int, theme Theme) string {
	var lines []string
	
	gridStyle := lipgloss.NewStyle().Foreground(theme.DeepPurple)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Create grid at intervals
			if x%10 == 0 || y%5 == 0 {
				if x%10 == 0 && y%5 == 0 {
					line.WriteString(gridStyle.Render("┼"))
				} else if x%10 == 0 {
					line.WriteString(gridStyle.Render("│"))
				} else {
					line.WriteString(gridStyle.Render("─"))
				}
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderPerspectiveGrid renders a perspective grid with vanishing point
func RenderPerspectiveGrid(width, height int, theme Theme) string {
	var lines []string
	
	// Vanishing point at center
	centerX := width / 2
	centerY := height / 2
	
	gridStyle := lipgloss.NewStyle().Foreground(theme.Purple)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Calculate distance from center
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)
			
			// Create perspective effect
			if int(distance)%8 == 0 {
				line.WriteString(gridStyle.Render("·"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderGeometricShapes renders geometric shapes as accents
func RenderGeometricShapes(width, height int, theme Theme, frame int) string {
	var lines []string
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Place shapes at specific positions with animation
			if (x+frame)%20 == 0 && y%10 == 0 {
				// Triangle
				style := lipgloss.NewStyle().Foreground(theme.Purple)
				line.WriteString(style.Render("◢"))
			} else if (x+frame)%15 == 0 && y%8 == 0 {
				// Diamond
				style := lipgloss.NewStyle().Foreground(theme.Cyan)
				line.WriteString(style.Render("◆"))
			} else if (x+frame)%25 == 0 && y%12 == 0 {
				// Hexagon
				style := lipgloss.NewStyle().Foreground(theme.NeonPink)
				line.WriteString(style.Render("⬡"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderScanlines renders a scanline effect for retro-futuristic feel
func RenderScanlines(width, height int, theme Theme, frame int) string {
	var lines []string
	
	scanlineStyle := lipgloss.NewStyle().Foreground(theme.DeepPurple)
	
	for y := 0; y < height; y++ {
		// Animate scanlines moving down
		if (y+frame)%3 == 0 {
			// Scanline
			lines = append(lines, scanlineStyle.Render(strings.Repeat("▔", width)))
		} else {
			lines = append(lines, strings.Repeat(" ", width))
		}
	}
	
	return strings.Join(lines, "\n")
}

// RenderNoiseTexture renders a subtle noise/grain texture
func RenderNoiseTexture(width, height int, theme Theme, seed int) string {
	var lines []string
	
	noiseChars := []string{"·", "∙", "•", " ", " ", " "}
	noiseStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Pseudo-random noise based on position and seed
			index := (x*7 + y*13 + seed) % len(noiseChars)
			char := noiseChars[index]
			line.WriteString(noiseStyle.Render(char))
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderFloatingShapes renders animated floating shapes
func RenderFloatingShapes(width, height int, theme Theme, frame int) string {
	var lines []string
	
	shapes := []string{"◆", "◇", "○", "◎", "△", "▽"}
	colors := []lipgloss.Color{
		theme.NeonPink,
		theme.Purple,
		theme.Cyan,
		theme.ElectricBlue,
	}
	
	// Initialize empty grid
	for y := 0; y < height; y++ {
		lines = append(lines, strings.Repeat(" ", width))
	}
	
	// Place floating shapes
	numShapes := 5
	for i := 0; i < numShapes; i++ {
		// Calculate position with animation
		x := (i*17 + frame) % width
		y := (i*11 + frame/2) % height
		
		shape := shapes[i%len(shapes)]
		color := colors[i%len(colors)]
		
		style := lipgloss.NewStyle().Foreground(color)
		
		// Replace character at position
		if y < len(lines) && x < len(lines[y]) {
			lineRunes := []rune(lines[y])
			if x < len(lineRunes) {
				lineRunes[x] = []rune(style.Render(shape))[0]
				lines[y] = string(lineRunes)
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

// RenderParticleEffect renders particle effects
func RenderParticleEffect(width, height int, theme Theme, frame int) string {
	var lines []string
	
	particleChars := []string{"·", "∙", "•", "✦", "✧"}
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Pseudo-random particles with animation
			value := (x*3 + y*7 + frame) % 100
			
			if value < 5 {
				// Show particle
				charIndex := (x + y + frame) % len(particleChars)
				char := particleChars[charIndex]
				
				// Gradient color based on position
				gradient := GenerateGradient(theme.Cyan, theme.Purple, width)
				style := lipgloss.NewStyle().Foreground(gradient[x%len(gradient)])
				
				line.WriteString(style.Render(char))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderParallaxLayer renders a parallax layer for depth
func RenderParallaxLayer(width, height int, theme Theme, frame int, speed float64) string {
	var lines []string
	
	offset := int(float64(frame) * speed)
	
	layerStyle := lipgloss.NewStyle().Foreground(theme.DeepPurple)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Create pattern with offset
			if (x+offset)%15 == 0 {
				line.WriteString(layerStyle.Render("│"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderHexagonPattern renders a hexagon pattern
func RenderHexagonPattern(width, height int, theme Theme) string {
	var lines []string
	
	hexStyle := lipgloss.NewStyle().Foreground(theme.Purple)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Hexagonal grid pattern
			if y%4 == 0 && x%6 == 0 {
				line.WriteString(hexStyle.Render("⬡"))
			} else if y%4 == 2 && (x+3)%6 == 0 {
				line.WriteString(hexStyle.Render("⬡"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderTrianglePattern renders a triangle pattern
func RenderTrianglePattern(width, height int, theme Theme) string {
	var lines []string
	
	triangleStyle := lipgloss.NewStyle().Foreground(theme.Cyan)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Triangle pattern
			if (x+y)%8 == 0 {
				line.WriteString(triangleStyle.Render("△"))
			} else if (x-y)%8 == 0 {
				line.WriteString(triangleStyle.Render("▽"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderCirclePattern renders a circle pattern
func RenderCirclePattern(width, height int, theme Theme) string {
	var lines []string
	
	circleStyle := lipgloss.NewStyle().Foreground(theme.ElectricBlue)
	
	centerX := width / 2
	centerY := height / 2
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Calculate distance from center
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)
			
			// Draw concentric circles
			if int(distance)%8 == 0 {
				line.WriteString(circleStyle.Render("○"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderWavePattern renders a wave pattern
func RenderWavePattern(width, height int, theme Theme, frame int) string {
	var lines []string
	
	waveStyle := lipgloss.NewStyle().Foreground(theme.Cyan)
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Calculate wave
			phase := float64(x+frame) / 5.0
			waveY := int(float64(height/2) + 5*math.Sin(phase))
			
			if y == waveY {
				line.WriteString(waveStyle.Render("~"))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderStarfield renders a starfield background
func RenderStarfield(width, height int, theme Theme, frame int) string {
	var lines []string
	
	stars := []string{"·", "∙", "•", "✦", "✧", "⋆"}
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Pseudo-random stars with twinkling
			value := (x*13 + y*17 + frame/3) % 200
			
			if value < 3 {
				// Show star
				starIndex := (x + y) % len(stars)
				star := stars[starIndex]
				
				// Vary brightness
				brightness := float64((frame+x+y)%10) / 10.0
				color := ColorInterpolate(theme.DeepPurple, theme.Foreground, brightness)
				
				style := lipgloss.NewStyle().Foreground(color)
				line.WriteString(style.Render(star))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// RenderMatrixRain renders a matrix-style rain effect
func RenderMatrixRain(width, height int, theme Theme, frame int) string {
	var lines []string
	
	chars := []string{"0", "1", "▀", "▄", "█", "░", "▒", "▓"}
	
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			// Create falling effect
			value := (x*7 + frame - y*3) % 100
			
			if value < 20 {
				charIndex := (x + frame) % len(chars)
				char := chars[charIndex]
				
				// Fade based on position
				fade := float64(value) / 20.0
				color := ColorInterpolate(theme.DeepPurple, theme.Cyan, fade)
				
				style := lipgloss.NewStyle().Foreground(color)
				line.WriteString(style.Render(char))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// OverlayPattern overlays a pattern on top of content with transparency
func OverlayPattern(content, pattern string, opacity float64) string {
	// Simple overlay - just return content for now
	// In a real implementation, we'd blend the two
	return content
}

// RenderGradientBackground renders a gradient background
func RenderGradientBackground(width, height int, theme Theme, direction string) string {
	var lines []string
	
	var gradient []lipgloss.Color
	
	switch direction {
	case "horizontal":
		gradient = GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], width)
		for y := 0; y < height; y++ {
			var line strings.Builder
			for x := 0; x < width; x++ {
				style := lipgloss.NewStyle().Background(gradient[x])
				line.WriteString(style.Render(" "))
			}
			lines = append(lines, line.String())
		}
	case "vertical":
		gradient = GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], height)
		for y := 0; y < height; y++ {
			style := lipgloss.NewStyle().Background(gradient[y])
			lines = append(lines, style.Render(strings.Repeat(" ", width)))
		}
	case "diagonal":
		gradient = GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], width+height)
		for y := 0; y < height; y++ {
			var line strings.Builder
			for x := 0; x < width; x++ {
				index := (x + y) % len(gradient)
				style := lipgloss.NewStyle().Background(gradient[index])
				line.WriteString(style.Render(" "))
			}
			lines = append(lines, line.String())
		}
	default:
		// Radial gradient from center
		centerX := width / 2
		centerY := height / 2
		maxDist := math.Sqrt(float64(centerX*centerX + centerY*centerY))
		gradient = GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], int(maxDist))
		
		for y := 0; y < height; y++ {
			var line strings.Builder
			for x := 0; x < width; x++ {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				dist := math.Sqrt(dx*dx + dy*dy)
				index := int(dist) % len(gradient)
				
				style := lipgloss.NewStyle().Background(gradient[index])
				line.WriteString(style.Render(" "))
			}
			lines = append(lines, line.String())
		}
	}
	
	return strings.Join(lines, "\n")
}
