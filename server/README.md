# Server

Go backend boundary for the Remi game. Implementation starts in Phase 1 of [`../docs/ROADMAP.md`](../docs/ROADMAP.md).

The first code added here must be the pure game engine and its tests. Do not add transport or database code before the engine contract is clear.

Backend agent ownership:

- Read the active packet in [`../docs/coordination/CURRENT_TASK.md`](../docs/coordination/CURRENT_TASK.md).
- Keep rules in pure Go packages before adding transport or persistence.
- Verify with `make be-test` and `make be-fmt`.
- Update the handoff with changed files, commands, results, and risks.
