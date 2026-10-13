# Development Prerequisites

## Required tools

- Git.
- Go version declared in `server/go.mod`.
- Flutter SDK compatible with `app/pubspec.yaml`.
- Android SDK and an emulator for local acceptance testing.
- Docker and Docker Compose for PostgreSQL and future local services.

Check the local toolchain before assigning an implementation task:

```bash
go version
flutter --version
dart --version
docker compose version
```

The repository does not pin a Flutter channel or Android API level yet. Record those choices in the task packet before starting FE implementation.

## Verification entrypoints

```bash
make be-test
make be-fmt
make fe-test
make fe-analyze
make check
```

Run the focused owner command first. Run `make check` before handoff. If a required tool is missing, report the exact command and leave the task incomplete.

## Agent startup order

1. PM confirms `docs/coordination/CURRENT_TASK.md` has one owner and one observable behavior.
2. BE implements the active backend packet and updates its handoff result.
3. PM or review verifies the BE acceptance criteria.
4. PM creates the next packet. FE does not infer an undocumented protocol shape.
5. FE starts after the relevant protocol packet and visibility fixtures are accepted.

Only one task may be `IN_PROGRESS` at a time. Cross-package work requires explicit allowed files and an agreed interface in the task packet.
