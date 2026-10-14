# Task — Persistence integration tests

Status: READY
Owner: BE
Priority: P1

## Goal

Add database-backed integration tests for the PostgreSQL guest and completed match repositories. The tests verify migration-backed round trips without changing application runtime behavior.

## Acceptance criteria

- [ ] Guest repository integration tests create, load, and reject unknown guest IDs.
- [ ] Completed match repository integration tests save and load a match, players, and round scores.
- [ ] Game-over result tests store and retrieve a not-calculated round score as nullable.
- [ ] Integration tests are skipped or marked clearly unavailable when PostgreSQL is not configured.
- [ ] Successful tests verify both create/load results.
- [ ] Repository changes are not made solely to make tests pass.
- [ ] `make be-fmt`, `go vet ./...`, and `go test ./...` pass when the database is available.
- [ ] Race tests do not require a database when skipping integration tests.

## Allowed files/packages

- `server/internal/persistence/...`

## Out of scope

- Runtime persistence wiring into active room/game commands.
- REST and WebSocket endpoint changes.
- Match lifecycle or scoring rule changes.
- Flutter integration.
- Migration updates.
- Changes to documented rules, decisions, or roadmaps.

## Dependencies and decisions

- Rules/decision records: `docs/decisions/0001-mvp-baseline.md`, `docs/decisions/0013-backend-stack.md`.
- Architecture: `docs/ARCHITECTURE.md` PostgreSQL, migration, and repository boundaries.
- Schemas: guest and completed result migrations.
- Product scope: `docs/PRODUCT_BACKLOG.md` Phase 4 persistence integration tests.
- Previous slices: guest repository, completed result repository, and result migration.
- Blockers: A running PostgreSQL instance and valid `DATABASE_URL` are required; otherwise tests skip integration assertions.

## Protocol impact

None.

## Verification

Expected commands:

```text
make db-migrate
go test ./...
go test -race ./...
```

## Handoff result

Changed:
Verified:
Known risks:
Next task:
