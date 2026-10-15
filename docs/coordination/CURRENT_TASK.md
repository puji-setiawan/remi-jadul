# Task — Complete WebSocket match lifecycle

Status: DONE
Owner: BE
Priority: P1

## Goal

Complete the authenticated WebSocket lifecycle after room start: personalized resync, per-recipient sequence handling, request idempotency, and three-round match completion events. Preserve server authority and the existing v1 protocol contract.

## Acceptance criteria

- [x] `state.resync` returns a full personalized snapshot with the documented replacement sequence metadata.
- [x] Each recipient has an independent monotonic sequence stream; private events do not create gaps for other recipients.
- [x] Repeating an accepted command `request_id` returns the original result without executing the command twice.
- [x] Repeating a rejected command `request_id` returns the original error without mutating state.
- [x] Reusing a request ID with a different payload returns `invalid_request`.
- [x] A completed round emits `round.ended` before `match.updated`.
- [x] After round three, the server emits `match.ended` with equal-score ties sharing a rank.
- [x] Tests cover reconnect, stale sequence, duplicate request, wrong turn, hidden-card visibility, and three-round completion.

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
Implemented personalized `state.resync`, recipient-scoped sequence assignment, multi-event request idempotency replay, per-room command serialization, pickup game-over completion, ordered lifecycle events, actual round numbers, cumulative scores, and display-name match payloads. Request claims are now atomic and in-flight duplicates wait for the completed response; validation/action failures are cached when a request ID is present, while malformed JSON remains an uncached safe protocol error. Room join now uses the same per-room mutation lock as start, leave, and game actions. Fixed active-joker discard and closed-deck exhaustion to emit the shared round lifecycle, added laydown room serialization, and expanded wire-level lifecycle coverage for replay, malformed input, and event ordering in `server/internal/transport/gateway_lifecycle_test.go`.
Verified: `make be-fmt`; `make be-test`; `cd server && go vet ./...`; `cd server && go test -race ./...` all passed. The focused lifecycle transport tests also passed with `cd server && go test ./internal/transport -run 'TestGatewayLifecycle' -count=1`.
Known risks: lifecycle integration remains in-memory and uses a fresh randomized deck for each next round, consistent with the existing server architecture. Existing reconnect/resync, visibility, discard, and deck-exhaustion wire coverage remains distributed across the transport/realtime test files rather than duplicated in the lifecycle test.
Next task:
