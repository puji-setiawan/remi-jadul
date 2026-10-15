# Task — Complete WebSocket match lifecycle

Status: READY
Owner: BE
Priority: P1

## Goal

Complete the authenticated WebSocket lifecycle after room start: personalized resync, per-recipient sequence handling, request idempotency, and three-round match completion events. Preserve server authority and the existing v1 protocol contract.

## Acceptance criteria

- [ ] `state.resync` returns a full personalized snapshot with the documented replacement sequence metadata.
- [ ] Each recipient has an independent monotonic sequence stream; private events do not create gaps for other recipients.
- [ ] Repeating an accepted command `request_id` returns the original result without executing the command twice.
- [ ] Repeating a rejected command `request_id` returns the original error without mutating state.
- [ ] Reusing a request ID with a different payload returns `invalid_request`.
- [ ] A completed round emits `round.ended` before `match.updated`.
- [ ] After round three, the server emits `match.ended` with equal-score ties sharing a rank.
- [ ] Tests cover reconnect, stale sequence, duplicate request, wrong turn, hidden-card visibility, and three-round completion.

## Allowed files/packages

- `server/internal/transport/...`
- `server/internal/realtime/...`
- `server/internal/room/...` only when required by the lifecycle interface
- `server/cmd/api/...` only for wiring required by this task
- `docs/coordination/CURRENT_TASK.md`

## Out of scope

- Flutter implementation.
- Pure game engine rule changes.
- New REST endpoints.
- New database migrations.
- Protocol shape changes. Any required contract change must stop and create a decision packet first.

## Dependencies and decisions

- Protocol: `docs/protocol/contract.md`, `docs/protocol/examples.md`, `docs/protocol/envelope.md`, `docs/protocol/events.md`, `docs/protocol/visibility.md`.
- Decision: `docs/decisions/0011-protocol-v1-contract.md`.
- Rules: `docs/GAME_RULES.md`.
- Previous task: REST room discovery, match history, and API server wiring.
- Blockers: None known.

## Verification

Expected commands:

```text
make be-fmt
make be-test
cd server && go vet ./...
cd server && go test -race ./...
```

## Handoff result

Changed:
Verified:
Known risks:
Next task:
