# Task — Guest identity with nickname

Status: READY
Owner: BE
Priority: P0

## Goal

Implement pure in-memory guest identity creation and lookup for the Android client bootstrap. Guests are identified by a server-generated ID and carry a trimmed non-empty display name. No transport, JWT, persistence, or room behavior is included in this slice.

## Acceptance criteria

- [ ] A guest can be created with a trimmed non-empty nickname.
- [ ] Blank/whitespace-only nicknames are rejected without partial state.
- [ ] Each created guest has a unique opaque server-generated ID.
- [ ] Created guests can be looked up by ID.
- [ ] Looking up an unknown guest returns a typed not-found error.
- [ ] Creation does not require database, network, clock-dependent production behavior, or transport.
- [ ] Table-driven tests cover valid nickname, blank nickname, unique IDs, and unknown lookup.
- [ ] `make be-fmt` and `make be-test` pass.

## Allowed files/packages

- `server/internal/auth/...`

## Out of scope

- REST endpoint and request/response handling.
- JWT generation, authentication middleware, and session validation.
- Room lifecycle, host start, and multi-player state.
- PostgreSQL repositories and migrations.
- Protocol payload serialization and Flutter integration.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`.
- Architecture: `docs/ARCHITECTURE.md` backend modules (`internal/auth`).
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 2 guest nickname identity.
- Blockers: No product decision is currently blocking this slice.

## Protocol impact

None. Guest identity wire shape and authentication remain Phase 3 work.

## Verification

Expected commands:

```text
make be-test
make be-fmt
```

## Handoff result

Changed:
Verified:
Known risks:
Next task:
