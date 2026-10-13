# Remi Jadul — Roadmap

## Phase 0 — Repository and agent foundation

- Keep specifications, rules, decision log, OpenCode commands, environment examples, and safe local development docs consistent.
- Validate that OpenCode loads root instructions and project commands.

## Phase 1 — Pure Go game engine foundation

- Card, deck, deal, joker, and combination validation.
- Table-driven tests for valid, invalid, and boundary rules.
- Use only accepted rules from `GAME_RULES.md` and `docs/decisions/`.

## Phase 1B — Complete game state engine

- Turn transitions, closed-deck draw, discard pickup, close, game-over, and deterministic scoring.
- Atomic state-transition tests and scoring partition tests.
- No implementation begins while an affected rule is still `Proposed`.

## Phase 2 — Room and guest identity

- Guest nickname identity.
- Public/private room lifecycle.
- Host start and 2–4 player validation.
- In-memory active rooms.

## Phase 3 — REST and WebSocket protocol

- Finalize the versioned JSON contract in `docs/protocol/` before room/realtime code.
- Personalized state projections.
- Action errors, event ordering, reconnect, and visibility tests.

## Phase 4 — PostgreSQL persistence

- Guest/profile data and completed match results.
- Migrations and repository integration tests.

## Phase 5 — Flutter Android client

- Lobby, room, hand/table view, turn actions, result/ranking, reconnect UX.
- Client contract tests and Android emulator acceptance flow.

## Phase 6 — MVP hardening

- Dockerized backend workflow, CI gates, structured logs, race detection, security review, and release checklist.
