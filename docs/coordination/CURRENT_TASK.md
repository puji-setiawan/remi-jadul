# Task — REST guest identity endpoint v1

Status: DONE
Owner: BE
Priority: P0

## Goal

Implement the documented `POST /api/v1/guests` endpoint using the existing in-memory guest identity store. The endpoint validates a nickname, enforces idempotency by key, issues a JWT with issuer and expiry, and returns only the safe response fields.

## Acceptance criteria

- [ ] `POST /api/v1/guests` accepts `{ "display_name": "..." }` under `/api/v1`.
- [ ] The endpoint returns `201` with `version`, `request_id`, `guest_id`, `display_name`, and `access_token` for a valid nickname.
- [ ] Blank or whitespace-only nicknames return a safe validation error without creating a guest.
- [ ] `Idempotency-Key` is required for guest creation; missing key returns a safe error.
- [ ] Replaying the same idempotency key with the same body returns the original response and status.
- [ ] Reusing an idempotency key with a different body returns HTTP `409` with `invalid_request`.
- [ ] `X-Request-ID` is echoed when supplied; the server generates one when absent.
- [ ] JWT claims include issuer and expiry, and the token is not logged or placed in a URL.
- [ ] Errors return safe REST error payloads without guest internals or hidden game data.
- [ ] HTTP handler tests cover valid creation, blank nickname, missing key, idempotency replay, key reuse conflict, and request correlation.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/cmd/api/...`
- `server/internal/transport/...`
- `server/internal/auth/...`

## Out of scope

- WebSocket upgrade and active game command handling.
- Room creation, discovery, membership, and game start over REST.
- PostgreSQL persistence and migrations.
- JWT validation middleware beyond token issuance.
- Protocol envelope/event ordering and reconnect over WebSocket.
- Flutter integration.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0013-backend-stack.md`.
- Protocol docs: `docs/protocol/contract.md`, `docs/protocol/envelope.md`, `docs/protocol/errors.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules and REST boundary.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 3 REST boundary and Phase 4 identity direction.
- Previous slices: guest identity, versioned envelope, and error model.
- Blockers: JWT signing secret must come from environment configuration, never source code.

## Protocol impact

Implements the documented REST guest creation contract without changing it.

## Verification

Expected commands:

```text
make be-fmt
make be-test
```

## Handoff result

Changed:
- Added REST guest identity endpoint and JWT/idempotency handling in server.
Verified:
- make be-fmt
- make be-test
Known risks:
- External JWT library not used; manual HMAC JWT generated.
Next task:
