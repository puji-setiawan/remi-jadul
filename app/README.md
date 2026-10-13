# Android app

Flutter Android client boundary for the Remi game. Implementation starts in Phase 5 of [`../docs/ROADMAP.md`](../docs/ROADMAP.md).

The client renders server projections and sends player intent. It must not implement authoritative rules, scoring, hidden-card logic, or turn decisions.

Frontend agent ownership:

- Read the active packet in [`../docs/coordination/CURRENT_TASK.md`](../docs/coordination/CURRENT_TASK.md).
- Read `DESIGN.md`, `docs/UI_PLAN.md`, and only the protocol docs needed for the task.
- Keep presentation order separate from server card identity.
- Verify with `make fe-test` and `make fe-analyze`.
- Update the handoff with changed files, commands, results, and risks.
