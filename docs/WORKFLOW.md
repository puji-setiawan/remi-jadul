# Vibe Coding Workflow

## The short loop

Use one small task per session:

```text
/plan <one behavior>
/build <approved plan>
/verify <changed behavior>
/review <diff>          # only for risky changes
```

The preferred first implementation slice is the pure Go deck/deal model, then combination validation, then turn transitions. Do not start with a large Flutter screen or a full realtime stack.

## Task packet

Every task should contain only:

```text
Goal: one observable behavior
Scope: files/packages allowed to change
Acceptance: 3–7 concrete checks
Tests: focused test command or cases
Out of scope: adjacent work to leave untouched
```

If a task cannot fit this packet, split it before coding.

## Context budget

| Task | Required context | Avoid loading |
|---|---|---|
| Game engine | `AGENTS.md`, `GAME_RULES.md`, relevant Go files | Flutter and visual docs |
| REST/WebSocket | `AGENTS.md`, `ARCHITECTURE.md`, protocol files | Full game UI |
| Flutter UI | `AGENTS.md`, `SPEC.md`, `DESIGN.md`, relevant client files | Backend internals unless contract changes |
| Database | `AGENTS.md`, `ARCHITECTURE.md`, migration files | UI and visual docs |
| Review | Diff, changed tests, one relevant spec | Entire repository |

Anti-slop loading follows the same budget: backend/rules tasks use the project rules and code guardrails; UI/review tasks additionally load the relevant anti-slop skill and `DESIGN.md`.

## Stop early

Stop instead of coding when a rule is ambiguous, a contract is missing, or a task crosses more than one subsystem without an agreed interface. Record the decision in `docs/decisions/` and continue only after it is accepted.

## Completion format

```text
Done: <behavior>
Changed: <files>
Verified: <commands and results>
Risk: <one line, or none>
Next: <one concrete task, if needed>
```
