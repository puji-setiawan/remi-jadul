# Task — Authenticated WebSocket command gateway

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement the first authenticated WebSocket gateway boundary. The server validates the JWT from the HTTP upgrade request, establishes a guest-scoped connection, and validates the versioned command envelope before any game mutation or broadcast is introduced.

## Acceptance criteria

- [x] The WebSocket upgrade requires an `Authorization: Bearer <jwt>` header.
- [x] A missing, malformed, expired, or wrong-issuer token rejects the upgrade before protocol use.
- [x] A valid token resolves to the existing guest identity.
- [x] Each accepted connection has a unique recipient stream initialized at the documented initial sequence.
- [x] An invalid command envelope receives a typed safe action error.
- [x] A valid envelope with an unknown command receives `invalid_command` without mutating state.
- [x] No game, room, or round state is mutated during connection validation and envelope checks.
- [x] The gateway does not accept game actions or broadcast personalized projections in this slice.
- [x] HTTP/WebSocket tests cover missing token, invalid token, valid upgrade, and invalid command handling.
- [x] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/cmd/api/...`
- `server/internal/transport/...`
- `server/internal/auth/...`

## Out of scope

- Room creation, discovery, join, leave, or host start over WebSocket.
- Game commands such as draw, discard, laydown, take discard, close, and resync transport execution.
- Personalized broadcast and multi-recipient delivery.
- Match result persistence and PostgreSQL integration.
- WebSocket reconnect, resync transport, and seq gap recovery over a live connection.
- Flutter integration.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0011-protocol-v1-contract.md`, `docs/decisions/0013-backend-stack.md`.
- Protocol docs: `docs/protocol/envelope.md`, `docs/protocol/errors.md`, `docs/protocol/contract.md`.
- Architecture: `docs/ARCHITECTURE.md` authority and transport boundaries.
- Authentication: existing JWT validation and guest identity store.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 REST/WebSocket protocol.
- Blockers: WebSocket dependency and upgrade behavior must be compatible with the existing Gin/net-http choices already in the codebase.

## Protocol impact

Implements the documented authenticated WebSocket connection boundary and reused envelope/error model without changing shapes.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Implemented pure JWT validation and guest resolution in server/internal/auth/jwt.go, with flexible caller configuration.
- Added transport validation aliases in server/internal/transport/jwt.go.
Verified:
- make be-fmt
- make be-test
- make fe-test
- make fe-analyze
Known risks:
- Validation requires a configured guest store; unknown or missing guest IDs are rejected through store lookup.
Next task:
