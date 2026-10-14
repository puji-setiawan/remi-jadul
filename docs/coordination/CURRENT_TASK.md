# Task — WebSocket host start and game.started broadcast

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement the authenticated `game.start` command. The host starts a waiting room with 2–4 players; the server generates and shuffles the deck server-side, creates the active round through the game engine, and broadcasts a personalized `game.started` event to connected room members.

## Acceptance criteria

- [x] Only the room host can successfully issue `game.start {}`.
- [x] `game.start` is rejected when the room is not waiting, already started, or has fewer than 2 members.
- [x] The server generates the standard 52-card deck and applies server-side randomness; clients never supply deck order or hands.
- [x] A successful start creates an active round and returns a personalized `game.started` event to the sender.
- [x] Every connected member of the room receives the personalized `game.started` event.
- [x] Each recipient receives only the sender-visible fields: own hand, hand counts, active joker rank, discard pile, table, turn/phase, closed deck count, scoreboard.
- [x] `game.started` does not expose opponent cards, marker identity, or deck order.
- [x] Rejected starts leave room, round, and recipient connection state unchanged.
- [x] Typed command errors are returned for unauthorized or invalid start conditions.
- [x] HTTP/WebSocket table-driven tests cover host start, non-host rejection, insufficient players, already-started rejection, and broadcast delivery.
- [x] `make be-fmt`, `make be-test`, `go vet ./...`, and `go test -race ./...` pass.

## Allowed files/packages

- `server/cmd/api/...`
- `server/internal/transport/...`
- `server/internal/room/...`
- `server/internal/game/...`
- `server/internal/realtime/...`

## Out of scope

- Gameplay commands: draw, discard, laydown, take discard, close.
- Active-joker discard, close validation, and scoring broadcasts.
- Reconnect round-trip delivery beyond an initial snapshot response path.
- Match result persistence and completed match history.
- Flutter integration.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0002-rules-determinism.md`, `docs/decisions/0004-three-round-match.md`, `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/commands.md`, `docs/protocol/events.md`, `docs/protocol/envelope.md`, `docs/protocol/visibility.md`.
- Architecture: `docs/ARCHITECTURE.md` authority, realtime, transport, and room boundaries.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 active game protocol and Phase 2 host start.
- Previous slices: room lifecycle, host start, game engine, personalized projection, authenticated connection registry, and personalized broadcast.
- Blockers: No product decision is currently blocking this slice. Server-side deck entropy must live outside the pure game engine.

## Protocol impact

Implements the documented `game.start` command and `game.started` event without changing their shapes.

## Verification

Expected commands:

```text
make be-fmt
make be-test
(cd server && go vet ./...)
(cd server && go test -race ./...)
```

## Handoff result

Changed:
- Added authenticated game.start command and personalized game.started broadcast.
Verified:
- make be-fmt
- make be-test
- go vet ./...
- go test -race ./...
Known risks:
- Server-side deck uses math/rand with time seed.
Next task:
