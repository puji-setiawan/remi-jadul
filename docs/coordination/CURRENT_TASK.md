# Task — Flutter protocol v1 contract models and fixture tests

Status: READY
Owner: FE
Priority: P1

## Goal

Start Phase 5 (Flutter Android client) by implementing pure Dart protocol v1 serialization/deserialization models and executable contract fixture tests in `app/` that match `docs/protocol/contract.md` and `docs/protocol/examples.md`. This unblocks `docs/coordination/CONTRACT_STATUS.md` before building lobby, room, and gameplay screens.

## Acceptance criteria

- [ ] Dart models in `app/lib/protocol/` encode all v1 client commands (`room.create`, `room.join`, `room.leave`, `game.start`, `game.draw`, `game.take_discard`, `game.laydown`, `game.discard`, `game.close`, `state.resync`) inside the v1 WebSocket envelope (`version`, `type`, `name`, `request_id`, optional `room_id`, `payload`) without client `sequence`.
- [ ] Dart models decode all v1 server events (`room.updated`, `game.started`, `game.state`, `game.drawn`, `round.ended`, `match.updated`, `match.ended`, `state.resynced`) and errors (`game.action_rejected`, `protocol.error`) with `snake_case` JSON fields.
- [ ] Dart models decode REST v1 responses and error shapes (`POST /api/v1/guests`, `GET /api/v1/rooms`, `GET /api/v1/matches/{match_id}`).
- [ ] Unknown JSON fields are ignored without throwing, and unknown message names or invalid envelopes surface a typed client decode error without inventing fallback game state.
- [ ] Opponent projections expose `hand_size` only and never invent or require opponent card lists, marker cards, or closed-deck order.
- [ ] Round-ended decoding supports both `score_status: "calculated"` (with `round_score`) and `score_status: "not_calculated"` (with nullable/omitted `round_score`).
- [ ] Unit tests in `app/test/protocol/` verify every fixture in `docs/protocol/examples.md` and `docs/protocol/contract.md`, including unhappy-path malformed JSON, unknown event/error codes, and boundary cases (empty combinations, nullable fields, resync snapshot replacement metadata).
- [ ] `make fe-test` and `make fe-analyze` pass with zero warnings.

## Allowed files/packages

- `app/lib/protocol/...`
- `app/test/protocol/...`
- `docs/coordination/CONTRACT_STATUS.md`
- `docs/coordination/CURRENT_TASK.md`

## Out of scope

- UI screens, widgets, navigation, or pseudo-3D tabletop rendering.
- Live HTTP or WebSocket network client connection management.
- Client-side game rule validation, combination checking, or score calculation.
- Any changes to `server/` or protocol specifications.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol specification: `docs/protocol/contract.md`, `docs/protocol/envelope.md`, `docs/protocol/commands.md`, `docs/protocol/events.md`, `docs/protocol/errors.md`, `docs/protocol/examples.md`, `docs/protocol/visibility.md`.
- Coordination gate: `docs/coordination/CONTRACT_STATUS.md` requires FE executable contract fixtures before shared vertical slices.
- Blockers: None. All protocol v1 decisions are `Accepted`.

## Protocol impact

None (implements the existing v1 contract on the Flutter client side).

## Verification

Expected commands:

```text
make fe-test
make fe-analyze
```

## Handoff result

Changed:
- 
Verified:
- 
Known risks:
- 
Next task:
- Guest entry and public/private lobby screen slice (FE) or HTTP/WebSocket runtime wiring in `server/cmd/api` (BE).
