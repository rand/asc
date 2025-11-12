# Screenshot Quick Reference

Quick lookup for what each screenshot shows and where it's used in the README.

## Screenshot Inventory

| File | Description | README Section | Status |
|------|-------------|----------------|--------|
| `tui-dashboard.svg` | Main TUI with 3-pane layout | Usage → TUI Dashboard | Placeholder |
| `wizard-welcome.svg` | Setup wizard welcome screen | Quick Start → Initial Setup | Placeholder |
| `wizard-templates.svg` | Template selection screen | Quick Start → Initial Setup | Placeholder |
| `wizard-checks.svg` | Dependency check results | Quick Start → Initial Setup | Placeholder |
| `wizard-api-keys.svg` | API key input form | Quick Start → Initial Setup | Placeholder |
| `modal-task-details.svg` | Task details modal overlay | Usage → Interactive Features | Placeholder |
| `modal-confirmation.svg` | Confirmation dialog | Usage → Interactive Features | Placeholder |
| `error-missing-deps.svg` | Missing dependencies error | Usage → Health Check | Placeholder |
| `error-config.svg` | Configuration error | Error Handling | Placeholder |
| `agent-states.svg` | All agent status indicators | Agent Status Indicators | Placeholder |

## Screenshot Specifications

### Technical Details
- **Format**: SVG (placeholder) → PNG (final)
- **Terminal size**: 120 columns × 40 rows
- **Color depth**: 256-color or true-color
- **Font**: Monospace with Unicode support
- **Theme**: Vaporwave (default)
- **Max file size**: 500KB per image

### Content Guidelines
- Use realistic but safe data
- No real API keys or personal information
- Show varied agent states when possible
- Include meaningful task titles and log messages
- Demonstrate the vaporwave aesthetic

## Capture Commands

### Quick Capture Session
```bash
# 1. Prepare
make build
resize -s 40 120
clear && printf '\e[3J'

# 2. Capture wizard flow
mv asc.toml asc.toml.bak
asc init
# Capture: welcome, templates, checks, api-keys

# 3. Restore and start TUI
mv asc.toml.bak asc.toml
asc up
# Capture: dashboard, modals (v, k)

# 4. Capture errors
# (Follow error capture steps in CAPTURE_CHECKLIST.md)

# 5. Optimize
optipng -o7 screenshots/*.png
```

## README Integration

### Where Screenshots Appear

1. **Quick Start → Initial Setup** (4 screenshots)
   - Welcome screen
   - Template selection
   - Dependency checks
   - API key input

2. **Usage → TUI Dashboard** (1 screenshot)
   - Main dashboard with all panes

3. **Usage → Interactive Features** (2 screenshots)
   - Task details modal
   - Confirmation dialog

4. **Usage → Health Check** (1 screenshot)
   - Missing dependencies error

5. **Agent Status Indicators** (1 screenshot)
   - All agent states

6. **Error Handling** (1 screenshot)
   - Configuration error

### Alt Text Format

All screenshots include descriptive alt text following this pattern:
```markdown
![Brief description of what's shown and its purpose](path/to/screenshot.svg)
```

Example:
```markdown
![TUI Dashboard showing agent status, task stream, and MCP interaction logs with vaporwave theme](screenshots/tui-dashboard.svg)
```

## Maintenance

### When to Update Screenshots

- Major UI changes (layout, colors, styling)
- New features added to TUI
- Error message format changes
- Wizard flow modifications
- Theme updates

### Update Process

1. Follow capture checklist
2. Optimize new images
3. Replace old files (keep same filenames)
4. Update alt text if description changed
5. Commit with descriptive message

### Version Tracking

Consider adding version or date to screenshot commits:
```bash
git commit -m "docs: Update screenshots for v1.2.0

- Updated TUI dashboard with new agent controls
- Added new modal for agent logs
- Refreshed wizard screens with improved styling
- All screenshots captured 2024-11-11"
```

## Tips

### Getting Good Captures

1. **Timing**: Wait for UI to fully render before capturing
2. **Content**: Use meaningful, realistic data
3. **States**: Show varied agent states (idle, working, error, offline)
4. **Clarity**: Ensure text is sharp and readable
5. **Consistency**: Use same terminal size for all captures

### Common Issues

**Blurry text**: Capture at native resolution, don't scale
**Wrong colors**: Check terminal color support (`echo $TERM`)
**Inconsistent size**: Always use `resize -s 40 120`
**Large files**: Use optipng or pngquant to optimize

### Tools

**macOS**:
- Cmd+Shift+4 (built-in)
- iTerm2 screenshot feature
- Terminal.app export

**Linux**:
- gnome-screenshot
- scrot
- flameshot

**Windows**:
- Windows Terminal export
- Snipping Tool
- PowerShell capture

## Related Documentation

- [SCREENSHOTS.md](../docs/SCREENSHOTS.md) - Detailed capture guide
- [CAPTURE_CHECKLIST.md](CAPTURE_CHECKLIST.md) - Step-by-step checklist
- [README.md](README.md) - Screenshots directory info

