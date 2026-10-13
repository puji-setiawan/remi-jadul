# Task — Card and deck foundation

Status: READY
Owner: BE
Priority: P0

## Goal

Implement the pure game-engine card and deck foundation: physical card identity, standard deck construction, server-side shuffle input, player-count deal sizes, and joker-rank marker selection. The package must remain transport-independent.

## Acceptance criteria

- [ ] Every card in a standard 52-card deck has a unique physical ID within a round.
- [ ] Deal sizes are 11, 9, and 7 cards for 2, 3, and 4 players.
- [ ] One remaining card is removed as the hidden marker and its rank is exposed only through the active joker rank.
- [ ] Invalid player counts are rejected without creating a partial deal.
- [ ] Table-driven tests cover valid counts, invalid counts, unique IDs, and marker exclusion.

## Allowed files/packages

- `server/internal/game/...`

## Out of scope

- Combination validation, jago pickup, turn lifecycle, scoring, rooms, transport, persistence, Flutter, and protocol payload implementation.

## Dependencies and decisions

- Rules/decision records: `docs/GAME_RULES.md`, `docs/decisions/0001-mvp-baseline.md`.
- Protocol dependencies: none for this pure engine slice.
- Blockers: Go toolchain must be available; no product decision is currently blocking this slice.

## Verification

Expected commands:

```text
make be-test
make be-fmt
```

## Handoff result

Changed:
Verified:
Known risks:
Next task:
