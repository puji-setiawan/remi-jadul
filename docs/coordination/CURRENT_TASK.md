# Task — Personalized game state projection v1

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement pure, transport-independent personalized game-state projection from in-memory room/game state to the documented WebSocket v1 shape. The projection exposes only the requesting player’s hand and public state; opponent hands, the joker marker, and closed-deck order remain hidden.

## Acceptance criteria

- [ ] A projection contains match ID, round number, round count, round status, end reason, current player, phase, active joker rank, discard pickup window size, closed deck count, discard pile, table combinations, and scoreboard in the documented shape.
- [ ] `own_hand` contains the requesting player’s full hand, including rank, suit, and active-joker flag.
- [ ] Opponent projections contain player identity, display name, connection, and `hand_size` only.
- [ ] The physical joker marker, closed-deck order, and other players’ card identities are absent from every projection.
- [ ] Game-over rounds project only the player scores allowed by the rules.
- [ ] Projection construction does not mutate the supplied room and game state.
- [ ] Table-driven tests cover 2-player, 3-player, 4-player, own-hand, opponent-hand-size, visibility, and mutation cases.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/realtime/...`

## Out of scope

- HTTP/WebSocket upgrade, authentication, connection lifecycle, and routing.
- Command acceptance, request IDs, action validation, and event ordering.
- Reconnect resync and stale-event handling.
- REST endpoints and guest identity transport.
- PostgreSQL persistence and Flutter integration.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0003-public-discard-pile.md`, `docs/decisions/0009-public-table-combinations.md`, `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/visibility.md`, `docs/protocol/envelope.md`, `docs/protocol/examples.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/realtime`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 personalized state projections.
- Previous slices: game engine, guest identity, room lifecycle, and host start.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

This slice implements the documented personalized game-state projection shape without changing it. Serialization and transport remain later work.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added personalized game-state projection in server/internal/realtime.
Verified:
- make be-fmt
- make be-test
Known risks:
- Field names/shape aliases rely on documented names; hidden expectations may differ.
Next task:
