# Task — Authenticated connection registry and personalized broadcast

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement an in-memory, thread-safe registry of authenticated WebSocket connections and a personalized broadcast primitive. The slice adds delivery plumbing only; it does not execute commands or mutate game/room state.

## Acceptance criteria

- [x] A connection can be registered with guest ID, recipient stream, and connection metadata.
- [x] A connection can be unregistered by guest ID or registration token.
- [x] Registering the same guest replacement does not leave duplicate active connections.
- [x] A broadcast helper delivers a personalized payload to each requested recipient.
- [x] A missing recipient is skipped without aborting delivery to other recipients.
- [x] Registry operations are safe for concurrent use.
- [x] Registry state does not expose card IDs, hand contents, marker identity, or deck order.
- [x] Table-driven tests cover register, unregister, duplicate replacement, missing recipient, and concurrent access.
- [x] `make be-fmt` and `go vet ./...` pass.

## Allowed files/packages

- `server/internal/transport/...`
- `server/internal/realtime/...`

## Out of scope

- Command routing and game-state mutation.
- Room creation, join, leave, and host start.
- Live round broadcast integration.
- Reconnect and resync transport delivery.
- Persistence, REST endpoints, and Flutter integration.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0011-protocol-v1-contract.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/envelope.md`, `docs/protocol/visibility.md`.
- Architecture: `docs/ARCHITECTURE.md` realtime and transport boundaries.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 personalized projections and reconnect UX.
- Previous slices: authenticated gateway, projection, sequence, reconnect handling, and room commands.
- Blockers: No product decision is currently blocking this slice. Thread-safety expectations are infrastructure-level, not game rule behavior.

## Protocol impact

Supports the documented per-recipient sequence and personalized delivery model without changing contract shapes.

## Verification

Expected commands:

```text
make be-fmt
make be-test
(cd server && go vet ./...)
```

## Handoff result

Changed:
- Added thread-safe authenticated connection registry in server/internal/transport.
Verified:
- make be-fmt
- make be-test
- go vet ./...
- go test -race ./...
Known risks:
- Personalized delivery uses caller-provided sender; no network broadcast integrated yet.
Next task:
