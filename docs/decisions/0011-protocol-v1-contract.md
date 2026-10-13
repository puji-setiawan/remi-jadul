# 0011 — Protocol v1 transport, state, and retry contract

## Status

Accepted

## Context

BE and FE need one explicit contract before room and realtime implementation. The game must preserve server authority, private hands, event ordering, reconnect safety, and retry behavior.

## Decision

- WebSocket JSON v1 carries active room and game commands, events, errors, and resynchronization.
- REST is used for guest identity, room discovery, and completed match history.
- Public room discovery is REST-only. Room create, join, and leave remain WebSocket commands.
- Every accepted action produces a personalized complete `game.state` projection. Private draw information is sent through `game.drawn` only to the drawing player.
- `request_id` is idempotent across the authenticated guest session for every client command, including room creation before a room exists. REST guest creation uses an `Idempotency-Key` header. Repeating an accepted or rejected request returns the original result without executing twice. Reusing an ID with a different payload returns `invalid_request`.
- Server events use monotonic per-recipient room-stream sequence numbers. Private events do not create gaps for recipients that did not receive them. Clients resync using their last received sequence after a gap or reconnect.
- Resync returns a full personalized snapshot for MVP instead of replaying individual events.
- Guest identity is created through REST; the WebSocket handshake authenticates the returned opaque session token and does not repeat identity creation.
- The WebSocket upgrade authenticates with an `Authorization: Bearer` header. REST requests use optional `X-Request-ID`; guest creation additionally requires `Idempotency-Key`.
- Repeating guest creation with the same idempotency key and body replays the original response. Reusing the key with a different body returns `409 invalid_request`.
- Guest creation idempotency records are retained for 24 hours; reuse after expiry is a new request.
- JSON field names are `snake_case`, IDs are opaque strings, and hidden cards, marker identity, and deck order are never projected.

## Examples

- A repeated `game.discard` request returns the original accepted result and does not create a second discard.
- A player receives their drawn card in `game.drawn`; other players receive state without that card identity.
- A reconnecting player receives their own hand and public table state, while opponent projections contain hand counts only.

## Consequences

`docs/protocol/contract.md` is the exact shape reference and `docs/protocol/examples.md` provides fixtures for contract tests. BE must validate and produce these shapes. FE must decode them without inventing authoritative fields. A breaking change requires a new protocol decision and version.
