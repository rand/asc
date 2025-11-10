# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records documenting key design decisions made during the development of the Agent Stack Controller.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences. ADRs help teams understand why certain decisions were made and provide historical context for future changes.

## ADR Format

Each ADR follows this structure:

```markdown
# ADR-XXXX: Title

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
What is the issue we're facing?

## Decision
What decision did we make?

## Consequences
What are the positive and negative consequences?

## Alternatives Considered
What other options did we evaluate?
```

## Index of ADRs

### Core Architecture

- [ADR-0001: Use Go for CLI and TUI](adr/ADR-0001-go-for-cli-tui.md) - **Accepted**
- [ADR-0002: Use Python for Agent Adapter](adr/ADR-0002-python-agent-adapter.md) - **Accepted**
- [ADR-0003: Layered Architecture Design](adr/ADR-0003-layered-architecture.md) - **Accepted**
- [ADR-0004: Process Management Strategy](adr/ADR-0004-process-management.md) - **Accepted**

### UI/UX

- [ADR-0005: Bubbletea for TUI Framework](adr/ADR-0005-bubbletea-tui.md) - **Accepted**
- [ADR-0006: Vaporwave Aesthetic Design](adr/ADR-0006-vaporwave-aesthetic.md) - **Accepted**
- [ADR-0007: Three-Pane Dashboard Layout](adr/ADR-0007-three-pane-layout.md) - **Accepted**

### Configuration

- [ADR-0008: TOML for Configuration Format](adr/ADR-0008-toml-configuration.md) - **Accepted**
- [ADR-0009: Configuration Templates System](adr/ADR-0009-configuration-templates.md) - **Accepted**
- [ADR-0010: Hot-Reload Configuration](adr/ADR-0010-hot-reload-config.md) - **Accepted**

### Security

- [ADR-0011: Age Encryption for Secrets](adr/ADR-0011-age-encryption.md) - **Accepted**
- [ADR-0012: File Permission Strategy](adr/ADR-0012-file-permissions.md) - **Accepted**
- [ADR-0013: API Key Management](adr/ADR-0013-api-key-management.md) - **Accepted**

### Communication

- [ADR-0014: MCP Agent Mail Protocol](adr/ADR-0014-mcp-protocol.md) - **Accepted**
- [ADR-0015: WebSocket for Real-Time Updates](adr/ADR-0015-websocket-updates.md) - **Accepted**
- [ADR-0016: Polling Fallback Strategy](adr/ADR-0016-polling-fallback.md) - **Accepted**

### Task Management

- [ADR-0017: Beads for Task Database](adr/ADR-0017-beads-task-db.md) - **Accepted**
- [ADR-0018: Git-Backed Task State](adr/ADR-0018-git-backed-state.md) - **Accepted**

### Agent Design

- [ADR-0019: Matrix Architecture for Agents](adr/ADR-0019-matrix-architecture.md) - **Accepted**
- [ADR-0020: Hephaestus Phase Loop Pattern](adr/ADR-0020-hephaestus-pattern.md) - **Accepted**
- [ADR-0021: ACE Playbook Learning System](adr/ADR-0021-ace-playbook.md) - **Accepted**
- [ADR-0022: Multi-LLM Support Strategy](adr/ADR-0022-multi-llm-support.md) - **Accepted**

### Testing

- [ADR-0023: Testing Strategy and Coverage](adr/ADR-0023-testing-strategy.md) - **Accepted**
- [ADR-0024: E2E Test Environment](adr/ADR-0024-e2e-environment.md) - **Accepted**

### Operations

- [ADR-0025: Logging and Observability](adr/ADR-0025-logging-observability.md) - **Accepted**
- [ADR-0026: Health Monitoring System](adr/ADR-0026-health-monitoring.md) - **Accepted**
- [ADR-0027: Auto-Recovery Strategy](adr/ADR-0027-auto-recovery.md) - **Accepted**

## Creating a New ADR

1. Copy the template: `cp docs/adr/ADR-TEMPLATE.md docs/adr/ADR-XXXX-title.md`
2. Fill in the sections
3. Submit for review
4. Update this index
5. Link from relevant documentation

## ADR Lifecycle

```
Proposed → Accepted → [Deprecated | Superseded]
```

- **Proposed**: Under discussion
- **Accepted**: Decision made and implemented
- **Deprecated**: No longer recommended but not replaced
- **Superseded**: Replaced by a newer ADR

## See Also

- [Design Document](../.kiro/specs/agent-stack-controller/design.md)
- [Requirements](../.kiro/specs/agent-stack-controller/requirements.md)
- [Project Structure](PROJECT_STRUCTURE.md)
