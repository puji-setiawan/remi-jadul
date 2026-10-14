# Task — WebSocket room create and join commands

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement authenticated WebSocket room lifecycle for create, join, and leave using the existing in-memory room registry. The gateway converts versioned commands into room mutations and responds with personalized room updates without exposing private codes or game state.

## Acceptance criteria

- [x] An authenticated `room.create` command creates a public or private in-memory room owned by the authenticated guest.
- [x] `room.create` accepts capacity 2, 3, or 4 only.
- [x] A private room creation accepts a caller-supplied access code; a public room does not expose or accept one.
- [x] An authenticated `room.join` command accepts either `room_id` or `private_code`.
- [x] Joining succeeds only for a waiting room with an open seat.
- [x] Joining a full, started, missing, or incorrectly coded room returns a typed error and leaves state unchanged.
- [x] An authenticated `room.leave` command removes the player; empty rooms are removed.
- [x] Responses are personalized: a private access code is exposed only to the creating host or a player who joined with it.
- [x] Public room lists are not delivered as a WebSocket broadcast; discovery remains REST-only per protocol.
- [x] Commands with invalid envelopes return the documented typed errors without mutating room state.
- [x] HTTP/WebSocket handler tests cover create public/private, join by ID/code, full/mismatched room rejection, invalid capacity, and unauthorized command behavior.
- [x] `make be-fmt`, `make be-test`, and `go vet ./...` pass.

## Allowed files/packages

- `server/cmd/api/...`
- `server/internal/transport/...`
- `server/internal/auth/...`
- `server/internal/room/...`

## Out of scope

- Host start, game start, and active round creation over WebSocket.
- Game commands: draw, discard, laydown, take discard, close.
- Personalized game-state broadcast and reconnect round-trip.
- Match result persistence and PostgreSQL repositories.
- Flutter integration and room lobby UX.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0003-public-discard-pile.md`, `docs/decisions/0009-public-table-combinations.md`, `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/commands.md`, `docs/protocol/envelope.md`, `docs/protocol/errors.md`, `docs/protocol/visibility.md`.
- Architecture: `docs/ARCHITECTURE.md` authority, realtime, and transport boundaries.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 2 room lifecycle and Phase 3 WebSocket protocol.
- Previous slices: guest identity, room lifecycle, JWT validation, command envelope, projection, sequence, reconnect behavior, and authenticated gateway.
- Blockers: No product decision is currently blocking this slice. WebSocket implementation must settle the coder/websocket dependency and keep package APIs vet-clean; the current `go vet` warning about copying `GuestStore` must be addressed if it affects this slice.

## Protocol impact

Implements the documented `room.create`, `room.join`, and `room.leave` commands without changing their shapes.

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
- Added room create/join/leave WebSocket command handling in server/internal/transport/gateway.go and registry support.
Verified:
- make be-fmt
- make be-test
- (cd server && go vet ./...)
- (cd server && go test -race ./...)
Known risks:
- Private code visibility and room update shape follow protocol as implemented here.
Next task:
