# Remi Jadul — Product Specification

## Status

MVP product specification. Product backlog lives in [`PRODUCT_BACKLOG.md`](PRODUCT_BACKLOG.md); detailed game rules live only in [`GAME_RULES.md`](GAME_RULES.md); architecture and boundaries live in [`ARCHITECTURE.md`](ARCHITECTURE.md).

## Product goal

An Android realtime Remi card game for 2–4 players. Players create or join public/private rooms, play a server-validated round, reconnect to an active room, and see the final score and ranking.

## MVP capabilities

- Guest identity with nickname.
- Public and private rooms with 2–4 player capacity.
- Host-controlled game start when the configured player count is available.
- Server shuffle, dealing, joker selection, turn order, action validation, scoring, and round completion.
- Draw from closed deck or top discard with the special discard rules in `GAME_RULES.md`.
- Hidden opponent hands and reconnect to an in-memory active room.
- REST for identity, room discovery, profile, and history; WebSocket for active game state and actions.
- PostgreSQL persistence for guest/profile data and final results.

This document defines product behavior at a high level. It must not duplicate combination, turn, close, or scoring rules from `GAME_RULES.md`.

## Explicit non-goals for MVP

- Redis or multi-instance room state.
- Persistent recovery of an active round after server restart.
- OAuth, payments, chat, spectators, tournaments, and matchmaking ratings.

## User flows

1. Player enters a nickname and receives a guest identity.
2. Player creates a public/private room or joins an available room/code.
3. Host starts after the required number of players joins.
4. Server deals cards and broadcasts each player’s private hand plus public state.
5. Current player draws, handles any discard-pickup requirement, and ends a normal turn by discarding exactly one card.
6. Server broadcasts a personalized state projection after every accepted action.
7. Round ends on close, game over, or closed-deck exhaustion; server calculates ranking and persists the result.

## Success criteria

- A malicious or buggy client cannot make an invalid move or learn another player’s hand.
- The same engine tests determine the same outcome regardless of transport or UI.
- A player can reconnect to an active room and receive the correct personalized state.
- Every rules-critical behavior has deterministic tests and concrete examples.
