# Screenshot Guide for Agent Stack Controller

This guide explains how to capture and add screenshots for the asc project documentation.

## Required Screenshots

### 1. TUI Main Dashboard
**File**: `screenshots/tui-dashboard.png`
**Description**: The main three-pane TUI interface showing agent status, task stream, and MCP logs
**How to capture**:
```bash
# Start the agent stack
asc up

# Wait for agents to be active and tasks to appear
# Resize terminal to 120x40 for optimal display
# Take screenshot using your terminal's screenshot feature or:
# - macOS: Cmd+Shift+4, then select terminal window
# - Linux: Use gnome-screenshot or scrot
# - Windows Terminal: Use Snipping Tool
```

### 2. Setup Wizard - Welcome Screen
**File**: `screenshots/wizard-welcome.png`
**Description**: The initial welcome screen of the setup wizard
**How to capture**:
```bash
# Remove existing config to trigger wizard
mv asc.toml asc.toml.bak
mv .env .env.bak

# Run init command
asc init

# Capture the welcome screen immediately
```

### 3. Setup Wizard - Template Selection
**File**: `screenshots/wizard-templates.png`
**Description**: Template selection screen showing available agent configurations
**How to capture**:
```bash
# Continue from welcome screen
# Press Enter to advance to template selection
# Capture when template options are displayed
```

### 4. Setup Wizard - Dependency Check
**File**: `screenshots/wizard-checks.png`
**Description**: Dependency check results showing pass/fail status
**How to capture**:
```bash
# Continue through wizard to dependency check step
# Capture when check results are displayed with colored indicators
```

### 5. Setup Wizard - API Key Input
**File**: `screenshots/wizard-api-keys.png`
**Description**: API key input screen with masked input fields
**How to capture**:
```bash
# Continue to API key input step
# Enter sample keys (they will be masked)
# Capture the input form
```

### 6. Modal Dialog - Task Details
**File**: `screenshots/modal-task-details.png`
**Description**: Modal showing detailed task information
**How to capture**:
```bash
# In running TUI, navigate to a task
# Press 'v' to view task details
# Capture the modal overlay
```

### 7. Modal Dialog - Confirmation
**File**: `screenshots/modal-confirmation.png`
**Description**: Confirmation dialog for destructive actions
**How to capture**:
```bash
# In running TUI, select an agent
# Press 'k' to kill agent (triggers confirmation)
# Capture the confirmation modal
```

### 8. Error State - Missing Dependencies
**File**: `screenshots/error-missing-deps.png`
**Description**: Error display when dependencies are missing
**How to capture**:
```bash
# Temporarily rename a required binary
sudo mv /usr/local/bin/bd /usr/local/bin/bd.bak

# Run check command
asc check

# Capture the error output
# Restore binary
sudo mv /usr/local/bin/bd.bak /usr/local/bin/bd
```

### 9. Error State - Configuration Error
**File**: `screenshots/error-config.png`
**Description**: Error display for invalid configuration
**How to capture**:
```bash
# Create invalid asc.toml
echo "invalid toml [[[" > asc.toml

# Try to start
asc up

# Capture the error message
# Restore valid config
```

### 10. Agent Status - All States
**File**: `screenshots/agent-states.png`
**Description**: Agent pane showing all possible states (idle, working, error, offline)
**How to capture**:
```bash
# Configure multiple agents in different states
# Start stack and wait for varied states
# Capture when agents show different status indicators
```

### 11. Vaporwave Theme
**File**: `screenshots/theme-vaporwave.png`
**Description**: TUI with vaporwave theme showing neon colors and effects
**How to capture**:
```bash
# Ensure vaporwave theme is active (default)
asc up
# Capture the colorful, neon-styled interface
```

### 12. Services Command Output
**File**: `screenshots/services-status.png`
**Description**: Output of services status command
**How to capture**:
```bash
asc services start
asc services status
# Capture the status output
```

## Screenshot Requirements

### Technical Specifications
- **Format**: PNG (preferred) or JPEG
- **Resolution**: At least 1200px wide for clarity
- **Terminal size**: 120 columns x 40 rows (recommended)
- **Color depth**: 256 colors or true color
- **File size**: Optimize to < 500KB per image

### Quality Guidelines
1. **Clean terminal**: Clear scrollback before capturing
2. **Realistic data**: Use meaningful agent names, task titles, and log messages
3. **Timing**: Capture when UI is fully rendered and stable
4. **Lighting**: Use dark terminal background for vaporwave theme
5. **Cropping**: Include terminal window chrome for context, or crop cleanly to content

### Recommended Terminal Settings
```bash
# Terminal size
resize -s 40 120

# Font: Use a monospace font with good Unicode support
# Recommended: JetBrains Mono, Fira Code, or SF Mono

# Color scheme: Ensure 256-color or true-color support
echo $TERM  # Should show xterm-256color or similar
```

## Screenshot Tools

### macOS
- **Built-in**: Cmd+Shift+4, then click terminal window
- **iTerm2**: Edit → Copy Mode → Select → Right-click → Copy
- **Terminal.app**: Shell → Export as PDF (then convert to PNG)

### Linux
- **gnome-screenshot**: `gnome-screenshot -w` (window mode)
- **scrot**: `scrot -u` (focused window)
- **flameshot**: `flameshot gui`
- **ImageMagick**: `import screenshot.png` (then click window)

### Windows
- **Windows Terminal**: Right-click → Export → PNG
- **Snipping Tool**: Win+Shift+S
- **PowerShell**: Use Windows.Graphics.Capture API

## Image Optimization

After capturing, optimize images:

```bash
# Install optimization tools
brew install optipng pngquant  # macOS
sudo apt install optipng pngquant  # Linux

# Optimize PNG files
optipng -o7 screenshots/*.png
pngquant --quality=80-95 screenshots/*.png

# Or use online tools:
# - TinyPNG: https://tinypng.com/
# - Squoosh: https://squoosh.app/
```

## Adding Screenshots to README

1. Create `screenshots/` directory in project root
2. Add screenshots with descriptive names
3. Update `.gitignore` if needed (screenshots should be committed)
4. Reference in README using relative paths:
   ```markdown
   ![TUI Dashboard](screenshots/tui-dashboard.png)
   ```
5. Always include alt text for accessibility

## Alt Text Guidelines

Good alt text describes the content and purpose:

**Good**: "TUI dashboard showing three agents in different states, with task stream displaying 5 open tasks and MCP log showing recent lease operations"

**Bad**: "Screenshot of asc"

## Maintenance

- Update screenshots when UI changes significantly
- Keep screenshots in sync with current version
- Date screenshots in commit messages
- Review screenshots quarterly for accuracy

## Example Capture Session

```bash
# 1. Prepare environment
cd ~/asc-project
make build
resize -s 40 120

# 2. Capture wizard flow
mv asc.toml asc.toml.bak
asc init
# Capture each step, pressing Enter to advance

# 3. Capture running TUI
asc up
# Wait for agents to start
# Capture main dashboard
# Press 'v' on a task, capture modal
# Press 'k' on agent, capture confirmation

# 4. Capture error states
# (Follow error capture steps above)

# 5. Optimize images
cd screenshots
optipng -o7 *.png

# 6. Commit
git add screenshots/
git commit -m "docs: Add TUI and wizard screenshots"
```

## Notes

- Screenshots should represent the **current** state of the application
- Use **realistic** but **safe** data (no real API keys, personal info)
- Capture in **high resolution** but optimize file size
- Include **diverse scenarios** (success, error, loading states)
- Test screenshots on **different displays** (retina, standard)
- Consider **dark mode** users (vaporwave theme is dark by default)

