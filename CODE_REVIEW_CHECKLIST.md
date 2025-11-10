# Code Review Checklist

Use this checklist when reviewing pull requests to ensure consistent, high-quality code reviews.

## General

- [ ] **PR description is clear** - Explains what, why, and how
- [ ] **Linked to relevant issues** - References issue numbers (Fixes #123)
- [ ] **Appropriate size** - PR is focused and not too large (< 500 lines preferred)
- [ ] **No merge conflicts** - Branch is up to date with main
- [ ] **CI checks pass** - All automated tests and lints pass

## Code Quality

### Design & Architecture

- [ ] **Follows existing patterns** - Consistent with codebase architecture
- [ ] **Appropriate abstraction** - Not over-engineered or under-engineered
- [ ] **Single responsibility** - Functions/types do one thing well
- [ ] **Minimal coupling** - Dependencies are minimal and well-defined
- [ ] **Proper error handling** - Errors are wrapped with context
- [ ] **No code duplication** - DRY principle followed

### Go Specifics

- [ ] **Follows Go conventions** - Idiomatic Go code
- [ ] **Proper naming** - Clear, descriptive names (not too short or long)
- [ ] **Exported symbols documented** - All public APIs have godoc comments
- [ ] **Package documentation** - Package has a clear purpose statement
- [ ] **Error handling** - All errors are checked and handled appropriately
- [ ] **Context usage** - Context passed correctly for cancellation/timeouts
- [ ] **Goroutine safety** - Concurrent access properly synchronized
- [ ] **Resource cleanup** - defer used for cleanup, no resource leaks

### Code Style

- [ ] **Formatted with gofmt** - Code is properly formatted
- [ ] **No linter warnings** - golangci-lint passes
- [ ] **No commented code** - Dead code removed
- [ ] **Consistent style** - Matches existing codebase style
- [ ] **Readable** - Code is easy to understand
- [ ] **Comments explain why** - Not what (code should be self-documenting)

## Testing

### Test Coverage

- [ ] **Tests included** - New functionality has tests
- [ ] **Tests pass** - All tests pass locally and in CI
- [ ] **Sufficient coverage** - 80%+ coverage for new code
- [ ] **Edge cases tested** - Boundary conditions covered
- [ ] **Error paths tested** - Failure scenarios covered
- [ ] **Table-driven tests** - Multiple cases use table-driven approach

### Test Quality

- [ ] **Tests are focused** - Each test tests one thing
- [ ] **Tests are independent** - Can run in any order
- [ ] **Tests are deterministic** - Same input = same output
- [ ] **No flaky tests** - Tests don't randomly fail
- [ ] **Fast tests** - Unit tests run quickly (< 1s each)
- [ ] **Clear test names** - TestFunction_Scenario_ExpectedBehavior
- [ ] **Mocks used appropriately** - External dependencies mocked

## Security

- [ ] **No hardcoded secrets** - API keys, passwords in env vars
- [ ] **Input validation** - User input is validated and sanitized
- [ ] **No SQL injection** - Queries use parameterization
- [ ] **No command injection** - Shell commands properly escaped
- [ ] **No path traversal** - File paths validated
- [ ] **Proper permissions** - Files have appropriate permissions
- [ ] **Dependencies secure** - No known vulnerabilities

## Performance

- [ ] **No obvious bottlenecks** - Algorithms are efficient
- [ ] **Appropriate data structures** - Right tool for the job
- [ ] **No unnecessary allocations** - Memory usage optimized
- [ ] **Concurrent operations safe** - No race conditions
- [ ] **Database queries optimized** - Indexes used, N+1 avoided
- [ ] **Caching used appropriately** - Expensive operations cached

## Documentation

- [ ] **README updated** - If user-facing changes
- [ ] **API docs updated** - If public API changes
- [ ] **Comments added** - Complex logic explained
- [ ] **Examples provided** - For new features
- [ ] **Migration guide** - If breaking changes
- [ ] **Changelog updated** - Notable changes documented

## User Experience

- [ ] **Error messages helpful** - Clear, actionable error messages
- [ ] **Logging appropriate** - Right level (debug/info/warn/error)
- [ ] **CLI output clear** - User-friendly output formatting
- [ ] **TUI responsive** - UI updates smoothly
- [ ] **Backwards compatible** - Or breaking changes documented

## Specific to asc

### Configuration

- [ ] **Config validation** - Invalid configs rejected with clear errors
- [ ] **Defaults sensible** - Good defaults for optional fields
- [ ] **Hot-reload safe** - Config changes don't break running system

### Process Management

- [ ] **Graceful shutdown** - Processes cleaned up properly
- [ ] **PID tracking** - PIDs saved and cleaned up
- [ ] **Log rotation** - Logs don't grow unbounded
- [ ] **Resource cleanup** - No zombie processes or file descriptor leaks

### TUI

- [ ] **Responsive layout** - Handles terminal resize
- [ ] **Keyboard shortcuts work** - All documented shortcuts functional
- [ ] **Visual consistency** - Follows vaporwave theme
- [ ] **Performance** - Renders smoothly (60fps)
- [ ] **Error display** - Errors shown clearly in UI

### Agent Integration

- [ ] **Environment variables set** - Agents receive correct env vars
- [ ] **Communication works** - MCP and beads integration functional
- [ ] **Heartbeat monitoring** - Agent health tracked correctly
- [ ] **Error recovery** - Failed agents handled gracefully

## Review Process

### Before Approving

- [ ] **Tested locally** - Checked out and tested the changes
- [ ] **All comments addressed** - Author responded to all feedback
- [ ] **No blocking issues** - All critical issues resolved
- [ ] **Documentation reviewed** - Docs are accurate and complete

### Feedback Guidelines

When providing feedback:

- **Be specific** - Point to exact lines and explain why
- **Be constructive** - Suggest improvements, don't just criticize
- **Be kind** - Focus on code, not the person
- **Distinguish severity**:
  - 🔴 **Blocking** - Must be fixed before merge
  - 🟡 **Non-blocking** - Should be fixed but not critical
  - 🟢 **Nit** - Minor style/preference issue
  - 💡 **Suggestion** - Optional improvement idea

### Example Comments

**Good:**
```
🔴 This could cause a race condition when multiple goroutines access the map.
Consider using sync.Map or adding a mutex.
```

**Good:**
```
💡 This function is getting long. Consider extracting the validation logic
into a separate validateConfig() function for better readability.
```

**Bad:**
```
This is wrong.
```

**Bad:**
```
Why didn't you use a better approach?
```

## After Review

- [ ] **Approved** - If all checks pass and no blocking issues
- [ ] **Request changes** - If blocking issues found
- [ ] **Comment** - If non-blocking feedback provided
- [ ] **Follow up** - Check that requested changes are made

## Quick Reference

### Approval Criteria

Approve when:
- ✅ All checklist items pass (or have good reason not to)
- ✅ Code quality is high
- ✅ Tests are comprehensive
- ✅ Documentation is complete
- ✅ No security concerns
- ✅ CI checks pass

Request changes when:
- ❌ Security vulnerabilities
- ❌ Breaking changes without migration path
- ❌ Missing tests for critical functionality
- ❌ Poor error handling
- ❌ Performance issues
- ❌ Doesn't follow project conventions

### Common Issues

Watch out for:
- Hardcoded values that should be configurable
- Missing error handling
- Race conditions in concurrent code
- Resource leaks (goroutines, file handles, connections)
- Overly complex logic that could be simplified
- Missing documentation for exported symbols
- Tests that don't actually test the functionality
- Flaky tests that sometimes fail

## Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Contributing Guide](CONTRIBUTING.md)
- [Testing Guidelines](CONTRIBUTING.md#testing-guidelines)

---

Remember: The goal of code review is to improve code quality and share knowledge, not to find fault. Be thorough but kind, and focus on helping the contributor succeed.
