# Task — Reconnect snapshot and stale-event resync

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement pure MVP resynchronization behavior: a reconnect or gap requests a personalized current snapshot, including replacement sequence metadata, without replaying individual events. State mutation only happens for the snapshot replacement, never for stale requests.

## Acceptance criteria

- [ ] A `state.resync` request includes the client’s last received sequence.
- [ ] The response is a personalized current snapshot rather than an event replay.
- [ ] The response includes the snapshot, latest server sequence, replaced-since sequence, and `mode: "snapshot"`.
- [ ] The client’s last received sequence at or below the latest sequence is accepted as stale-handling input.
- [ ] A gap request does not mutate room, game, hand, or locked-table state.
- [ ] The snapshot is valid for the requesting recipient only.
- [ ] Reconnect returns the same personalized projection used for live state.
- [ ] Table-driven tests cover first reconnect, stale sequence, gap, equal sequence, and projection equivalence.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/realtime/...`

## Out of scope

- HTTP/WebSocket upgrade, connection authentication, and I/O.
- Command acceptance, action execution, and game mutation.
- Event persistence/replay and multi-instance recovery.
- REST guest identity, PostgreSQL, and Flutter integration.
- Protocol document changes.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/envelope.md`, `docs/protocol/examples.md`, `docs/protocol/visibility.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/realtime`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 reconnect and stale-event behavior.
- Previous slices: personalized projection, envelope, and recipient sequence handling.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

Implements the documented snapshot-based MVP resync behavior without changing it.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added reconnect snapshot/resync model in server/internal/realtime.
Verified:
- make be-fmt
- make be-test
Known risks:
- Snapshot resets recipient stream state only when a stream is provided.
Next task:
