# 0001 — MVP baseline

## Status

Accepted

## Context

The project needs a stable baseline before autonomous implementation begins.

## Decision

- Use a Flutter Android client and Go backend in one repository.
- Keep the game engine pure and server-authoritative.
- Use PostgreSQL for identity and completed results.
- Keep active room state in memory for MVP; no Redis or multi-instance recovery.
- Use guest identity with nickname.
- Use REST for ordinary resources and versioned explicit JSON over WebSocket for active games.
- Resolve unclear game rules through additional decision records instead of agent guesses.

## Consequences

The first coding phase should focus on a deterministic pure Go engine. Client and transport work must consume that engine rather than duplicate its rules.
