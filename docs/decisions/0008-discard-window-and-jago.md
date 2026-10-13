# 0008 — Discard window and jago pickup

## Status

Accepted

## Context

Players need to be able to use a useful card in the discard pile while limiting how far back they can search. The selected card must still be committed to a combination rather than saved for later.

## Decision

The eligible discard window is the bottom `N` cards, where `N` is the starting hand size for the player count. A player selects one eligible card and takes it plus every card above it. The selected card must immediately be laid with exactly two cards from the player's hand as the jago basis. The jago may be extended with additional transferred or held cards when the final same-suit consecutive seri or same-rank different-suit paralel remains valid. The cards above the selected card may remain in hand or be used in other legal actions on that turn.

## Examples

- In a two-player round, the bottom 11 discard cards are eligible. Selecting the ninth card takes that card and the four cards above it.
- `2♦, 2♣` may take eligible discard `2♥` and be laid as a jago paralel.
- `4♥, 5♥` may take eligible discard `6♥`; if transferred `3♥` is available, all four may be laid as the extended seri `3♥, 4♥, 5♥, 6♥`.
- A failed jago validation leaves the discard pile and hand unchanged.

## Consequences

Pickup validation must calculate the eligible window from the player count, transfer the selected suffix atomically, and validate the immediate jago before mutation. Hidden cards and discard cards outside the eligible window remain protected from pickup.
