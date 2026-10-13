# Remi Jadul

Server-authoritative online Remi for Android.

## Repository layout

- `AGENTS.md` — instructions and quality guardrails for coding agents.
- `docs/SPEC.md` — product scope and MVP flows.
- `docs/PRODUCT_BACKLOG.md` — product backlog and delivery scope.
- `docs/GAME_RULES.md` — authoritative detailed game rules.
- `docs/ARCHITECTURE.md` — system boundaries and data flow.
- `docs/ROADMAP.md` — implementation phases.
- `docs/decisions/` — durable architecture and product decisions.
- `.opencode/` — compact OpenCode commands and token policy.
- `server/` — Go backend boundary; implementation starts in Phase 1.
- `app/` — Flutter Android client boundary; implementation starts in Phase 5.

## Recommended OpenCode workflow

Use one scoped task at a time:

```text
/plan implement the pure deck and deal model
/build implement the approved deck and deal plan
/verify deck and deal
/review deck and deal diff
```

For the three-agent workflow:

```text
/pm define the next pure deck task
/be implement the active deck task
/fe implement the active UI task
/handoff close the active task
```

PM owns `docs/coordination/CURRENT_TASK.md`; BE and FE must not work without an assigned packet.

For simple tasks, skip review. For rules, protocol, security, or cross-package changes, use review.

## Local setup

1. Copy `.env.example` to `.env`.
2. Start local dependencies with `docker compose up -d`.
3. Follow the current phase in `docs/ROADMAP.md`.

Stable verification entrypoints are available in `Makefile`: `make be-test`, `make be-fmt`, `make fe-test`, and `make fe-analyze`.

## Current blockers

The currently accepted rules decisions are recorded in `docs/decisions/0002-rules-determinism.md`. New rule ambiguity must be recorded as a proposed decision before implementation.
