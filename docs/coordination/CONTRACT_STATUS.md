# Contract Status

## Current vertical slice

Status: DOCUMENTED, CONTRACT TESTS NOT IMPLEMENTED

The exact v1 JSON shapes, REST boundary, retry semantics, visibility rules, and review fixtures are documented. Before the first shared slice starts, BE and FE must add executable contract fixtures and agree on the serialized implementation types.

## Rules

- BE owns authoritative command validation and event production.
- FE consumes documented projections and sends intent only.
- PM accepts a contract change only when both sides’ affected tests and examples are updated.
- A breaking contract change requires a protocol version decision record.

## Required evidence before integration

- Example JSON for each command/event.
- Error examples for invalid turn, stale request, unknown card, and hidden-data violation.
- Visibility matrix for own hand, opponent hand count, discard pile, table combinations, joker rank, scoreboard, and deck order.
- Per-recipient sequence/reconnect behavior and the snapshot-based MVP resync semantics.
