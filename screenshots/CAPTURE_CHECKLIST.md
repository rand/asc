# Screenshot Capture Checklist

Use this checklist when capturing actual screenshots to replace the SVG placeholders.

## Pre-Capture Setup

- [ ] Build the latest version: `make build`
- [ ] Set terminal size: `resize -s 40 120`
- [ ] Use a terminal with 256-color or true-color support
- [ ] Clear terminal scrollback: `clear && printf '\e[3J'`
- [ ] Ensure good lighting/contrast for captures
- [ ] Use a monospace font with good Unicode support (JetBrains Mono, Fira Code, SF Mono)

## Screenshot Capture Tasks

### 1. TUI Dashboard (tui-dashboard.png)
- [ ] Start the agent stack: `asc up`
- [ ] Wait for agents to be in varied states (idle, working, offline)
- [ ] Ensure task stream shows multiple tasks
- [ ] Ensure MCP log shows recent activity
- [ ] Capture the full TUI interface
- [ ] Replace `screenshots/tui-dashboard.svg` with `screenshots/tui-dashboard.png`

### 2. Setup Wizard - Welcome (wizard-welcome.png)
- [ ] Backup existing config: `mv asc.toml asc.toml.bak && mv .env .env.bak`
- [ ] Run: `asc init`
- [ ] Capture the welcome screen immediately
- [ ] Replace `screenshots/wizard-welcome.svg` with `screenshots/wizard-welcome.png`

### 3. Setup Wizard - Templates (wizard-templates.png)
- [ ] Continue from welcome screen (press Enter)
- [ ] Capture when template selection is displayed
- [ ] Show cursor on one of the options
- [ ] Replace `screenshots/wizard-templates.svg` with `screenshots/wizard-templates.png`

### 4. Setup Wizard - Checks (wizard-checks.png)
- [ ] Continue through wizard to dependency check step
- [ ] Capture when check results are displayed
- [ ] Ensure mix of pass/fail/warn indicators if possible
- [ ] Replace `screenshots/wizard-checks.svg` with `screenshots/wizard-checks.png`

### 5. Setup Wizard - API Keys (wizard-api-keys.png)
- [ ] Continue to API key input step
- [ ] Enter sample keys (they will be masked with bullets)
- [ ] Capture the input form with at least one field filled
- [ ] Replace `screenshots/wizard-api-keys.svg` with `screenshots/wizard-api-keys.png`

### 6. Modal - Task Details (modal-task-details.png)
- [ ] In running TUI, navigate to a task using arrow keys
- [ ] Press 'v' to view task details
- [ ] Capture the modal overlay with task information
- [ ] Replace `screenshots/modal-task-details.svg` with `screenshots/modal-task-details.png`

### 7. Modal - Confirmation (modal-confirmation.png)
- [ ] In running TUI, select an agent
- [ ] Press 'k' to kill agent (triggers confirmation)
- [ ] Capture the confirmation modal
- [ ] Press 'n' to cancel (don't actually kill the agent)
- [ ] Replace `screenshots/modal-confirmation.svg` with `screenshots/modal-confirmation.png`

### 8. Error - Missing Dependencies (error-missing-deps.png)
- [ ] Temporarily rename a required binary: `sudo mv /usr/local/bin/bd /usr/local/bin/bd.bak`
- [ ] Run: `asc check`
- [ ] Capture the error output showing missing dependency
- [ ] Restore binary: `sudo mv /usr/local/bin/bd.bak /usr/local/bin/bd`
- [ ] Replace `screenshots/error-missing-deps.svg` with `screenshots/error-missing-deps.png`

### 9. Error - Configuration (error-config.png)
- [ ] Backup valid config: `cp asc.toml asc.toml.valid`
- [ ] Create invalid config: `echo "invalid toml [[[" > asc.toml`
- [ ] Try to start: `asc up`
- [ ] Capture the error message
- [ ] Restore config: `mv asc.toml.valid asc.toml`
- [ ] Replace `screenshots/error-config.svg` with `screenshots/error-config.png`

### 10. Agent States (agent-states.png)
- [ ] Configure multiple agents in asc.toml
- [ ] Start stack: `asc up`
- [ ] Wait for agents to show different states
- [ ] Capture when agents show: idle, working, error, offline
- [ ] May need to simulate states (stop an agent, trigger an error)
- [ ] Replace `screenshots/agent-states.svg` with `screenshots/agent-states.png`

## Post-Capture Processing

- [ ] Optimize all PNG files: `optipng -o7 screenshots/*.png`
- [ ] Or use pngquant: `pngquant --quality=80-95 screenshots/*.png`
- [ ] Verify file sizes are < 500KB each
- [ ] Check images display correctly in README
- [ ] Verify alt text is descriptive and accurate

## Commit Changes

```bash
# Remove SVG placeholders
rm screenshots/*.svg

# Add PNG screenshots
git add screenshots/*.png

# Commit
git commit -m "docs: Add actual TUI and wizard screenshots

- Replace SVG placeholders with real terminal captures
- Optimize images for web display
- All screenshots captured at 120x40 terminal size
- Vaporwave theme with 256-color support"
```

## Quality Checklist

- [ ] All screenshots use realistic but safe data (no real API keys)
- [ ] Terminal size is consistent (120x40)
- [ ] Vaporwave theme is clearly visible
- [ ] Text is readable and not blurry
- [ ] Colors are accurate and vibrant
- [ ] No personal information visible
- [ ] File sizes are optimized
- [ ] Alt text in README is descriptive

## Notes

- If you can't capture all states in one session, that's okay
- Focus on the most important screenshots first (TUI dashboard, wizard welcome)
- You can capture additional screenshots later as needed
- Consider creating a video walkthrough as well for the documentation

## Troubleshooting

**Terminal too small**: Use `resize -s 40 120` or adjust terminal window manually

**Colors not showing**: Ensure `$TERM` is set to `xterm-256color` or similar

**Can't capture modal**: Make sure you're pressing the right key ('v' for task details, 'k' for confirmation)

**Agent states not varied**: Manually stop/start agents or wait for natural state changes

