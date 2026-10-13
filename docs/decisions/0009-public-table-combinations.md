# 0009 — Public table combinations

## Status

Accepted

## Context

Players need to see all cards already committed to the table when estimating which cards may remain available in the closed deck. Table cards must not become hidden based on which player laid them down.

## Decision

Every locked seri and paralel from every player is public and displayed in the shared table information. The closed-deck order and unexposed hands remain hidden.

## Examples

- A seri laid by one player is visible to all players.
- A player may count visible table cards and the discard pile when estimating draw possibilities, but cannot know the closed-deck order.

## Consequences

Personalized state projections must include all locked table combinations while continuing to hide private hands and the deck. Visibility tests must verify both players receive the same public table combinations.
