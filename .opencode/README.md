# Project-local OpenCode configuration

The root `opencode.json` loads durable project instructions. This directory contains short, task-oriented commands so routine sessions do not need to restate the full specification.

The root `opencode.json` intentionally does **not** load every decision record or every product document in every session. Commands load the smallest relevant context on demand.

Anti-slop is installed project-scoped in `.opencode/skills/`: core, UI, accessibility, mobile layout, and code-comment review. Copywriting is intentionally omitted until the product has user-facing marketing/content work.

To conserve context, backend/rules sessions should not load `antislop-ui`, `antislop-human`, `antislop-layoutmobile`, or `DESIGN.md`. UI and review sessions load only the relevant skill.

## Token policy

- Read `AGENTS.md` first.
- Read only the relevant document from `docs/`.
- Use `plan` for cross-package, protocol, rules, or security changes.
- Use `build` for one scoped implementation task.
- Use `review` only when a second pass materially reduces risk.
- Use `verify` after implementation; do not repeat the entire repository context.
- Keep outputs compact and include command results instead of explanations of unchanged code.
- Use `DESIGN.md` only for Flutter/UI tasks.
- Use `docs/GAME_RULES.md` only for rules-engine, scoring, or game-state tasks.
