# 0013 — Backend transport, persistence, migration, and guest session stack

## Status

Accepted

## Context

The project needs a concrete backend stack before room, REST, WebSocket, and persistence implementation begins. The game engine must remain pure and independent from these choices.

## Decision

- Gin is the HTTP framework for REST routes and the WebSocket upgrade boundary.
- `coder/websocket` manages WebSocket connections after upgrade.
- GORM with the PostgreSQL driver handles query and repository mapping.
- dbmate manages versioned SQL migrations. GORM `AutoMigrate` is not used.
- JWT is used for guest session authentication with environment-provided secret, issuer, and expiry.

## Consequences

The pure game engine cannot import Gin, WebSocket, GORM, JWT, or database packages. Transport and persistence packages depend on the engine through local interfaces. Make targets expose API run, PostgreSQL lifecycle, migration, rollback, and status commands without hiding missing implementation behind fake success.
