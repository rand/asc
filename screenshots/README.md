# Screenshots Directory

This directory contains screenshots for the Agent Stack Controller documentation.

## Status

⚠️ **Screenshots are pending capture**

The screenshots need to be captured following the guide in [docs/SCREENSHOTS.md](../docs/SCREENSHOTS.md).

## Required Screenshots

The following screenshots are referenced in the main README and need to be captured:

1. ✅ `tui-dashboard.png` - Main TUI dashboard (placeholder created)
2. ✅ `wizard-welcome.png` - Setup wizard welcome screen (placeholder created)
3. ✅ `wizard-templates.png` - Template selection screen (placeholder created)
4. ✅ `wizard-checks.png` - Dependency check results (placeholder created)
5. ✅ `wizard-api-keys.png` - API key input screen (placeholder created)
6. ✅ `modal-task-details.png` - Task details modal (placeholder created)
7. ✅ `modal-confirmation.png` - Confirmation dialog (placeholder created)
8. ✅ `error-missing-deps.png` - Missing dependencies error (placeholder created)
9. ✅ `error-config.png` - Configuration error (placeholder created)
10. ✅ `agent-states.png` - All agent states (placeholder created)

## How to Capture

Follow the detailed instructions in [docs/SCREENSHOTS.md](../docs/SCREENSHOTS.md) to capture these screenshots.

Quick start:
```bash
# 1. Build the project
make build

# 2. Set terminal size
resize -s 40 120

# 3. Follow capture guide
cat ../docs/SCREENSHOTS.md
```

## Placeholder Images

Placeholder images have been created with instructions. Replace them with actual screenshots by:

1. Following the capture guide
2. Saving screenshots with the exact filenames listed above
3. Optimizing images: `optipng -o7 *.png`
4. Committing the updated screenshots

## Image Specifications

- **Format**: PNG
- **Max file size**: 500KB per image
- **Recommended terminal size**: 120x40
- **Color depth**: 256-color or true-color
- **Optimization**: Use optipng or pngquant

## Notes

- All screenshots should use realistic but safe data
- No real API keys or personal information
- Capture with vaporwave theme (default)
- Include alt text in README references
- Update screenshots when UI changes significantly

