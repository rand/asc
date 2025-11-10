# Interactive TUI Features

This document describes the interactive features added to the Agent Stack Controller TUI in task 23.

## Overview

The TUI now supports rich interactive features for managing tasks, controlling agents, and filtering logs. All interactions are keyboard-driven for efficient workflow.

## Task Interaction Features (23.1)

### Navigation
- **Arrow Keys (↑/↓)**: Navigate through the task list
- Selected task is highlighted with a `▶` indicator and background color

### Task Actions
- **c**: Claim the selected task for the current user
- **v**: View full task details in a modal dialog
- **n**: Create a new task with an interactive input form

### Task Detail Modal
When viewing a task (press 'v'), a modal displays:
- Task ID
- Title
- Status
- Phase
- Assignee (if assigned)

Press 'v' or 'esc' to close the modal.

### Create Task Modal
When creating a task (press 'n'), an input form appears:
- Type the task title
- Press 'enter' to create
- Press 'esc' to cancel

## Agent Control Features (23.2)

### Agent Selection
- **1-9**: Select an agent by number (shown in the agent pane)
- Selected agent is highlighted with a `▶` indicator

### Agent Actions
- **p**: Pause/resume the selected agent (not yet fully implemented)
- **k**: Kill the selected agent (shows confirmation dialog)
- **R** (Shift+R): Restart the selected agent (shows confirmation dialog)
- **l**: View the log file path for the selected agent

### Confirmation Dialogs
Destructive actions (kill, restart) show a confirmation modal:
- Press 'y' to confirm
- Press 'n' or 'esc' to cancel

## Log Filtering and Search Features (23.3)

### Search Mode
- **/**: Enter search mode
- Type search text to filter messages by content or source
- Press 'enter' to apply the filter
- Press 'esc' to cancel

### Filter Controls
- **a**: Cycle through agent name filters (filters logs by specific agent)
- **m**: Cycle through message type filters (lease, beads, error, message)
- **x**: Clear all active filters

### Log Export
- **e**: Export filtered logs to a timestamped file (asc-logs-YYYYMMDD-HHMMSS.txt)

### Active Filters Display
The log pane header shows active filters:
- `[search:term]` - Active search filter
- `[agent:name]` - Active agent filter
- `[type:message_type]` - Active message type filter

## Keybinding Reference

### Global Keys
- **q** or **Ctrl+C**: Quit and shutdown agents
- **r**: Force refresh all data
- **t**: Run stack health test

### Task Pane Keys
- **↑/↓**: Navigate task list
- **c**: Claim selected task
- **v**: View task details
- **n**: Create new task

### Agent Pane Keys
- **1-9**: Select agent
- **p**: Pause/resume agent
- **k**: Kill agent (with confirmation)
- **R**: Restart agent (with confirmation)
- **l**: View agent logs

### Log Pane Keys
- **/**: Enter search mode
- **a**: Cycle agent filter
- **m**: Cycle message type filter
- **x**: Clear all filters
- **e**: Export logs

## Implementation Details

### State Management
The TUI model tracks:
- `selectedTaskIndex`: Currently selected task in the filtered list
- `selectedAgentIndex`: Currently selected agent (0-based, maps to 1-9 keys)
- `showTaskModal`: Whether task detail modal is visible
- `showCreateModal`: Whether create task modal is visible
- `showConfirmModal`: Whether confirmation dialog is visible
- `searchMode`: Whether in search input mode
- `searchInput`: Current search text
- `logFilterAgent`: Active agent name filter
- `logFilterType`: Active message type filter

### Modal Rendering
Modals are rendered as overlays on top of the main TUI:
- Centered on screen
- Styled with borders and padding
- Handle their own keyboard input when active

### Filter Implementation
Log filtering uses a multi-stage approach:
1. Filter by agent name (if set)
2. Filter by message type (if set)
3. Filter by search text (if set)
4. Display filtered results in log pane

## Testing

Comprehensive tests are provided in `internal/tui/interactive_test.go`:
- `TestTaskNavigation`: Tests arrow key navigation
- `TestTaskModalToggle`: Tests opening/closing task modal
- `TestCreateTaskModal`: Tests create task modal
- `TestAgentSelection`: Tests number key agent selection
- `TestConfirmModal`: Tests confirmation dialog
- `TestSearchMode`: Tests search mode entry/exit
- `TestFilterCycling`: Tests filter cycling

All tests pass successfully.

## Future Enhancements

Potential improvements for future iterations:
- Vim-style navigation (hjkl) without conflicts
- Mouse support for clicking on tasks/agents
- Multi-select for batch operations
- Task editing in modal
- Real-time log streaming in detail view
- Configurable keybindings
- Color themes
