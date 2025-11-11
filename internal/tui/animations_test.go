package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestNewAnimation(t *testing.T) {
	duration := 1 * time.Second
	anim := NewAnimation(duration, EaseLinear)

	if anim == nil {
		t.Fatal("NewAnimation returned nil")
	}

	if anim.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, anim.Duration)
	}

	if anim.Easing == nil {
		t.Error("Easing function not set")
	}
}

func TestAnimationProgress(t *testing.T) {
	duration := 100 * time.Millisecond
	anim := NewAnimation(duration, EaseLinear)

	// Immediately after creation, progress should be near 0
	progress := anim.Progress()
	if progress < 0 || progress > 0.1 {
		t.Errorf("Initial progress should be near 0, got %f", progress)
	}

	// Wait for animation to complete
	time.Sleep(duration + 10*time.Millisecond)

	// Progress should be 1.0
	progress = anim.Progress()
	if progress != 1.0 {
		t.Errorf("Expected progress 1.0 after duration, got %f", progress)
	}
}

func TestAnimationIsComplete(t *testing.T) {
	duration := 50 * time.Millisecond
	anim := NewAnimation(duration, EaseLinear)

	// Should not be complete immediately
	if anim.IsComplete() {
		t.Error("Animation should not be complete immediately")
	}

	// Wait for completion
	time.Sleep(duration + 10*time.Millisecond)

	// Should be complete
	if !anim.IsComplete() {
		t.Error("Animation should be complete after duration")
	}
}

func TestAnimationReset(t *testing.T) {
	duration := 50 * time.Millisecond
	anim := NewAnimation(duration, EaseLinear)

	// Wait a bit
	time.Sleep(30 * time.Millisecond)

	// Reset
	anim.Reset()

	// Progress should be near 0 again
	progress := anim.Progress()
	if progress > 0.1 {
		t.Errorf("Progress after reset should be near 0, got %f", progress)
	}
}

func TestEasingFunctions(t *testing.T) {
	tests := []struct {
		name   string
		easing EasingFunction
	}{
		{"Linear", EaseLinear},
		{"InQuad", EaseInQuad},
		{"OutQuad", EaseOutQuad},
		{"InOutQuad", EaseInOutQuad},
		{"InCubic", EaseInCubic},
		{"OutCubic", EaseOutCubic},
		{"InOutCubic", EaseInOutCubic},
		{"InSine", EaseInSine},
		{"OutSine", EaseOutSine},
		{"InOutSine", EaseInOutSine},
		{"InExpo", EaseInExpo},
		{"OutExpo", EaseOutExpo},
		{"InOutExpo", EaseInOutExpo},
		{"InElastic", EaseInElastic},
		{"OutElastic", EaseOutElastic},
		{"InOutElastic", EaseInOutElastic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test at key points
			testPoints := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

			for _, point := range testPoints {
				result := tt.easing(point)

				// Result should be a valid number
				if result < -1 || result > 2 {
					t.Errorf("%s(%f) = %f, expected value in reasonable range", tt.name, point, result)
				}

				// At t=0, most easing functions should return 0
				if point == 0.0 && result > 0.1 {
					t.Errorf("%s(0) = %f, expected near 0", tt.name, result)
				}

				// At t=1, most easing functions should return 1
				if point == 1.0 && (result < 0.9 || result > 1.1) {
					t.Errorf("%s(1) = %f, expected near 1", tt.name, result)
				}
			}
		})
	}
}

func TestEaseLinear(t *testing.T) {
	// Linear should return the input value
	tests := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, input := range tests {
		result := EaseLinear(input)
		if result != input {
			t.Errorf("EaseLinear(%f) = %f, want %f", input, result, input)
		}
	}
}

func TestAnimateColorTransition(t *testing.T) {
	from := lipgloss.Color("#FF0000")
	to := lipgloss.Color("#0000FF")

	tests := []struct {
		name     string
		progress float64
	}{
		{"start", 0.0},
		{"middle", 0.5},
		{"end", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnimateColorTransition(from, to, tt.progress)
			if result == "" {
				t.Error("AnimateColorTransition returned empty color")
			}
		})
	}
}

func TestAnimateFade(t *testing.T) {
	theme := VaporwaveTheme()
	content := "Test content"

	tests := []struct {
		name     string
		progress float64
		fadeIn   bool
	}{
		{"fade_in_start", 0.0, true},
		{"fade_in_end", 1.0, true},
		{"fade_out_start", 0.0, false},
		{"fade_out_end", 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnimateFade(content, theme, tt.progress, tt.fadeIn)

			if tt.fadeIn && tt.progress >= 1.0 {
				if result != content {
					t.Error("Fade in at 100% should return full content")
				}
			}

			if !tt.fadeIn && tt.progress >= 1.0 {
				if result != "" {
					t.Error("Fade out at 100% should return empty string")
				}
			}
		})
	}
}

func TestNewAnimationFrame(t *testing.T) {
	maxFrame := 60
	af := NewAnimationFrame(maxFrame, true)

	if af == nil {
		t.Fatal("NewAnimationFrame returned nil")
	}

	if af.Frame != 0 {
		t.Errorf("Initial frame should be 0, got %d", af.Frame)
	}

	if af.MaxFrame != maxFrame {
		t.Errorf("MaxFrame should be %d, got %d", maxFrame, af.MaxFrame)
	}
}

func TestAnimationFrameNext(t *testing.T) {
	t.Run("looping", func(t *testing.T) {
		af := NewAnimationFrame(5, true)

		// Advance through frames
		for i := 0; i < 10; i++ {
			af.Next()
		}

		// Should have looped
		if af.Frame >= af.MaxFrame {
			t.Errorf("Frame should have looped, got %d", af.Frame)
		}
	})

	t.Run("non_looping", func(t *testing.T) {
		af := NewAnimationFrame(5, false)

		// Advance past max
		for i := 0; i < 10; i++ {
			af.Next()
		}

		// Should be clamped at max-1
		if af.Frame != af.MaxFrame-1 {
			t.Errorf("Frame should be clamped at %d, got %d", af.MaxFrame-1, af.Frame)
		}
	})
}

func TestAnimationFrameProgress(t *testing.T) {
	af := NewAnimationFrame(10, false)

	// At start
	if af.Progress() != 0.0 {
		t.Errorf("Initial progress should be 0, got %f", af.Progress())
	}

	// Advance halfway
	for i := 0; i < 5; i++ {
		af.Next()
	}

	progress := af.Progress()
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("Progress at frame 5/10 should be around 0.5, got %f", progress)
	}
}

func TestAnimationFrameIsComplete(t *testing.T) {
	t.Run("looping_never_complete", func(t *testing.T) {
		af := NewAnimationFrame(5, true)

		for i := 0; i < 10; i++ {
			af.Next()
			if af.IsComplete() {
				t.Error("Looping animation should never be complete")
			}
		}
	})

	t.Run("non_looping_completes", func(t *testing.T) {
		af := NewAnimationFrame(5, false)

		// Advance to end
		for i := 0; i < 10; i++ {
			af.Next()
		}

		if !af.IsComplete() {
			t.Error("Non-looping animation should be complete")
		}
	})
}

func TestAnimationFrameReset(t *testing.T) {
	af := NewAnimationFrame(10, false)

	// Advance
	for i := 0; i < 5; i++ {
		af.Next()
	}

	// Reset
	af.Reset()

	if af.Frame != 0 {
		t.Errorf("Frame after reset should be 0, got %d", af.Frame)
	}
}

func TestWaveEffect(t *testing.T) {
	theme := VaporwaveTheme()
	text := "Hello World"

	result := WaveEffect(text, theme, 0)
	if result == "" {
		t.Error("WaveEffect returned empty string")
	}

	// Test with different frames
	result2 := WaveEffect(text, theme, 10)
	if result2 == "" {
		t.Error("WaveEffect returned empty string for frame 10")
	}
}

func TestRotateAnimation(t *testing.T) {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	tests := []struct {
		frame int
		want  int
	}{
		{0, 0},
		{5, 5},
		{10, 0}, // Should wrap around
		{15, 5},
	}

	for _, tt := range tests {
		result := RotateAnimation(chars, tt.frame)
		expected := chars[tt.want]
		if result != expected {
			t.Errorf("RotateAnimation(frame=%d) = %s, want %s", tt.frame, result, expected)
		}
	}
}

func TestBlinkAnimation(t *testing.T) {
	content := "Blink"
	interval := 5

	// Even intervals should show content
	result := BlinkAnimation(content, 0, interval)
	if result != content {
		t.Error("BlinkAnimation should show content at frame 0")
	}

	// Odd intervals should hide content
	result = BlinkAnimation(content, interval, interval)
	if result != "" {
		t.Error("BlinkAnimation should hide content at odd interval")
	}
}

func TestTypewriterAnimation(t *testing.T) {
	text := "Hello"
	speed := 2

	tests := []struct {
		frame int
		want  int // expected length
	}{
		{0, 0},
		{2, 1},
		{4, 2},
		{10, 5}, // Full text
		{20, 5}, // Should not exceed text length
	}

	for _, tt := range tests {
		result := TypewriterAnimation(text, tt.frame, speed)
		if len(result) != tt.want {
			t.Errorf("TypewriterAnimation(frame=%d) length = %d, want %d", tt.frame, len(result), tt.want)
		}
	}
}

func TestMarqueeAnimation(t *testing.T) {
	text := "Hello World"
	width := 5

	// Test that it returns something
	result := MarqueeAnimation(text, width, 0)
	if len(result) > width {
		t.Errorf("MarqueeAnimation should not exceed width %d, got %d", width, len(result))
	}

	// Test with short text
	shortText := "Hi"
	result = MarqueeAnimation(shortText, width, 0)
	if result != shortText {
		t.Error("MarqueeAnimation should return full text if shorter than width")
	}
}

func TestBounceAnimation(t *testing.T) {
	height := 10

	// Test various frames
	for frame := 0; frame < 60; frame++ {
		result := BounceAnimation(height, frame)
		if result < 0 || result > height {
			t.Errorf("BounceAnimation(frame=%d) = %d, should be in range [0, %d]", frame, result, height)
		}
	}
}

func TestShakeAnimation(t *testing.T) {
	intensity := 5

	// Test that it returns values within expected range
	for frame := 0; frame < 10; frame++ {
		result := ShakeAnimation(frame, intensity)
		if result != intensity && result != -intensity {
			t.Errorf("ShakeAnimation should return ±%d, got %d", intensity, result)
		}
	}
}

func TestRainbowAnimation(t *testing.T) {
	theme := VaporwaveTheme()
	text := "Rainbow"

	result := RainbowAnimation(text, theme, 0)
	if result == "" {
		t.Error("RainbowAnimation returned empty string")
	}

	// Test with different frame
	result2 := RainbowAnimation(text, theme, 5)
	if result2 == "" {
		t.Error("RainbowAnimation returned empty string for frame 5")
	}
}

func TestGradientShiftAnimation(t *testing.T) {
	theme := VaporwaveTheme()
	text := "Gradient"

	result := GradientShiftAnimation(text, theme, 0)
	if result == "" {
		t.Error("GradientShiftAnimation returned empty string")
	}
}

func TestParticleBurst(t *testing.T) {
	theme := VaporwaveTheme()
	centerX, centerY := 10, 10

	particles := ParticleBurst(centerX, centerY, theme, 5)

	if len(particles) == 0 {
		t.Error("ParticleBurst should create particles")
	}

	// Check that particles have valid properties
	for _, p := range particles {
		if p.Char == "" {
			t.Error("Particle should have a character")
		}
		if p.Color == "" {
			t.Error("Particle should have a color")
		}
	}
}

func TestRenderParticles(t *testing.T) {
	theme := VaporwaveTheme()
	width, height := 20, 10

	particles := []Particle{
		{X: 5, Y: 5, Char: "·", Color: theme.Cyan},
		{X: 10, Y: 5, Char: "·", Color: theme.NeonPink},
	}

	result := RenderParticles(particles, width, height)
	if result == "" {
		t.Error("RenderParticles returned empty string")
	}
}

func TestGlowAnimation(t *testing.T) {
	theme := VaporwaveTheme()
	text := "Glow"

	result := GlowAnimation(text, theme, 0)
	if result == "" {
		t.Error("GlowAnimation returned empty string")
	}

	// Test with different frame
	result2 := GlowAnimation(text, theme, 10)
	if result2 == "" {
		t.Error("GlowAnimation returned empty string for frame 10")
	}
}

func TestShimmerAnimation(t *testing.T) {
	theme := VaporwaveTheme()
	text := "Shimmer"

	result := ShimmerAnimation(text, theme, 0)
	if result == "" {
		t.Error("ShimmerAnimation returned empty string")
	}
}
