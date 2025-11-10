# Vaporwave Aesthetic Design System

## Overview

The Agent Stack Controller (asc) features a comprehensive vaporwave aesthetic design system that provides an elegant, retro-futuristic visual experience. This document describes the design system components and how to use them.

## Design Philosophy

The vaporwave aesthetic is characterized by:
- **Neon colors**: Bright, luminous colors that pop against dark backgrounds
- **Gradients**: Smooth color transitions for depth and visual interest
- **Geometric patterns**: Clean lines, grids, and shapes
- **Glow effects**: Luminous, pulsing elements that draw attention
- **Retro-futuristic**: Combining 80s/90s aesthetics with modern design
- **Elegant spacing**: Golden ratio proportions and consistent grid system

## Color Palette

### Core Colors
- **Neon Pink** (#FF71CE): Primary accent, used for highlights and important elements
- **Electric Blue** (#01CDFE): Secondary accent, used for interactive elements
- **Purple** (#B967FF): Tertiary accent, used for borders and decorative elements
- **Cyan** (#05FFA1): Success states and positive indicators
- **Sunset Orange** (#FFFB96): Warnings and attention-grabbing elements

### Base Colors
- **Deep Purple** (#1A0933): Primary background shade
- **Midnight Blue** (#0D0221): Main background color
- **Dark Teal** (#0F0E17): Alternative background shade

### Glow Colors
Brighter versions of core colors used for luminous effects:
- Glow Pink (#FF9EE5)
- Glow Blue (#4DE4FF)
- Glow Purple (#D49FFF)
- Glow Cyan (#5FFFC4)

### Status Colors
- **Success**: Cyan (#05FFA1) - Completed tasks, healthy agents
- **Warning**: Sunset Orange (#FFFB96) - Caution states
- **Error**: Neon Pink (#FF71CE) - Errors and critical states
- **Info**: Electric Blue (#01CDFE) - Informational messages

## Typography

### Text Hierarchy
1. **H1**: Large headers with gradient fills and letter spacing
2. **H2**: Medium headers with bold styling
3. **Body**: Standard text for content
4. **Caption**: Muted, smaller text for secondary information
5. **Code**: Monospace text with neon accents for IDs and technical content

### Text Effects
- **Gradient text**: Character-by-character color gradients
- **Holographic shimmer**: Rainbow color cycling effect
- **Neon glow**: Bold text with bright colors
- **Letter spacing**: Elegant spacing for headers
- **Text shadows**: Depth through simulated shadows

## Borders and Frames

### Border Styles
- **Rounded**: Smooth, friendly corners (default)
- **Double**: Elegant double-line borders
- **Thick**: Bold, prominent borders
- **Neon**: Decorative block-style borders

### Border Effects
- **Glow**: Pulsing luminous borders
- **Gradient**: Color-shifting borders
- **Animated**: Color cycling and pulsing
- **Layered**: Multiple borders for depth
- **Decorative**: Geometric ornaments at corners

## Status Indicators

### Agent Status
- **Idle**: Green filled circle (●) - Agent ready for work
- **Working**: Blue rotating icon (⟳) - Agent processing task
- **Error**: Red exclamation (!) - Agent encountered error
- **Offline**: Gray empty circle (○) - Agent not responding

### Progress Indicators
- **Progress bars**: Gradient-filled bars with shine effects
- **Loading spinners**: Animated rotating indicators
- **Pulsing dots**: Breathing animation for activity
- **Signal waves**: Animated connection indicators

### Task Status
- **Open**: Gray empty circle (○)
- **In Progress**: Yellow filled circle with dot (◉)
- **Completed**: Green checkmark (✓)
- **Blocked**: Red cross (✗)

## Patterns and Backgrounds

### Grid Patterns
- **Subtle grid**: Light grid overlay for depth
- **Perspective grid**: Vanishing point effect
- **Hexagon pattern**: Geometric hexagonal grid
- **Triangle pattern**: Alternating triangle shapes

### Effects
- **Scanlines**: Retro CRT monitor effect
- **Noise texture**: Subtle grain for depth
- **Floating shapes**: Animated geometric shapes
- **Particles**: Twinkling particle effects
- **Starfield**: Animated star background
- **Matrix rain**: Falling character effect

## Animations

### Easing Functions
- **Linear**: No easing
- **Ease In/Out Quad**: Smooth acceleration/deceleration
- **Ease In/Out Cubic**: More pronounced curves
- **Ease In/Out Sine**: Sinusoidal smoothness
- **Ease In/Out Expo**: Exponential curves
- **Elastic**: Bouncy, spring-like motion

### Animation Types
- **Fade**: Smooth opacity transitions
- **Slide**: Content sliding in/out
- **Pulse**: Breathing/pulsing effect
- **Shimmer**: Moving highlight effect
- **Wave**: Undulating motion
- **Glow**: Intensity pulsing
- **Rainbow**: Color cycling
- **Ripple**: Expanding circle effect

## Layout System

### Grid System
- **Base unit**: 1 character (8px equivalent)
- **Spacing levels**: Multiples of base unit (1, 2, 3, etc.)
- **Golden ratio**: 1.618 for proportions
- **Consistent padding**: Grid-based padding and margins

### Responsive Breakpoints
- **Small** (0-80 chars): Single column layout
- **Medium** (81-120 chars): Two column layout
- **Large** (121-160 chars): Three pane layout
- **XLarge** (161+ chars): Full three pane layout

### Layout Patterns
- **Split layout**: Divide space with golden ratio
- **Grid layout**: Multi-column grid
- **Stack layout**: Vertical stacking with spacing
- **Centered layout**: Center content in available space
- **Masonry layout**: Pinterest-style layout

## Modals and Overlays

### Modal Styles
- **Glass morphism**: Frosted glass effect with transparency
- **Neon**: High-contrast with glowing borders
- **Minimal**: Clean, simple design

### Modal Components
- **Title bar**: Gradient text with ornaments
- **Content area**: Padded content with proper spacing
- **Buttons**: Primary and secondary button styles
- **Close button**: Styled close icon
- **Backdrop**: Semi-transparent overlay

### Modal Animations
- **Fade in/out**: Smooth appearance/disappearance
- **Floating**: Subtle vertical movement
- **Pulsing**: Breathing border effect
- **Shimmering**: Animated title effect

## Theme System

### Built-in Themes
1. **Vaporwave** (default): Full vaporwave aesthetic
2. **Cyberpunk**: High-contrast neon colors
3. **Minimal**: Clean, understated design
4. **Accessibility**: High-contrast for readability

### Custom Themes
Themes can be customized via JSON configuration files:
```json
{
  "name": "my-theme",
  "colors": {
    "neon_pink": "#FF71CE",
    "electric_blue": "#01CDFE",
    "purple": "#B967FF",
    ...
  }
}
```

### Theme Management
- Load themes: `themeManager.LoadTheme("vaporwave")`
- Save themes: `themeManager.SaveTheme("my-theme")`
- Export themes: `themeManager.ExportTheme("path/to/theme.json")`
- Import themes: `themeManager.ImportTheme("path/to/theme.json")`

## Performance Optimization

### Rendering Optimization
- **Frame rate targeting**: 60 FPS target
- **Render caching**: Cache unchanged content
- **Dirty tracking**: Only re-render changed panes
- **Batch updates**: Group multiple updates together
- **Throttling**: Limit update frequency
- **Debouncing**: Delay rapid updates

### Best Practices
1. Use cached renders when content hasn't changed
2. Mark panes as dirty only when data changes
3. Batch multiple updates together
4. Throttle expensive operations
5. Use appropriate animation frame rates

## Terminal Compatibility

### Color Support
- **True color** (24-bit): Full color support
- **256 color**: Good color approximation
- **16 color (ANSI)**: Fallback to basic colors
- **Graceful degradation**: Adapts to terminal capabilities

### Tested Terminals
- ✅ iTerm2 (macOS)
- ✅ Alacritty (cross-platform)
- ✅ Windows Terminal
- ✅ Kitty
- ✅ Hyper
- ⚠️ Terminal.app (limited color support)
- ⚠️ Basic terminals (fallback mode)

## Usage Examples

### Creating a Vaporwave Pane
```go
theme := VaporwaveTheme()
pane := CreateGlowPaneStyle(width, height, "Agent Status", theme)
content := pane.Render(agentContent)
```

### Rendering with Gradient Text
```go
title := "Agent Stack Controller"
gradient := GenerateGradient(theme.GradientPrimary[0], theme.GradientPrimary[1], len(title))
gradientTitle := ApplyGradientToText(title, gradient)
```

### Creating Animated Indicators
```go
phase := float64(frame) / 10.0
indicator := RenderGlowingOrb(IndicatorWorking, theme, phase)
```

### Using Responsive Layout
```go
layout := GetResponsiveLayout(width, height)
dimensions := CalculatePaneDimensions(width, height, layout)
```

## Micro-Interactions

Subtle animations that enhance user experience:
- **Button hover**: Color shift on hover
- **Selection highlight**: Background glow
- **Status transitions**: Smooth color morphing
- **Loading states**: Pulsing indicators
- **Success feedback**: Brief flash animation
- **Error shake**: Subtle shake effect

## Accessibility Considerations

### High Contrast Mode
The accessibility theme provides:
- Pure white text on black background
- High contrast status colors
- No gradients or subtle effects
- Clear, bold indicators

### Best Practices
1. Provide text alternatives for icons
2. Use color + shape for status (not just color)
3. Ensure sufficient contrast ratios
4. Support keyboard navigation
5. Provide clear focus indicators

## Design Guidelines

### Do's
✅ Use consistent spacing (grid system)
✅ Apply golden ratio for proportions
✅ Use gradients for visual interest
✅ Add subtle animations for delight
✅ Maintain visual hierarchy
✅ Provide breathing room (whitespace)

### Don'ts
❌ Overuse animations (can be distracting)
❌ Use too many colors at once
❌ Ignore terminal limitations
❌ Sacrifice readability for aesthetics
❌ Use inconsistent spacing
❌ Clutter the interface

## Future Enhancements

Planned improvements:
- [ ] More built-in themes
- [ ] Theme editor UI
- [ ] Custom animation curves
- [ ] More geometric patterns
- [ ] Advanced particle systems
- [ ] 3D perspective effects
- [ ] Sound effects integration
- [ ] Custom font support

## Resources

- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Bubbletea Framework](https://github.com/charmbracelet/bubbletea)
- [Vaporwave Aesthetics](https://en.wikipedia.org/wiki/Vaporwave)
- [Golden Ratio in Design](https://www.canva.com/learn/what-is-the-golden-ratio/)

## Contributing

To contribute to the design system:
1. Follow existing patterns and conventions
2. Test on multiple terminal emulators
3. Ensure graceful degradation
4. Document new components
5. Provide usage examples
6. Consider performance impact

## License

The vaporwave design system is part of the Agent Stack Controller project and follows the same license.
