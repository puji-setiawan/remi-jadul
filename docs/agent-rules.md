# Agent Rules

This document contains the project-specific quality bar referenced by `opencode.json`.

## Plan before edits

Before changing files, the agent must identify:

- the user-visible behavior;
- the current source of truth;
- acceptance criteria;
- interfaces or data contracts affected;
- tests required;
- risks and unresolved assumptions.

For a task that changes game behavior, the plan must include concrete card examples and explain how the server validates the action.

## Implement narrowly

Make focused changes. Do not opportunistically reorganize unrelated code, replace working libraries, or add speculative extension points. If a prerequisite is missing, document it as a blocker or create the smallest prerequisite that is independently useful.

## Verify behavior, not appearance

Tests must exercise the behavior that matters. Rules tests should cover valid inputs, invalid inputs, boundaries, and state transitions. Client tests must not replace server-rule tests. Protocol tests must check both serialization shape and visibility of private information.

## Stop conditions

Stop and ask for a product decision when:

- two rules in `docs/GAME_RULES.md` produce different outcomes;
- a required protocol field or error behavior is unspecified;
- a security or privacy decision cannot be inferred safely;
- implementing a request would make the client authoritative;
- the requested behavior conflicts with an existing recorded decision.

Do not select a convenient default silently.
