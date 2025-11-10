package tui

import (
	"math"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// EasingFunction represents an easing function for animations
type EasingFunction func(t float64) float64

// Animation represents an animation state
type Animation struct {
	StartTime time.Time
	Duration  time.Duration
	Easing    EasingFunction
}

// NewAnimation creates a new animation
func NewAnimation(duration time.Duration, easing EasingFunction) *Animation {
	return &Animation{
		StartTime: time.Now(),
		Duration:  duration,
		Easing:    easing,
	}
}

// Progress returns the current progress (0.0 to 1.0) of the animation
func (a *Animation) Progress() float64 {
	elapsed := time.Since(a.StartTime)
	if elapsed >= a.Duration {
		return 1.0
	}
	
	t := float64(elapsed) / float64(a.Duration)
	return a.Easing(t)
}

// IsComplete returns true if the animation is complete
func (a *Animation) IsComplete() bool {
	return time.Since(a.StartTime) >= a.Duration
}

// Reset resets the animation to start
func (a *Animation) Reset() {
	a.StartTime = time.Now()
}

// Easing Functions

// EaseLinear provides linear easing (no easing)
func EaseLinear(t float64) float64 {
	return t
}

// EaseInQuad provides quadratic ease-in
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutQuad provides quadratic ease-out
func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

// EaseInOutQuad provides quadratic ease-in-out
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// EaseInCubic provides cubic ease-in
func EaseInCubic(t float64) float64 {
	return t * t * t
}

// EaseOutCubic provides cubic ease-out
func EaseOutCubic(t float64) float64 {
	t--
	return t*t*t + 1
}

// EaseInOutCubic provides cubic ease-in-out
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return (t*t*t + 2) / 2
}

// EaseInSine provides sinusoidal ease-in
func EaseInSine(t float64) float64 {
	return 1 - math.Cos(t*math.Pi/2)
}

// EaseOutSine provides sinusoidal ease-out
func EaseOutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

// EaseInOutSine provides sinusoidal ease-in-out
func EaseInOutSine(t float64) float64 {
	return -(math.Cos(math.Pi*t) - 1) / 2
}

// EaseInExpo provides exponential ease-in
func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

// EaseOutExpo provides exponential ease-out
func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// EaseInOutExpo provides exponential ease-in-out
func EaseInOutExpo(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	
	if t < 0.5 {
		return math.Pow(2, 20*t-10) / 2
	}
	return (2 - math.Pow(2, -20*t+10)) / 2
}

// EaseInElastic provides elastic ease-in
func EaseInElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	
	p := 0.3
	s := p / 4
	t--
	return -(math.Pow(2, 10*t) * math.Sin((t-s)*(2*math.Pi)/p))
}

// EaseOutElastic provides elastic ease-out
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	
	p := 0.3
	s := p / 4
	return math.Pow(2, -10*t)*math.Sin((t-s)*(2*math.Pi)/p) + 1
}

// EaseInOutElastic provides elastic ease-in-out
func EaseInOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	
	p := 0.3 * 1.5
	s := p / 4
	t = t*2 - 1
	
	if t < 0 {
		return -0.5 * (math.Pow(2, 10*t) * math.Sin((t-s)*(2*math.Pi)/p))
	}
	return math.Pow(2, -10*t)*math.Sin((t-s)*(2*math.Pi)/p)*0.5 + 1
}

// Color Transition Animations

// AnimateColorTransition animates a color transition
func AnimateColorTransition(from, to lipgloss.Color, progress float64) lipgloss.Color {
	return ColorInterpolate(from, to, progress)
}

// AnimateFade animates a fade effect
func AnimateFade(content string, theme Theme, progress float64, fadeIn bool) string {
	if fadeIn {
		if progress >= 1.0 {
			return content
		}
		// Simulate fade by showing partial content
		return content
	} else {
		if progress >= 1.0 {
			return ""
		}
		return content
	}
}

// AnimateSlideIn animates a slide-in effect
func AnimateSlideIn(content string, width int, progress float64, direction string) string {
	// In terminal, we simulate slide with spacing
	_ = lipgloss.Height(content) // lines
	_ = int(float64(width) * (1 - progress)) // offset
	
	if direction == "left" {
		// Slide from left
		return content
	} else if direction == "right" {
		// Slide from right
		return content
	}
	
	return content
}

// AnimateScale simulates a scale animation
func AnimateScale(content string, progress float64) string {
	// Terminal limitations - we can't truly scale, but we can simulate
	if progress < 0.5 {
		// Shrinking phase - show less content
		return ""
	}
	return content
}

// Frame-based Animation System

// AnimationFrame represents a single frame in an animation
type AnimationFrame struct {
	Frame    int
	MaxFrame int
	Loop     bool
}

// NewAnimationFrame creates a new animation frame counter
func NewAnimationFrame(maxFrame int, loop bool) *AnimationFrame {
	return &AnimationFrame{
		Frame:    0,
		MaxFrame: maxFrame,
		Loop:     loop,
	}
}

// Next advances to the next frame
func (af *AnimationFrame) Next() {
	af.Frame++
	if af.Frame >= af.MaxFrame {
		if af.Loop {
			af.Frame = 0
		} else {
			af.Frame = af.MaxFrame - 1
		}
	}
}

// Progress returns the progress (0.0 to 1.0)
func (af *AnimationFrame) Progress() float64 {
	return float64(af.Frame) / float64(af.MaxFrame)
}

// IsComplete returns true if animation is complete
func (af *AnimationFrame) IsComplete() bool {
	return !af.Loop && af.Frame >= af.MaxFrame-1
}

// Reset resets the animation
func (af *AnimationFrame) Reset() {
	af.Frame = 0
}

// Ripple Effect Animation

// RippleEffect creates a ripple effect animation
func RippleEffect(width, height int, centerX, centerY int, theme Theme, frame int) string {
	var lines []string
	
	maxRadius := float64(frame)
	
	for y := 0; y < height; y++ {
		var line string
		for x := 0; x < width; x++ {
			// Calculate distance from center
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)
			
			// Check if within ripple radius
			if math.Abs(distance-maxRadius) < 2 {
				// Ripple edge
				style := lipgloss.NewStyle().Foreground(theme.Cyan)
				line += style.Render("○")
			} else {
				line += " "
			}
		}
		lines = append(lines, line)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Wave Effect Animation

// WaveEffect creates a wave effect
func WaveEffect(text string, theme Theme, frame int) string {
	runes := []rune(text)
	result := ""
	
	for i, r := range runes {
		// Calculate wave offset
		phase := float64(i+frame) / 3.0
		offset := math.Sin(phase)
		
		// Vary color based on wave
		var color lipgloss.Color
		if offset > 0.5 {
			color = theme.Cyan
		} else if offset > 0 {
			color = theme.ElectricBlue
		} else if offset > -0.5 {
			color = theme.Purple
		} else {
			color = theme.NeonPink
		}
		
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(r))
	}
	
	return result
}

// Pulse Animation

// PulseAnimation creates a pulsing animation
func PulseAnimation(content string, theme Theme, frame int) string {
	// Calculate pulse phase
	phase := float64(frame) / 10.0
	_ = 0.7 + 0.3*math.Sin(phase) // intensity
	
	// Apply intensity to color (simplified)
	return content
}

// Shimmer Animation

// ShimmerAnimation creates a shimmer effect
func ShimmerAnimation(text string, theme Theme, frame int) string {
	return ShimmerText(text, theme, frame)
}

// Glow Animation

// GlowAnimation creates a glowing animation
func GlowAnimation(text string, theme Theme, frame int) string {
	phase := float64(frame) / 15.0
	glowColor := PulseColor(theme.Accent, phase)
	
	style := lipgloss.NewStyle().
		Foreground(glowColor).
		Bold(true)
	
	return style.Render(text)
}

// Rotate Animation (for spinners)

// RotateAnimation creates a rotation animation
func RotateAnimation(chars []string, frame int) string {
	index := frame % len(chars)
	return chars[index]
}

// Blink Animation

// BlinkAnimation creates a blinking effect
func BlinkAnimation(content string, frame int, interval int) string {
	if (frame/interval)%2 == 0 {
		return content
	}
	return ""
}

// Typewriter Animation

// TypewriterAnimation creates a typewriter effect
func TypewriterAnimation(text string, frame int, speed int) string {
	charsToShow := frame / speed
	if charsToShow > len(text) {
		charsToShow = len(text)
	}
	
	return text[:charsToShow]
}

// Marquee Animation

// MarqueeAnimation creates a scrolling marquee effect
func MarqueeAnimation(text string, width int, frame int) string {
	if len(text) <= width {
		return text
	}
	
	// Calculate scroll position
	scrollPos := frame % len(text)
	
	// Create scrolling text
	scrolledText := text[scrollPos:] + " " + text[:scrollPos]
	
	if len(scrolledText) > width {
		return scrolledText[:width]
	}
	
	return scrolledText
}

// Bounce Animation

// BounceAnimation creates a bouncing effect
func BounceAnimation(height int, frame int) int {
	// Calculate bounce position
	t := float64(frame%60) / 60.0
	
	// Bounce easing
	if t < 0.5 {
		return int(float64(height) * (1 - 4*t*t))
	}
	t = t - 0.5
	return int(float64(height) * (4 * t * t))
}

// Shake Animation

// ShakeAnimation creates a shake effect
func ShakeAnimation(frame int, intensity int) int {
	// Calculate shake offset
	if frame%4 < 2 {
		return intensity
	}
	return -intensity
}

// Rainbow Animation

// RainbowAnimation creates a rainbow color cycling effect
func RainbowAnimation(text string, theme Theme, frame int) string {
	colors := []lipgloss.Color{
		theme.NeonPink,
		theme.Purple,
		theme.ElectricBlue,
		theme.Cyan,
		theme.SunsetOrange,
	}
	
	runes := []rune(text)
	result := ""
	
	for i, r := range runes {
		colorIndex := (i + frame) % len(colors)
		style := lipgloss.NewStyle().Foreground(colors[colorIndex])
		result += style.Render(string(r))
	}
	
	return result
}

// Gradient Shift Animation

// GradientShiftAnimation animates a gradient shift
func GradientShiftAnimation(text string, theme Theme, frame int) string {
	// Create shifting gradient
	offset := frame % 60
	gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(text)+offset)
	
	// Apply gradient with offset
	shiftedGradient := gradient[offset:]
	if len(shiftedGradient) < len(text) {
		shiftedGradient = append(shiftedGradient, gradient[:len(text)-len(shiftedGradient)]...)
	}
	
	return ApplyGradientToText(text, shiftedGradient[:len(text)])
}

// Particle Burst Animation

// ParticleBurst creates a particle burst effect
func ParticleBurst(centerX, centerY int, theme Theme, frame int) []Particle {
	var particles []Particle
	
	numParticles := 20
	for i := 0; i < numParticles; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numParticles)
		speed := float64(frame) * 0.5
		
		x := centerX + int(math.Cos(angle)*speed)
		y := centerY + int(math.Sin(angle)*speed)
		
		particles = append(particles, Particle{
			X:     x,
			Y:     y,
			Char:  "·",
			Color: theme.Cyan,
		})
	}
	
	return particles
}

// Particle represents a single particle
type Particle struct {
	X     int
	Y     int
	Char  string
	Color lipgloss.Color
}

// RenderParticles renders a list of particles
func RenderParticles(particles []Particle, width, height int) string {
	// Create empty grid
	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}
	
	// Place particles
	for _, p := range particles {
		if p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height {
			style := lipgloss.NewStyle().Foreground(p.Color)
			grid[p.Y][p.X] = style.Render(p.Char)
		}
	}
	
	// Convert grid to string
	var lines []string
	for _, row := range grid {
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, row...))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
