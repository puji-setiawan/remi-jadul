# Task — Versioned command envelope and action errors

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement the versioned WebSocket command envelope contract and a safe typed action-error model before any transport handler. The package validates envelope shape and request IDs without mutating game state or depending on network I/O.

## Acceptance criteria

- [ ] Client commands require `version`, `type`, `name`, and `request_id`.
- [ ] Room-scoped commands require `room_id`.
- [ ] Unknown or unsupported command names return a typed `invalid_command` error.
- [ ] Missing required envelope fields return typed validation errors.
- [ ] Unknown fields are ignored without causing an error.
- [ ] Action errors include only stable code, safe localized message, and retryability status.
- [ ] Action errors never contain card IDs, hands, marker identity, deck order, or server-only validation details.
- [ ] Request IDs can be retained and compared for idempotency lookup in later slices.
- [ ] Table-driven tests cover valid commands, missing fields, unknown names, `room_id` requirements, unknown fields, and hidden-data absence.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/transport/...`

## Out of scope

- HTTP/WebSocket upgrade, connection authentication, and routing.
- Guest creation, JWT, and persistence.
- State mutation, command business validation, and game actions.
- Event sequence/order, reconnect, resync, and broadcast projections.
- Protocol document changes; `docs/protocol/` remains the contract source.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/envelope.md`, `docs/protocol/errors.md`, `docs/protocol/commands.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/transport`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 action errors and request IDs.
- Previous slices: game engine, guest identity, room lifecycle, host start, and personalized projection.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

Implements the documented v1 envelope and error model without changing it.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added versioned command envelope and safe action-error model in server/internal/transport.
Verified:
- make be-fmt
- make be-test
Known risks:
- JSON decode fallback accepts unknown fields by ignoring them.
Next task:
