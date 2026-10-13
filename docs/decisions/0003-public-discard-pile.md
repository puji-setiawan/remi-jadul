# 0003 — Public discard pile history

## Status

Accepted

## Context

Players need to see which cards have already been discarded, while discard pickup must remain limited to the current top card.

## Decision

The server exposes the complete discard pile to every player in discard order. The last card is the current top discard and is the only card eligible for pickup.

## Examples

- If `A♠` is discarded and then `9♣` is discarded, all players see `[A♠, 9♣]`; only `9♣` can be picked up.
- A rejected pickup leaves the complete discard pile unchanged.

## Consequences

Personalized public-state projections and their contract tests must include the complete discard pile without exposing hidden hands, the joker marker, or deck order.
