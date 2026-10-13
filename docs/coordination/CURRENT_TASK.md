# Task — Recipient event sequence and stale-event handling

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement pure, transport-independent recipient-stream sequence tracking and stale-event detection for the WebSocket event contract. A recipient stream has monotonic sequence numbers; missed sequences are detected without mutating room or game state.

## Acceptance criteria

- [ ] Each recipient stream starts at a documented initial sequence.
- [ ] Each delivered event receives the next monotonic sequence number.
- [ ] A private event delivered to one recipient consumes a sequence only in that recipient’s stream.
- [ ] A detected sequence gap reports the last received sequence and latest server sequence without mutating state.
- [ ] An event at or below the last received recipient sequence is marked stale and ignored.
- [ ] Receiving all expected sequences never reports a gap.
- [ ] A new snapshot resets the recipient stream’s last received sequence to the server stream’s latest sequence.
- [ ] Sequence state is safe to use independently per room and per recipient.
- [ ] Table-driven tests cover first sequence, consecutive events, gap, stale event, per-recipient streams, and snapshot reset.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/realtime/...`

## Out of scope

- Network/WebSocket upgrade, connection authentication, and I/O.
- Command acceptance, action validation, and game-state mutation.
- Reconnect HTTP/WebSocket handling and snapshot transport.
- REST endpoints, guest identity, and persistence.
- Changes to documented protocol rules or sequence allocation.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/envelope.md`, `docs/protocol/contract.md`, `docs/protocol/errors.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/realtime`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 event sequence/order handling and reconnect/resync behavior.
- Previous slices: game engine, room lifecycle, host start, personalized projection, and versioned envelope.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

Implements the documented recipient-stream sequence and stale-event behavior without changing it.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added recipient stream sequence tracking and gap/stale/snapshot handling in server/internal/realtime.
Verified:
- make be-fmt
- make be-test
Known risks:
- Initial/gap semantics based on documented monotonic recipient streams.
Next task:
