# Protocol Contract

This directory is the source of truth for client/server wire contracts. It becomes mandatory before room/realtime implementation begins.

## Ownership

- Go defines and validates the authoritative command/event behavior.
- Flutter consumes the documented contract and sends player intent only.
- A protocol change requires updating this directory and both sides’ contract tests in one change.

## Envelope requirements

Every WebSocket message must have a versioned envelope with:

- protocol version;
- message type;
- request ID for client commands and correlated responses/errors;
- room ID where the message belongs to a room;
- sequence number for ordered server events;
- typed payload or typed error.

The exact JSON names are to be fixed in the first protocol decision record before Phase 3 implementation. Do not duplicate undocumented JSON maps in Flutter and Go.

## Contract rules

- Commands express player intent; clients never submit authoritative state.
- Server errors are stable, machine-readable, and do not leak hidden cards or deck order.
- Rejected commands do not mutate game state.
- State projections are personalized per player.
- A player's hand display order is client-controlled presentation state; card ownership and physical card identity remain server-authoritative.
- Public game state includes the complete discard pile in discard order; pickup eligibility is governed by the discard window and jago rules.
- Public match state includes the current round and cumulative scoreboard; it must not include hidden hands or deck order.
- Public game state includes every locked table combination from every player, so clients can display all public cards when estimating draw possibilities.
- Events are ordered and reconnect can identify stale or missing sequence numbers.
- REST and WebSocket contracts use explicit versioning.

## Required contract groups

- Guest identity and nickname.
- Create/list/join/leave room.
- Start game.
- Draw closed deck.
- Take an eligible discard and the cards above it with exactly one immediate jago combination.
- Lay down combination.
- Discard.
- Close round.
- Personalized room/game state.
- Round result and ranking.
- Reconnect and resynchronization.
