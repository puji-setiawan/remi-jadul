# 0010 — Discarding an active joker ends the round

## Status

Accepted

## Context

An active joker is a critical card and must not be discarded during a turn.

## Decision

If a player discards an active joker, the round ends immediately. The discarding player receives `-50`. Other players are not scored for that round, and the end reason is `game_over`.

## Examples

- If rank 4 is the active joker rank, discarding `4♣` ends the round and gives the discarding player `-50`.
- A non-joker card remains eligible for ordinary discard after a closed-deck draw.

## Consequences

The server must identify active jokers before committing a discard and atomically end the round when one is discarded. The action must not continue to opponent turns or normal close scoring.
