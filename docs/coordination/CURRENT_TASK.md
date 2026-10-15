# Task — REST room discovery, match history, and API server wiring

Status: DONE
Owner: BE
Priority: P1

## Goal

Wire the REST discovery (`GET /api/v1/rooms`), match history (`GET /api/v1/matches/{match_id}`), and WebSocket gateway (`/ws`) endpoints together with the existing guest endpoint (`POST /api/v1/guests`) into a single runnable HTTP server in `server/internal/transport/` and `server/cmd/api/main.go`.

## Acceptance criteria

- [x] `GET /api/v1/rooms?visibility=public` returns HTTP `200` with `{ "version": "v1", "request_id": "...", "rooms": [...] }` containing only waiting public rooms with open seats (`room_id`, `visibility`, `capacity`, `occupied_seats`, `status`) and never exposes `private_code` or player hands.
- [x] `GET /api/v1/rooms` with an unsupported `visibility` query parameter returns HTTP `400` with the standard v1 REST error shape (`version`, `code: "invalid_request"`, `message`, `request_id`).
- [x] `GET /api/v1/matches/{match_id}` returns HTTP `200` with `{ "version": "v1", "request_id": "...", "match_id": "...", "round_count": 3, "final_ranking": [...] }` from the match history store/repository, or HTTP `404` with `code: "not_found"` when the match ID does not exist.
- [x] Both new REST endpoints echo `X-Request-ID` when provided and generate a `request_id` when omitted.
- [x] `server/cmd/api/main.go` mounts `POST /api/v1/guests`, `GET /api/v1/rooms`, `GET /api/v1/matches/{match_id}`, and the WebSocket gateway (`/ws`) sharing the same guest store and room registry (and optional PostgreSQL repositories when `DATABASE_URL` is configured).
- [x] Handler and integration tests in `server/internal/transport/` cover happy paths, invalid query/path inputs, unknown match IDs (`404`), and verify private rooms/codes are excluded from public room discovery.
- [x] `make be-fmt`, `make be-test`, `cd server && go vet ./...`, and `cd server && go test -race ./...` pass.

## Allowed files/packages

- `server/cmd/api/...`
- `server/internal/transport/...`
- `docs/coordination/CURRENT_TASK.md`

## Out of scope

- Multi-round WebSocket progression (`match.updated` / `match.ended` / `state.resync` gateway handlers — next BE slice).
- Changes to the pure game engine (`server/internal/game/`).
- New database migrations (`server/migrations/`).
- Flutter client changes (`app/`).

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0011-protocol-v1-contract.md`, `docs/decisions/0013-backend-stack.md`.
- Protocol contract: `docs/protocol/contract.md` (REST v1 boundary: `POST /api/v1/guests`, `GET /api/v1/rooms`, `GET /api/v1/matches/{match_id}`).
- Architecture: `docs/ARCHITECTURE.md`.
- Blockers: None.

## Protocol impact

None (implements the already-specified REST v1 endpoints in `docs/protocol/contract.md`).

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
- Added `RoomDiscoveryHandler` (`server/internal/transport/rooms_handler.go`) serving `GET /api/v1/rooms?visibility=public` with `X-Request-ID` correlation, deterministic room ordering, strict `visibility=public` validation, and zero exposure of private rooms, private codes, or player hand data.
- Added `MatchHistoryStore` and `MatchHistoryHandler` (`server/internal/transport/matches_handler.go`) serving `GET /api/v1/matches/{match_id}` with rank-ordered `final_ranking`, `X-Request-ID` correlation, and `404 not_found` handling for unknown or invalid match IDs.
- Added unified `transport.Server` (`server/internal/transport/server.go`) and updated `server/cmd/api/main.go` to mount `POST /api/v1/guests`, `GET /api/v1/rooms`, `GET /api/v1/matches/{match_id}`, and `/ws` over shared guest, room, connection, and match history stores (with optional PostgreSQL `GuestRepository` and `MatchRepository` wiring when `DATABASE_URL` is set).
- Extended `GuestHandler` and `WebSocketGateway` (`server/internal/transport/guest_handler.go`, `server/internal/transport/gateway.go`, `server/internal/transport/jwt.go`) to support optional `GuestIdentityStore` persistence repositories alongside in-memory stores.
- Added unit and end-to-end integration tests in `server/internal/transport/rest_handlers_test.go`.

Verified:
- `make be-fmt` — PASS
- `make be-test` — PASS
- `cd server && go vet ./...` — PASS
- `cd server && go test -race ./...` — PASS

Known risks:
- None.

Next task:
- Complete WebSocket gateway end-to-end lifecycle (`state.resync`, 3-round match progression with `match.updated` & `match.ended`, per-recipient `sequence`, and command `request_id` idempotency).
