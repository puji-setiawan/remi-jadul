# 0004 — Three-round cumulative match

## Status

Accepted

## Context

The MVP needs a defined match boundary and cumulative scoring instead of ending the product flow after one round.

## Decision

A match consists of exactly three rounds. Each round is scored independently, and each player's match score is the sum of their completed-round scores. The player with the highest cumulative score wins. Final winner and ranking are shown only after round three.

## Examples

- A player scoring `24`, `-10`, and `8` finishes the match with `22`.
- A game-over round contributes its defined score to the cumulative total.

## Consequences

Match state must track the current round number and cumulative scores. Round results can be recorded after each round, but winner and final ranking must not be presented until the third round is complete. Tie handling uses the existing equal-score rule: tied final scores share the same rank.
