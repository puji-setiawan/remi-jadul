# 0005 — Client-controlled hand display order

## Status

Accepted

## Context

Cards are dealt randomly, but players need to arrange their own hand for readability and play. Display order must not be confused with server-authoritative card state.

## Decision

The server deals cards from the shuffled deck and owns card identity and ownership. A player's hand may be displayed in any client-selected order. Reordering a hand is presentation-only and does not mutate the game state.

## Examples

- A player may arrange `A♥, 2♥, 3♥` before `5♠` without changing the hand.
- A client cannot use display order to claim a card it does not own or change a card's identity.

## Consequences

Hand-order commands, if supported by the client, do not require game-rule validation beyond ownership of the displayed cards. Server projections and actions must continue to use physical card IDs rather than hand positions.
