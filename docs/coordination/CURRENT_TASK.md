# Task — Host start and active round initialization

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement host-controlled game start for a waiting room with 2–4 players. A valid start creates an in-memory active round through the existing pure game engine and marks the room as started. No transport, protocol, or persistence behavior is included in this slice.

## Acceptance criteria

- [ ] Only the room's host can start the game.
- [ ] Starting succeeds only when the room is waiting and has between 2 and 4 joined players.
- [ ] Starting uses the game engine's server-side deck, deal, and joker foundation; client input is never authoritative for dealing.
- [ ] A successful start creates an active in-memory round and marks the room as started/active.
- [ ] Starting an already started game is rejected.
- [ ] A non-host attempt or a room with fewer than 2 players is rejected.
- [ ] Rejected starts do not mutate room membership, host, availability, or round state.
- [ ] The active round records player count, hands, joker rank, and initial turn in a form available to later transport work.
- [ ] Table-driven tests cover host start, non-host rejection, insufficient players, already started, and no-mutation-on-rejection.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/room/...`
- `server/internal/game/...`

## Out of scope

- REST/WebSocket command handling and authenticated sessions.
- Personalized state projections, event ordering, and reconnect.
- Match result persistence and match history.
- Flutter room lobby and start button UX.
- Protocol payload serialization.
- Changes to documented game rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/room`, `internal/game`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 2 (host start, 2–4 validation, in-memory active rooms).
- Previous slices: guest identity, room lifecycle, card identities, deal, and round state.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

None. Internal room/round initialization only. Transport and visibility remain Phase 3 work.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added host start and active round initialization in server/internal/room.
Verified:
- make be-fmt
- make be-test
Known risks:
- Deck sourcing is caller-provided for purity; host/transport authorization added later.
Next task:
