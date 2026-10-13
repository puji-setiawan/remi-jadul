# Remi Jadul — Architecture

## Repository boundaries

```text
remi-jadul/
├── app/                 # Flutter Android client
├── server/              # Go API, game engine, realtime, persistence
├── docs/                # Specifications, protocol, decisions, roadmap
├── .opencode/           # Project-local OpenCode commands and notes
├── docker-compose.yml   # Local PostgreSQL and future services
├── .env.example         # Safe local configuration template
└── README.md
```

## Backend modules

```text
server/
├── cmd/api/             # process entrypoint
├── internal/game/       # pure rules engine; no I/O dependencies
├── internal/room/       # room lifecycle and membership
├── internal/auth/       # guest identity and nickname
├── internal/realtime/   # WebSocket connections and projections
├── internal/transport/  # REST and WebSocket adapters
├── internal/persistence/# PostgreSQL repositories
└── migrations/          # versioned schema migrations
```

The exact package layout may evolve, but the game engine must remain isolated from transport and persistence.

## Backend technology decisions

- Gin handles REST routing and the HTTP upgrade boundary for realtime connections.
- `coder/websocket` handles WebSocket connections after the Gin upgrade boundary.
- GORM handles PostgreSQL query and repository mapping outside the pure game engine.
- `dbmate` owns versioned SQL migrations. GORM `AutoMigrate` is not used for schema management.
- JWT authenticates guest sessions. JWT claims identify the guest and include issuer and expiry; the token secret remains environment-only.

## Authority and data flow

```text
Flutter intent
    -> WebSocket/REST transport
    -> server validation
    -> pure game engine transition
    -> persistence/event handling
    -> personalized state projection
    -> each connected Flutter client
```

The client sends intent, not authoritative state. The server rejects invalid actions without mutating the round.

## State model

- Room state: room ID, visibility, host, capacity, members, status.
- Round state: deck, discard pile, joker rank, hands, table combinations, current player, status, end reason.
- Card state: physical card ID, rank, suit, joker interpretation when applicable.
- Projection state: public fields plus only the requesting player’s hand and private messages.

Active room state is in memory for MVP. PostgreSQL stores identity and completed results, not live round recovery.

## Protocol policy

Use explicit versioned JSON for REST and WebSocket. `docs/protocol/` is the wire-contract source of truth. Document commands, events, payloads, error codes, request IDs, sequence ordering, idempotency, and visibility before room/realtime implementation. Do not duplicate undocumented DTOs across packages.

## Local development

Docker Compose provides PostgreSQL. The API and Flutter app may run from the host during early development. CI must run Go checks and contract tests without requiring an Android emulator; emulator verification remains a local acceptance step.
