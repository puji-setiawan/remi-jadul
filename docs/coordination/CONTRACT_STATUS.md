# Contract Status

## Current vertical slice

Status: NOT STARTED

The first shared slice is room identity and state projection. Before it starts, BE and FE must agree on the exact JSON shapes in `docs/protocol/` and add contract fixtures.

## Rules

- BE owns authoritative command validation and event production.
- FE consumes documented projections and sends intent only.
- PM accepts a contract change only when both sides’ affected tests and examples are updated.
- A breaking contract change requires a protocol version decision record.

## Required evidence before integration

- Example JSON for each command/event.
- Error examples for invalid turn, stale request, unknown card, and hidden-data violation.
- Visibility matrix for own hand, opponent hand count, discard pile, table combinations, joker rank, scoreboard, and deck order.
- Sequence/reconnect behavior.
