# Usability Testing Guide

## Overview

This guide provides comprehensive instructions for conducting usability testing of the Agent Stack Controller (asc). It covers test scenarios, evaluation criteria, and feedback collection methods.

## Test Objectives

1. Evaluate first-time user experience
2. Assess common workflow efficiency
3. Test error recovery mechanisms
4. Validate keyboard navigation and shortcuts
5. Verify terminal responsiveness
6. Ensure accessibility compliance
7. Gather qualitative user feedback
8. Document common issues and solutions

## Test Environment Setup

### Prerequisites

- Clean test environment (no existing asc installation)
- Various terminal emulators (iTerm2, Alacritty, Terminal.app, Windows Terminal)
- Different terminal sizes (80x24, 120x40, 200x60)
- Screen recording software for session capture
- Note-taking tools for observations

### Test User Profiles

1. **Novice Developer**: Limited CLI experience, new to agent systems
2. **Experienced Developer**: Comfortable with CLI, familiar with Docker/K8s
3. **Power User**: TUI expert, uses vim/tmux daily
4. **Accessibility User**: Relies on screen readers or high contrast

## Test Scenarios

### Scenario 1: First-Time User Experience

**Objective**: Evaluate the complete onboarding journey

**Steps**:
1. User discovers asc (via README or documentation)
2. User installs asc
3. User runs first command (likely `asc --help`)
4. User runs `asc init` to set up
5. User configures API keys
6. User starts agents with `asc up`
7. User views agent status in TUI
8. User stops agents with `asc down`

**Success Criteria**:
- User completes setup in < 15 minutes
- User understands what asc does
- User successfully starts and stops agents
- User knows where to find help

**Observations to Record**:
- Where does user get stuck?
- What questions does user ask?
- What documentation does user consult?
- What errors does user encounter?

### Scenario 2: Common Workflows

**Objective**: Test typical day-to-day usage patterns

**Workflow A: Check Before Start**

1. Run `asc check` to verify dependencies
2. Review check results
3. Fix any issues identified
4. Run `asc up` to start agents

**Success Criteria**: User understands check output and can fix issues

**Workflow B: Monitor Agent Activity**
1. Start agents with `asc up`
2. View agent status in TUI
3. Monitor task stream
4. Review interaction logs
5. Use keyboard shortcuts (r for refresh, t for test)

**Success Criteria**: User can interpret TUI information and use shortcuts

**Workflow C: Troubleshoot Issues**
1. Notice agent is stuck or erroring
2. View agent logs
3. Identify problem
4. Stop and restart agents
5. Verify fix

**Success Criteria**: User can diagnose and resolve common issues

### Scenario 3: Error Recovery

**Objective**: Test how users handle and recover from errors

**Error A: Missing Configuration**
- User runs `asc up` without config
- Expected: Clear error message suggesting `asc init`
- User runs `asc init` to create config
- User successfully starts agents

**Error B: Invalid Configuration**
- User edits asc.toml and introduces syntax error
- User runs `asc check`
- Expected: Clear error pointing to line number
- User fixes syntax error
- User successfully starts agents

**Error C: Missing Dependencies**
- User runs `asc check` with missing dependencies
- Expected: List of missing items with installation instructions
- User installs missing dependencies
- User successfully starts agents

**Error D: Port Conflict**
- User starts asc when port 8765 is in use
- Expected: Clear error with port number and suggestion
- User stops conflicting service or changes port
- User successfully starts agents

**Error E: API Key Issues**
- User starts agents without API keys
- Expected: Agent fails with clear authentication error
- User adds API keys to .env
- User restarts and agents work

**Success Criteria**:
- Error messages are clear and actionable
- User can resolve errors in < 5 minutes
- User doesn't need to consult external resources

### Scenario 4: Keyboard Navigation

**Objective**: Test TUI keyboard interactions

**Navigation Tests**:
1. Launch TUI with `asc up`
2. Press 'q' to quit (should prompt if agents running)
3. Press 'r' to refresh (should update all panes)
4. Press 't' to run test (should show results in log)
5. Use arrow keys to scroll in panes
6. Use Tab to switch between panes
7. Press '?' or 'h' for help (if implemented)

**Success Criteria**:
- All shortcuts work as expected
- Shortcuts are discoverable (shown in footer)
- Keyboard-only navigation is complete
- No mouse required for any function

### Scenario 5: Terminal Resize

**Objective**: Test responsive layout behavior

**Resize Tests**:
1. Start TUI at 80x24 (minimum size)
2. Verify layout is usable
3. Resize to 120x40 (medium size)
4. Verify layout adjusts smoothly
5. Resize to 200x60 (large size)
6. Verify layout uses space well
7. Resize to very wide (200x30)
8. Verify horizontal layout works
9. Resize to very tall (80x60)
10. Verify vertical space is used

**Success Criteria**:
- No visual corruption during resize
- Layout adjusts immediately
- Content reflows appropriately
- Minimum size warning if too small

### Scenario 6: Accessibility

**Objective**: Verify accessibility features

**High Contrast Mode**:
1. Enable high contrast mode (if available)
2. Verify all text is readable
3. Verify status indicators are clear
4. Check WCAG contrast ratios

**Color Blind Testing**:
1. View TUI with deuteranopia simulation
2. Verify status is clear without color
3. Check that icons/symbols convey status
4. Test with protanopia and tritanopia

**Screen Reader Testing**:
1. Run CLI commands with screen reader
2. Verify output is readable
3. Check that status information is clear
4. Test help text readability

**Keyboard-Only Testing**:
1. Use TUI without mouse
2. Verify all features accessible
3. Check tab order is logical
4. Verify shortcuts are documented

**Success Criteria**:
- Meets WCAG 2.1 Level AA standards
- Usable without color perception
- Screen reader compatible
- Fully keyboard accessible

## Evaluation Criteria

### Usability Metrics

**Time to First Success**
- Measure: Time from install to first successful agent run
- Target: < 5 minutes (experienced), < 15 minutes (novice)

**Task Completion Rate**
- Measure: Percentage of users completing tasks without help
- Target: > 90% for common tasks

**Error Recovery Time**
- Measure: Time from error to resolution
- Target: < 2 minutes for common errors

**Command Discoverability**
- Measure: Can user find command without docs?
- Target: > 80% find commands via help

**Cognitive Load**
- Measure: Number of concepts user must understand
- Target: Minimize required knowledge

### Qualitative Assessment

**Clarity**
- Are instructions clear?
- Are error messages helpful?
- Is feedback immediate?

**Efficiency**
- Are workflows streamlined?
- Are shortcuts available?
- Is information dense but not overwhelming?

**Satisfaction**
- Does user feel in control?
- Is experience pleasant?
- Would user recommend to others?

**Learnability**
- Can user remember commands?
- Are patterns consistent?
- Is help easily accessible?

## Feedback Collection

### During Testing

**Think-Aloud Protocol**
- Ask user to verbalize thoughts
- Note confusion points
- Record questions asked
- Observe hesitations

**Screen Recording**
- Record entire session
- Capture mouse/keyboard input
- Note timestamps of issues
- Review for patterns

**Observer Notes**
- Document user actions
- Note unexpected behaviors
- Record error encounters
- Track time on tasks

### Post-Testing

**Interview Questions**
1. What was your first impression?
2. What was most confusing?
3. What was most helpful?
4. What would you change?
5. Would you use this tool?
6. How does it compare to alternatives?

**Satisfaction Survey**
- Rate ease of use (1-5)
- Rate documentation quality (1-5)
- Rate error messages (1-5)
- Rate overall experience (1-5)

**System Usability Scale (SUS)**
- Administer standard SUS questionnaire
- Calculate SUS score
- Target: > 70 (above average)

## Common Issues Documentation

### Issue Template

For each identified issue, document:

**Issue**: Brief description
**Frequency**: How often observed (% of users)
**Severity**: Critical / High / Medium / Low
**User Impact**: How it affects user experience
**Current Behavior**: What happens now
**Expected Behavior**: What should happen
**Workaround**: Temporary solution (if any)
**Proposed Fix**: How to resolve permanently
**Priority**: P0 / P1 / P2 / P3

### Example Issues

**Issue: API Keys Not Loaded**
- Frequency: 40% of users
- Severity: High
- Impact: Agents can't authenticate
- Current: Generic error message
- Expected: Clear message about .env file
- Workaround: Check .env location and permissions
- Fix: Improve error message, add to check command
- Priority: P1

**Issue: TUI Display Corrupted**
- Frequency: 10% of users
- Severity: Medium
- Impact: Can't read TUI
- Current: Garbled display
- Expected: Clean display or warning
- Workaround: Use different terminal
- Fix: Detect terminal capabilities, show warning
- Priority: P2

## Testing Schedule

### Alpha Testing (Internal)
- Duration: 1 week
- Participants: Development team
- Focus: Basic functionality, critical bugs

### Beta Testing (External)
- Duration: 2-4 weeks
- Participants: 10-20 volunteers
- Focus: Real-world usage, edge cases

### Usability Testing (Moderated)
- Duration: 1 week
- Participants: 5-10 representative users
- Focus: Detailed observation, feedback

### Continuous Testing
- Ongoing user feedback collection
- Issue tracking and resolution
- Iterative improvements

## Test Execution

### Before Testing

1. Prepare test environment
2. Install screen recording software
3. Prepare note-taking templates
4. Brief test participants
5. Obtain consent for recording

### During Testing

1. Welcome participant
2. Explain think-aloud protocol
3. Start recording
4. Present scenarios
5. Observe without interfering
6. Take detailed notes
7. Ask clarifying questions

### After Testing

1. Stop recording
2. Conduct interview
3. Administer surveys
4. Thank participant
5. Organize notes and recordings
6. Identify patterns and issues

## Analysis and Reporting

### Data Analysis

**Quantitative Data**
- Calculate completion rates
- Measure task times
- Compute SUS scores
- Identify error frequencies

**Qualitative Data**
- Transcribe interviews
- Code observations
- Identify themes
- Extract quotes

### Report Structure

1. **Executive Summary**
   - Key findings
   - Critical issues
   - Recommendations

2. **Methodology**
   - Participants
   - Scenarios
   - Metrics

3. **Results**
   - Quantitative findings
   - Qualitative findings
   - User quotes

4. **Issues and Recommendations**
   - Prioritized issue list
   - Proposed solutions
   - Implementation plan

5. **Appendices**
   - Raw data
   - Recordings
   - Transcripts

## Continuous Improvement

### Feedback Channels

**In-App Feedback**
- Add feedback command
- Collect usage statistics
- Track error rates

**Community Feedback**
- GitHub issues
- Discussion forums
- User surveys

**Analytics**
- Command usage frequency
- Error occurrence rates
- Feature adoption

### Iteration Cycle

1. Collect feedback
2. Analyze patterns
3. Prioritize issues
4. Implement fixes
5. Test improvements
6. Release updates
7. Repeat

## Success Criteria

### Launch Readiness

- [ ] SUS score > 70
- [ ] Task completion rate > 90%
- [ ] No P0 issues
- [ ] < 5 P1 issues
- [ ] Documentation complete
- [ ] Accessibility compliant
- [ ] Performance acceptable

### Post-Launch Goals

- Maintain SUS score > 75
- Reduce error recovery time
- Increase feature discoverability
- Improve documentation based on feedback
- Address all P1 issues within 2 weeks

## Resources

### Tools

- **Screen Recording**: OBS Studio, QuickTime, Windows Game Bar
- **Analytics**: Telemetry (if implemented)
- **Surveys**: Google Forms, Typeform
- **Note-Taking**: Notion, Evernote, Google Docs

### References

- Nielsen Norman Group: Usability Testing 101
- WCAG 2.1 Guidelines
- System Usability Scale (SUS)
- Think-Aloud Protocol Guide

## Appendix: Test Scripts

### Script 1: First-Time User

```
Welcome! Today you'll be trying out a new tool called asc.
Please think aloud as you work - tell me what you're thinking.
There are no wrong answers, and you can't break anything.

Task 1: Install and set up asc
- You've just heard about asc and want to try it
- Start from the README
- Get to the point where agents are running

Task 2: Monitor your agents
- View the status of your agents
- Check what tasks they're working on
- View the interaction logs

Task 3: Stop the agents
- Shut down the agent stack cleanly
```

### Script 2: Error Recovery

```
You're working with asc and encounter some issues.
Let's see how you handle them.

Scenario 1: Missing Config
- Try to start asc without configuration
- Resolve the issue

Scenario 2: Invalid Config
- Your config file has a syntax error
- Find and fix the error

Scenario 3: Port Conflict
- The MCP server port is already in use
- Resolve the conflict
```

### Script 3: Advanced Features

```
You're now familiar with asc basics.
Let's explore some advanced features.

Task 1: Add a new agent
- Add another agent to your configuration
- Configure it with different phases
- Start it up

Task 2: Troubleshoot an issue
- One of your agents is stuck
- Diagnose the problem
- Fix it

Task 3: Customize your setup
- Change the refresh interval
- Modify agent phases
- Test your changes
```

## Conclusion

Usability testing is an ongoing process. Use this guide to:
- Conduct structured testing sessions
- Collect meaningful feedback
- Identify and prioritize issues
- Continuously improve the user experience

Remember: The goal is to make asc intuitive, efficient, and pleasant to use for developers of all skill levels.
