# Remi Jadul Test Matrix

This matrix is the minimum verification plan before the BE, FE, and PM agents coordinate an implementation slice. Server rules remain the source of truth; client tests must not replace server-rule tests.

## Phase 1 pure game engine

| Area | Valid cases | Invalid cases | Boundary cases | Owner |
|---|---|---|---|---|
| Card identity | 52 distinct physical IDs | Duplicate ID rejected | First and last generated card | BE |
| Deal | 2, 3, and 4 player deals | 0, 1, and 5 players rejected | 11, 9, and 7 starting cards | BE |
| Marker and joker | One marker excluded, matching rank becomes active joker | Marker card used in play | Marker is the last remaining card | BE |
| Seri | Same-suit consecutive cards and valid joker representation | Wrong suit, gap without joker, duplicate physical card | Minimum 3 cards and one-card extension | BE |
| Paralel | Same rank with different suits | Duplicate suit, wrong rank, reused physical card | Minimum 3 cards and maximum natural suits | BE |
| Table locking | Valid laydown becomes locked | Move, split, merge, or reuse locked card | First and second locked combinations | BE |
| Scoring | Highest valid partition and close value | Invalid partition counted as valid | Equal-score deterministic partition and tie rank | BE |

## Phase 1B state engine

- Draw only on the current player's turn.
- Exactly one discard ends a normal turn.
- Discard pickup validates the eligible window, jago basis, transferred suffix, and atomic rollback.
- Closed-deck exhaustion ends the round with the defined reason.
- Active-joker discard ends the round with `-50`.
- Close removes exactly one closing card and validates the remaining partition atomically.
- A rejected action does not mutate hands, table, discard pile, turn, or score.

## Protocol and visibility

- Every message uses the versioned envelope.
- Command request IDs correlate with responses and errors.
- Server events have monotonically ordered sequence numbers.
- Own hand is visible only to its owner.
- Opponent hand count is public; opponent card identity is absent.
- Complete discard pile, active joker rank, locked table combinations, and completed-round scoreboard are public.
- Marker card and closed-deck order are absent from every projection, error, log, and fixture.
- Reconnect detects stale sequences and returns a personalized snapshot or replay.

## Flutter client

- Lobby and room states render loading, empty, error, and success states.
- Gameplay renders 2, 3, and 4 player table layouts.
- Tap deck requests draw; accepted draw animates into the hand.
- Dragging a hand card to discard requests discard; rejected discard returns the card.
- Selecting two hand cards and one eligible discard sends one pickup intent.
- Hand reorder is presentation-only.
- No combination suggestions are shown.
- Public table cards and full discard history are visible.
- Reconnect does not expose hidden cards or lose accepted state.

## Standard commands

```text
make be-test
make be-fmt
make fe-test
make fe-analyze
make check
```

The Android emulator acceptance flow is a local FE check. CI must run Go, protocol, and Flutter checks without requiring an emulator.
