# Agent Coordination

This directory is the handoff boundary between PM, BE, and FE agents.

## Operating rule

Only one active task may be in `CURRENT_TASK.md`. PM creates or updates it; the assigned agent implements it; PM or review verifies acceptance before the next task starts.

## Ownership

- PM owns scope, priority, acceptance criteria, decisions, and handoff.
- BE owns Go implementation, server rules, persistence, and server-side contract behavior.
- FE owns Flutter implementation, presentation state, interaction, and accessibility.
- Shared protocol changes require BE and FE agreement plus a PM task packet.

## Handoff sequence

```text
PM writes task -> BE/FE implements -> owner verifies -> review/PM accepts -> task archived
```

Do not use chat history as the only handoff. Record the minimum durable context in the coordination files.
