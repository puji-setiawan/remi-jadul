# Remi Jadul — Agent Instructions

## Mission

Build Remi Jadul as a server-authoritative online card game for Android. The Go backend owns game state, rules, validation, dealing, turns, scoring, and round completion. The Flutter client renders state and collects player intent; it must never decide authoritative outcomes.

## Required workflow

For every task, follow this sequence:

1. **Plan** — inspect the relevant files, state the intended behavior, acceptance criteria, affected interfaces, and risks.
2. **Implement** — make the smallest coherent change that satisfies the plan. Preserve unrelated user work.
3. **Verify** — run the narrowest relevant tests first, then formatter/linter/build checks. Report exact commands and results.

Do not claim completion without verification evidence. If verification cannot run, explain why and mark the task incomplete.

## Source of truth

- `docs/PRODUCT_BACKLOG.md` is the current product backlog and scope baseline.
- `docs/GAME_RULES.md` is the sole game-rules authority.
- `docs/decisions/` records decisions that add, clarify, or change behavior.
- Do not silently resolve an ambiguous or conflicting game rule. Stop, describe the concrete scenario, and request a decision before changing the game engine.
- When behavior changes, update the relevant specification or decision record in the same change.

Read context progressively to conserve tokens:

1. Always read this file and the task-relevant section of `docs/ROADMAP.md`.
2. Read only the domain document needed for the task: `SPEC.md`, `PRODUCT_BACKLOG.md`, `ARCHITECTURE.md`, `GAME_RULES.md`, protocol docs, or a decision record.
3. Search the relevant package before opening broad file lists.
4. Do not paste or restate whole specifications in plans or completion reports.
5. For Flutter/UI work, read `DESIGN.md`; do not load it for backend-only work.

## Engineering rules

- Keep the game engine pure and independently testable. It must not require a database, network, WebSocket, clock, or Flutter runtime to validate rules.
- Validate every player action on the server. Treat all client input and client-reported state as untrusted.
- Give every physical card a unique identity for the duration of a round.
- Never expose hidden opponent cards through REST responses, WebSocket events, logs, errors, fixtures, or debug endpoints.
- Keep REST and WebSocket contracts explicit, versioned, and documented before client/server integration.
- Prefer simple, local abstractions. Do not introduce a framework, dependency, service, or generalized abstraction without a concrete current use case.
- Preserve existing behavior unless the task explicitly changes it.
- Never commit secrets, credentials, local machine paths, generated build output, or environment-specific configuration.

## Anti-slop guardrails

The following are defects, not acceptable shortcuts:

- Placeholder implementations presented as finished features.
- Empty error handlers, silent fallbacks, ignored errors, or fake success responses.
- Tests that only assert that code does not crash, or mocks that bypass the behavior under test.
- TODO comments without a linked task, decision, or explicit follow-up.
- Duplicated protocol types or undocumented JSON shapes between Flutter and Go.
- UI-only implementations for server-owned rules.
- Broad refactors mixed into a focused feature.
- Unnecessary comments that repeat the code instead of explaining an invariant or decision.

For each feature, include unhappy-path tests and at least one boundary case. For realtime behavior, cover invalid actions, turn ordering, event ordering, reconnect behavior, and state visibility.

## Token-efficient workflow

- One task should target one behavior or one vertical slice.
- Prefer `rg` with a narrow path over reading the entire repository.
- Use the smallest relevant test command before the full suite.
- Do not spawn a subagent for a trivial edit; use review only for cross-package, protocol, security, or rules changes.
- Keep plans to: goal, files, behavior, tests, risks.
- Keep completion reports to: changed behavior, checks run, result, known risk.
- Reuse existing commands and fixtures instead of generating duplicate examples.
- If the task is blocked by a product decision, stop early and report the exact decision needed.
- For UI or UI review work, follow `DESIGN.md` and load only the relevant anti-slop skills.
- For backend-only work, do not load UI, accessibility, mobile-layout, or copywriting guidance.

## Completion report

End every implementation task with:

- Summary of behavior changed.
- Files changed.
- Tests, formatters, linters, or builds run and their results.
- Known limitations, follow-up work, and any unresolved risk.

## Repository map

The intended future layout is:

- `app/` — Flutter Android client.
- `server/` — Go backend and pure game engine.
- `docs/` — specifications, protocol, decisions, testing, and development guides.
- `docs/coordination/` — PM task packets and BE/FE handoffs.
- `.opencode/` — project-local commands, skills, and OpenCode notes.
- `infra/` — Docker Compose and local infrastructure.
- `scripts/` — reproducible validation commands.
- `Makefile` — stable BE/FE verification entrypoints.

Keep application code out of documentation-only changes. Create `server/` and `app/` when their implementation phase starts.

<!-- antislop:start -->
## antislop
For UI, copy, people, mobile layout, or code comments work, read these installed skill files directly (use these paths even if a same-named global skill exists):
- Core filter for UI/review tasks: `antislop`: `.opencode/skills/antislop/SKILL.md`
- UI / visual: `antislop-ui`: `.opencode/skills/antislop-ui/SKILL.md`
- People: `antislop-human`: `.opencode/skills/antislop-human/SKILL.md`
- Mobile / responsive: `antislop-layoutmobile`: `.opencode/skills/antislop-layoutmobile/SKILL.md`
- Code comments: `antislop-code`: `.opencode/skills/antislop-code/SKILL.md`
For UI or review tasks, follow the core's "Two Usage Modes" section in strict order: explicit session instruction first, then global preference, then ask. A session instruction always wins. For a resolved mode, say `antislop active: <mode> (session override).` or `antislop active: <mode> (global preference).` once before presenting findings or making edits, using the actual mode and source. Acknowledging the user's request without naming the source does not replace this notice.
Only an explicit choice of antislop during or after selects a session mode. A request to review, audit, or avoid file edits does not select a mode; read the global preference in that case. Another skill's mode does not select antislop's mode.
If the mode is unresolved for a UI or review task, ask during/after and end the response; wait for the answer before any UI review, planning, or concept. For read-only tasks, put the active-mode notice only at the start of the final answer, never in progress messages. For editing tasks, announce before the first edit and omit it from the final answer. Backend-only tasks do not load these skills.
To update antislop later: `npx antislop-ai --update`, or run `npx antislop-ai` and pick Overwrite them.
<!-- antislop:end -->
