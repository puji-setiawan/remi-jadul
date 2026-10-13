# Protocol Contract v1

This document defines the exact JSON shapes shared by the Go server and Flutter client. Commands express player intent. The server owns card identity, ownership, validation, turn, scoring, and state transitions.

## Transport

- WebSocket JSON v1 carries active room lifecycle commands, game commands, events, errors, and resynchronization.
- REST carries guest identity, public room discovery, and completed match history.
- Public room discovery is REST-only through `GET /api/v1/rooms`; there is no WebSocket `room.list` command. Room creation, join, and leave remain WebSocket commands.
- Every WebSocket message uses the envelope in `envelope.md`.
- JSON field names are `snake_case`.
- IDs are opaque strings. Clients must not infer meaning from ID format.
- Unknown fields are ignored. Unknown message names produce a typed error without mutation.

## Envelope shapes

### Client command

```json
{
  "version": "v1",
  "type": "command",
  "name": "game.draw",
  "request_id": "req_01JDRAW",
  "room_id": "room_123",
  "payload": {
    "source": "closed_deck"
  }
}
```

### Server event

```json
{
  "version": "v1",
  "type": "event",
  "name": "game.state",
  "request_id": "req_01JDRAW",
  "room_id": "room_123",
  "sequence": 42,
  "payload": {}
}
```

`request_id` is echoed when an event is the accepted result of a command. Broadcast events may omit it.

### Server error

```json
{
  "version": "v1",
  "type": "error",
  "name": "game.action_rejected",
  "request_id": "req_01JDRAW",
  "room_id": "room_123",
  "sequence": 43,
  "payload": {
    "code": "not_your_turn",
    "message": "Aksi belum tersedia pada giliran ini.",
    "retryable": false
  }
}
```

Errors never contain hidden cards, marker identity, deck order, or server-only validation details.

## REST v1 boundary

REST uses the versioned base path `/api/v1`. Active room commands and game actions remain on WebSocket.

### Create guest identity

`POST /api/v1/guests`

Required headers:

- `Idempotency-Key`: opaque client-generated retry key for this guest creation attempt.
- `X-Request-ID`: optional client correlation ID. The server generates one if absent and returns it as `request_id`.

Request:

```json
{ "display_name": "Raka" }
```

Response `201`:

```json
{
  "version": "v1",
  "request_id": "req_guest_01",
  "guest_id": "guest_123",
  "display_name": "Raka",
  "session_token": "opaque-session-token"
}
```

The client sends the session token in the WebSocket upgrade `Authorization: Bearer <session_token>` header. Tokens are opaque, must not be placed in a URL, and must not be logged.

`guest.identify` is not a WebSocket command. The REST response creates the identity, and the WebSocket authentication handshake uses the returned session token.

### List public rooms

`GET /api/v1/rooms?visibility=public`

Optional header: `X-Request-ID`, echoed as `request_id` in the response or error.

Response `200`:

```json
{
  "version": "v1",
  "request_id": "req_rooms_01",
  "rooms": [
    {
      "room_id": "room_123",
      "visibility": "public",
      "capacity": 4,
      "occupied_seats": 2,
      "status": "waiting"
    }
  ]
}
```

Private codes and player hand data are never included in room discovery.

### Match history

`GET /api/v1/matches/{match_id}`

Optional header: `X-Request-ID`, echoed as `request_id` in the response or error.

Response `200`:

```json
{
  "version": "v1",
  "request_id": "req_history_01",
  "match_id": "match_123",
  "round_count": 3,
  "final_ranking": [
    {
      "player_id": "player_123",
      "display_name": "Raka",
      "cumulative_score": 60,
      "rank": 1
    }
  ]
}
```

REST errors use HTTP status plus the same safe error payload fields: `version`, `code`, `message`, and `request_id`. REST responses never include active-round hidden hands or deck order.

## Shared types

### Player summary

```json
{
  "player_id": "player_123",
  "display_name": "Raka",
  "connection": "online",
  "is_host": true
}
```

`connection` is `online` or `offline`.

### Card projection

```json
{
  "card_id": "c_001",
  "rank": "Q",
  "suit": "diamonds",
  "is_active_joker": false
}
```

Card projections are included only where visibility allows them. The physical marker card is never projected. Valid ranks are `A`, `2` through `10`, `J`, `Q`, and `K`. Valid suits are `spades`, `hearts`, `diamonds`, and `clubs`.

### Table combination

```json
{
  "combination_id": "combo_123",
  "owner_player_id": "player_123",
  "type": "seri",
  "card_ids": ["c_010", "c_011", "c_012"],
  "joker_assignments": []
}
```

`type` is `seri` or `paralel`. `joker_assignments` contains representations for active jokers in the combination:

```json
{
  "joker_card_id": "c_020",
  "represents_rank": "8",
  "represents_suit": "hearts"
}
```

### Personalized game state

```json
{
  "match_id": "match_123",
  "round_number": 2,
  "round_count": 3,
  "round_status": "active",
  "end_reason": null,
  "current_player_id": "player_123",
  "phase": "discard_required",
  "active_joker_rank": "6",
  "own_hand": [
    {
      "card_id": "c_101",
      "rank": "A",
      "suit": "hearts",
      "is_active_joker": false
    }
  ],
  "opponents": [
    {
      "player_id": "player_456",
      "display_name": "Nina",
      "hand_size": 7,
      "connection": "online"
    }
  ],
  "discard_pile": [
    {
      "card_id": "c_301",
      "rank": "9",
      "suit": "clubs",
      "is_active_joker": false
    }
  ],
  "discard_pickup_window_size": 11,
  "table_combinations": [],
  "closed_deck_count": 32,
  "scoreboard": [
    {
      "player_id": "player_123",
      "display_name": "Raka",
      "completed_round_score": 24,
      "cumulative_score": 24
    }
  ]
}
```

`own_hand` is private. Opponent projections contain `hand_size` only. `discard_pile`, `table_combinations`, `active_joker_rank`, `closed_deck_count`, turn, phase, and completed-round scoreboard are public. `closed_deck_count` never implies deck order.

## Command payloads

| Name | Payload |
|---|---|
| `room.create` | `{ "visibility": "public", "player_count": 2 }` |
| `room.join` | `{ "room_id": "room_123" }` or `{ "private_code": "JADUL7" }` |
| `room.leave` | `{}` |
| `game.start` | `{}` |
| `game.draw` | `{ "source": "closed_deck" }` |
| `game.take_discard` | `{ "discard_card_id": "c_301", "jago_card_ids": ["c_101", "c_102"], "extension_card_ids": ["c_103"] }` |
| `game.laydown` | `{ "card_ids": ["c_101", "c_102", "c_103"], "type": "seri", "joker_assignments": [] }` |
| `game.discard` | `{ "card_id": "c_101" }` |
| `game.close` | `{ "closing_card_id": "c_101", "combinations": [] }` |
| `state.resync` | `{ "last_received_sequence": 41 }` |

The take-discard command selects two jago basis cards before the selected discard in the UI, but the wire payload contains IDs only. The server computes the eligible suffix, validates the jago and extensions, and commits the transfer atomically.

For `game.take_discard`:

- `jago_card_ids` must contain exactly two distinct physical card IDs owned by the player before pickup.
- `discard_card_id` must be one eligible card in the discard window.
- `extension_card_ids` is optional. Each ID must be distinct, must not equal the selected discard or a jago card, and may refer to an owned card or a card above the selected discard that will be transferred.
- The final newly laid combination contains the selected discard, both jago cards, and any extensions. It must be one valid seri or paralel.
- The server transfers the selected discard suffix, validates the complete action, and mutates nothing if validation fails.

For `game.laydown`, `card_ids` must be distinct cards owned by the player before the action. The server validates the complete combination and locks it atomically.

For `game.close`:

- `closing_card_id` must be one card owned by the player and is excluded from `combinations`.
- Each submitted combination intent has `card_ids`, `type`, and `joker_assignments`. It does not contain server-assigned `combination_id` or `owner_player_id`.
- Every remaining hand card must occur exactly once across the submitted combinations.
- The server validates the partition and closing discard atomically.

## Event payloads

| Name | Payload |
|---|---|
| `room.updated` | `{ "room": <personalized room shape> }` using the room shape below. |
| `game.started` | `{ "state": <personalized game state> }` using the state shape above. |
| `game.state` | Personalized game state from the accepted server state. |
| `game.drawn` | `{ "card": <card projection> }` to the drawing player only. |
| `game.action_rejected` | Error payload with a stable code. |
| `round.ended` | Round ended shape below with allowed scores only. |
| `match.updated` | Match update shape below. |
| `match.ended` | Final ranking shape below. |
| `state.resynced` | Personalized current snapshot and replacement sequence metadata. |

### Room shape

```json
{
  "room_id": "room_123",
  "visibility": "public",
  "private_code": null,
  "host_player_id": "player_123",
  "capacity": 4,
  "status": "waiting",
  "players": [
    {
      "player_id": "player_123",
      "display_name": "Raka",
      "connection": "online",
      "is_host": true
    }
  ]
}
```

`private_code` is present only for the creating host or a player who joined using that private room code. Public room lists omit it.

### Round ended shape

```json
{
  "round_number": 2,
  "reason": "close",
  "scores": [
    {
      "player_id": "player_123",
      "score_status": "calculated",
      "round_score": 36,
      "cumulative_score": 60
    },
    {
      "player_id": "player_456",
      "score_status": "calculated",
      "round_score": 14,
      "cumulative_score": 33
    }
  ]
}
```

For a game-over round, players whose scores are not calculated use `score_status: "not_calculated"` and omit `round_score`. The causing player includes the defined penalty.

### Match update and final ranking shapes

```json
{
  "round_number": 2,
  "scoreboard": [
    {
      "player_id": "player_123",
      "display_name": "Raka",
      "cumulative_score": 60
    }
  ]
}
```

```json
{
  "round_number": 3,
  "final_ranking": [
    {
      "player_id": "player_123",
      "display_name": "Raka",
      "cumulative_score": 60,
      "rank": 1
    }
  ]
}
```

Equal final scores receive the same rank. No hidden tie-break is included.

### Resync shape

```json
{
  "snapshot": {
    "match_id": "match_123",
    "round_number": 2,
    "round_count": 3,
    "round_status": "active",
    "end_reason": null,
    "current_player_id": "player_123",
    "phase": "discard_required",
    "active_joker_rank": "6",
    "own_hand": [],
    "opponents": [],
    "discard_pile": [],
    "discard_pickup_window_size": 11,
    "table_combinations": [],
    "closed_deck_count": 32,
    "scoreboard": []
  },
  "latest_sequence": 42,
  "replaced_since_sequence": 38,
  "mode": "snapshot"
}
```

For MVP, resync returns a personalized current snapshot rather than replaying individual events. The client discards locally buffered events at or below `latest_sequence` and applies the snapshot as the new source of truth.

## Idempotency and ordering

- Every client command has a `request_id`. It is unique across the authenticated guest session, including commands without a `room_id` such as `room.create`. Before authentication, REST guest creation uses an `Idempotency-Key` header.
- Repeating a previously accepted request returns the original accepted result and does not execute the action again.
- Repeating a previously rejected request returns the original error and does not mutate state.
- A new request with a reused ID and different payload returns `invalid_request`.
- Client commands do not include `sequence`.
- Server events include monotonically increasing `sequence` values per recipient stream. A private event delivered to one player does not consume a sequence value in another player's stream.
- A client detecting a gap sends `state.resync` with its last received sequence.
- Reconnect returns a personalized snapshot and enough metadata to discard stale events.

Room-scoped errors, including rejected commands, participate in the room sequence. Errors before a room is established may omit `sequence`.

Accepted action event order is deterministic within each recipient stream:

1. `game.drawn` to the drawing player only, if the action is a closed-deck draw.
2. Personalized `game.state` to every player.
3. `round.ended` after a completed round.
4. `match.updated` after `round.ended`.
5. `match.ended` after `match.updated` and only after round three.

Each delivered event receives its own recipient-stream sequence number. `game.started` is the initial personalized snapshot and is not duplicated by `game.state` for the same start action.

REST guest creation idempotency is defined as follows:

- Repeating the same `Idempotency-Key` with the same request body replays the original response, including its original status and `request_id`.
- Reusing the key with a different request body returns HTTP `409` with `invalid_request`.
- The key and original result are scoped to `POST /api/v1/guests` and retained for 24 hours. After expiration, reuse is treated as a new request.
