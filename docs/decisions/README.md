# Decision Log

Decision records are the durable source for product and architecture choices that are not fully captured by `docs/SPEC.md`, `docs/PRODUCT_BACKLOG.md`, or `docs/GAME_RULES.md`.

## When to add a record

Add a record when a decision:

- clarifies an ambiguous game rule;
- changes an existing behavior or public contract;
- introduces an architectural dependency or constraint;
- affects security, privacy, persistence, or compatibility.

Accepted decisions are binding implementation input. A `Proposed` decision is a blocker for behavior that depends on it.

## Format

Use one Markdown file per decision, named `NNNN-short-description.md`.

```markdown
# NNNN — Decision title

## Status

Proposed | Accepted | Superseded

## Context

What problem or ambiguity required a decision?

## Decision

What is the chosen behavior?

## Examples

Concrete inputs and expected outcomes.

## Consequences

Implementation, testing, compatibility, and follow-up impact.
```

Do not use a decision record to hide an unresolved choice. Mark it `Proposed` and stop implementation where the choice matters.
