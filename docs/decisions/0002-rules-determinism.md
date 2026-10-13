# 0002 — Deterministic rules and scoring behavior

## Status

Accepted

## Context

The initial game specification left several rules open. Autonomous implementation must not choose behavior based on iteration order or transport convenience.

## Decision

- Score a hand using the valid partition with the highest total score.
- If multiple partitions have the same score, choose deterministically by sorted physical card IDs and combination representation. The exact internal comparator must be covered by tests and must not depend on map iteration order.
- Equal final scores receive the same rank. Do not apply hidden seat-order or closing-player tie-breaks.
- After drawing from the closed deck, a player may discard any card. A combination is not required for a normal deck draw.
- A discard pickup must immediately create exactly one new valid combination containing the picked discard.
- A player closes by discarding exactly one closing card while every remaining hand card can be partitioned into valid combinations.
- Closing is valid on the first turn if all conditions pass.
- A game-over caused by an unresolved discard pickup gives the causing player `-50`; other players are not scored for that round.

## Consequences

The engine needs deterministic partition search, explicit tie tests, atomic discard-pickup validation, and a close validator that evaluates the hand after removing the closing card. The client does not determine or suggest authoritative combinations.

## Examples

- A hand with two possible valid partitions uses the higher total, even if the lower total is found first.
- Two players at `14` both receive the same rank.
- A player drawing a useless card may discard it or another owned card, subject to the normal one-discard rule.
- A discard pickup that would require two newly laid combinations is rejected.
