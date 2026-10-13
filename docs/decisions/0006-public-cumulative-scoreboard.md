# 0006 — Public cumulative scoreboard

## Status

Accepted

## Context

Players need match progress information while the three-round match is in progress, without receiving hidden card state or an early final ranking.

## Decision

Public match information includes the current round and each player's cumulative score from completed rounds. The scoreboard is updated after each round. Final winner and ranking are shown only after round three.

## Examples

- After round one, the scoreboard can show a player's cumulative score of `24` while the match remains in progress.
- During an active round, the scoreboard does not reveal another player's hand or uncommitted current-round score.

## Consequences

Personalized state projections must expose the same public scoreboard to all players while preserving hand and deck secrecy. Contract tests must verify scoreboard visibility and the absence of hidden card data.
