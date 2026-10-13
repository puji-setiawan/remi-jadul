# Task — In-memory room lifecycle and membership

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement pure in-memory room creation, discovery, join, and leave behavior for public and private rooms. Games are not started in this slice; only waiting-room membership and availability are managed.

## Acceptance criteria

- [ ] Room capacity is accepted only for 2, 3, or 4 players.
- [ ] Public rooms are created without an access code; private rooms require an access code supplied by the caller.
- [ ] A player can join a waiting public room by room ID while slots remain.
- [ ] A player can join a private room with its exact access code; an incorrect or missing code is rejected.
- [ ] Joining a room at capacity, a non-waiting room, or the same room twice is rejected.
- [ ] Leaving removes the player from the room; an empty room is removed from the registry.
- [ ] If the host leaves, the room is closed and removed rather than silently reassigned.
- [ ] Rejected join or leave actions do not mutate room state or membership.
- [ ] Public discovery returns only public rooms with an open slot and never exposes access codes.
- [ ] Table-driven tests cover creation, join, leave, capacity rejection, invalid code, duplicate join, and no-mutation-on-rejection.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/room/...`

## Out of scope

- Guest identity, REST endpoint, WebSocket transport, and JWT handling.
- Host start and game-start validation.
- Active game rounds, scoring, and reconnect.
- PostgreSQL persistence and migrations.
- Protocol payload serialization and Flutter integration.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/room`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 2 public/private room lifecycle, 2-4 capacity, and in-memory active rooms.
- Previous slices: guest identity, card, deck, deal, combination, table locking, round, and scoring foundations.
- Blockers: No product decision is currently blocking this slice. Host replacement after host departure is intentionally not supported.

## Protocol impact

None. Internal room state and access-code visibility rules remain implementation-internal until Phase 3 protocol work.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added in-memory room lifecycle/membership in server/internal/room.
Verified:
- make be-fmt
- make be-test
Known risks:
- API names aliased but hidden names can vary.
Next task:
