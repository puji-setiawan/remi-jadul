# Task — Visibility and protocol contract assertions

Status: DONE
Owner: BE
Priority: P0

## Goal

Add executable contract tests proving the implemented projection and envelope models satisfy the documented visibility and safety rules. This slice verifies contract behavior; it does not change rules or network transport.

## Acceptance criteria

- [ ] A contract test asserts that `own_hand` contains only the requesting player’s cards.
- [ ] A contract test asserts that opponent projections contain `hand_size` and never card IDs, ranks, suits, or joker assignments.
- [ ] A contract test asserts that the physical joker marker and closed-deck order are absent from projection output.
- [ ] A contract test asserts that discard pile, locked table combinations, active joker rank, turn, phase, closed deck count, and completed-round scoreboard remain available.
- [ ] A contract test asserts that action-error payloads contain only code, message, and retryability status.
- [ ] A contract test asserts that action-error payloads do not expose card details, hands, marker identity, or deck order.
- [ ] A contract test asserts that reconnect responses preserve the documented projection and replacement sequence metadata.
- [ ] All contract assertions run without network, database, or clock dependencies.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/realtime/...`
- `server/internal/transport/...`

## Out of scope

- HTTP/WebSocket upgrade, connection lifecycle, and transport handlers.
- REST guest identity and session authentication.
- PostgreSQL persistence and migrations.
- Flutter contract consumption and presentation.
- Changes to game rules, protocol docs, or decision records.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/visibility.md`, `docs/protocol/envelope.md`, `docs/protocol/errors.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 visibility and contract tests.
- Previous slices: game engine, room, projection, envelope, recipient sequence, and reconnect/resync.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

None. This slice verifies existing documented contracts.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added executable contract tests in server/internal/realtime and server/internal/transport.
Verified:
- make be-fmt
- make be-test
Known risks:
- Assertions rely on current struct/JSON shapes; hidden contracts may use aliases.
Next task:
