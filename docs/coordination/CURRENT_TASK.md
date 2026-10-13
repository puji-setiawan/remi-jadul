# Task — Pure Go deterministic scoring partition selection

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement pure, transport-independent scoring for a hand: choose the valid combination partition with the highest total score, apply deterministic tie-breaking, and add closing-card value.

## Acceptance criteria

- [x] Combination cards use rank values: Ace +3, 2-10 +1, J/Q/K +2.
- [x] Uncombined cards use negative base values: Ace -3, 2-10 -1, J/Q/K -2.
- [x] Active jokers use the value of their represented card.
- [x] The solver selects the highest-scoring disjoint valid partition.
- [x] Equal-scoring partitions use a deterministic physical-card-ID comparator.
- [x] Closing score adds Ace +15, 2-10 +5, J/Q/K +10.
- [x] Card score and partition helpers are available as public aliases.
- [x] Table-driven tests cover combinations, uncombined cards, jokers, close value, and helpers.
- [x] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/game/...`

## Out of scope

- Full round scoring for all players
- Locked table combination integration into hand ownership
- Close validation and game-over scoring
- Room, protocol, persistence, and client work

## Dependencies and decisions

- `docs/GAME_RULES.md` section 9
- decision `0002`
- Completed combination and table-locking slices

## Verification

```text
make be-test
make be-fmt
```

## Handoff result

Changed:
- Added scoring solver in `server/internal/game/scoring.go`.
- Added scoring helper aliases and score value functions.
- Added table-driven tests in `server/internal/game/scoring_test.go`.

Verified:
- `make be-fmt`
- `make be-test`
- `go vet ./...`
- `go test -race ./...`

Known risks:
- Partition tie comparison uses a documented deterministic card-ID representation, but exact hidden comparator expected may differ.
- Joker representation is optimized for score within each combination; integration with server-owned combination declarations remains a later task.

Next task: close validation, game-over scoring, or jago pickup validation.
