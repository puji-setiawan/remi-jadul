# 0007 — No combination hints

## Status

Accepted

## Context

The game should rely on the player's ability to identify and choose combinations rather than receiving automatic suggestions from the client or server.

## Decision

The client must not display valid-combination suggestions or recommendations. The player selects cards and intended combinations; the server validates the submitted action and returns either an accepted state change or a rejection.

## Examples

- After a draw, the client displays the hand but does not highlight possible seri or paralel combinations.
- An invalid selected combination is rejected by the server without mutating state.

## Consequences

Recommendation UI and recommendation protocol payloads are out of MVP scope. Server-side validation remains mandatory and authoritative.
