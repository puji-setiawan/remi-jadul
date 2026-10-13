# Protocol v1 Contract Fixtures

These examples are review fixtures. They are intentionally small but must preserve the visibility rules from `visibility.md`.

## Draw command and private result

Command:

```json
{
  "version": "v1",
  "type": "command",
  "name": "game.draw",
  "request_id": "req_draw_01",
  "room_id": "room_123",
  "payload": { "source": "closed_deck" }
}
```

Private result for the drawing player:

```json
{
  "version": "v1",
  "type": "event",
  "name": "game.drawn",
  "request_id": "req_draw_01",
  "room_id": "room_123",
  "sequence": 12,
  "payload": {
    "card": {
      "card_id": "c_901",
      "rank": "Q",
      "suit": "hearts",
      "is_active_joker": false
    }
  }
}
```

The same card identity is absent from every other player's event.

## Discard pickup command

```json
{
  "version": "v1",
  "type": "command",
  "name": "game.take_discard",
  "request_id": "req_pickup_01",
  "room_id": "room_123",
  "payload": {
    "discard_card_id": "c_301",
    "jago_card_ids": ["c_101", "c_102"],
    "extension_card_ids": ["c_103"]
  }
}
```

The server checks the discard window, takes the selected suffix atomically, validates the jago basis and extensions, and then emits personalized `game.state` events.

## Visibility-safe public state

```json
{
  "version": "v1",
  "type": "event",
  "name": "game.state",
  "room_id": "room_123",
  "sequence": 13,
  "payload": {
    "current_player_id": "player_123",
    "active_joker_rank": "6",
    "discard_pile": [{ "card_id": "c_301", "rank": "9", "suit": "clubs", "is_active_joker": false }],
    "table_combinations": [],
    "opponents": [{ "player_id": "player_456", "hand_size": 7, "connection": "online" }]
  }
}
```

This projection intentionally contains no opponent card IDs, ranks, suits, marker card, or deck order.

## Rejected action

```json
{
  "version": "v1",
  "type": "error",
  "name": "game.action_rejected",
  "request_id": "req_pickup_02",
  "room_id": "room_123",
  "sequence": 14,
  "payload": {
    "code": "invalid_jago",
    "message": "Kartu jago tidak valid.",
    "retryable": false
  }
}
```

The discard pile, hand, table, turn, and scores remain unchanged.

## Required error fixtures

These errors use the same envelope and payload shape:

```json
{
  "version": "v1",
  "type": "error",
  "name": "game.action_rejected",
  "request_id": "req_error_01",
  "room_id": "room_123",
  "sequence": 15,
  "payload": {
    "code": "not_your_turn",
    "message": "Aksi belum tersedia pada giliran ini.",
    "retryable": false
  }
}
```

Required code variants:

```json
[
  { "code": "not_your_turn", "message": "Aksi belum tersedia pada giliran ini.", "retryable": false },
  { "code": "stale_request", "message": "Permintaan sudah tidak berlaku.", "retryable": true },
  { "code": "unknown_card", "message": "Kartu tidak dikenali.", "retryable": false },
  { "code": "invalid_request", "message": "Permintaan tidak valid.", "retryable": false }
]
```

A client attempting to submit another player's card receives `card_not_owned`, not any information about that player's actual hand. A client attempting to submit hidden card data receives `invalid_request` or `card_not_owned` according to the public error contract, without server-only validation details.

Unknown message:

```json
{
  "version": "v1",
  "type": "error",
  "name": "protocol.error",
  "request_id": "req_unknown_01",
  "room_id": "room_123",
  "sequence": 16,
  "payload": {
    "code": "unknown_message",
    "message": "Jenis pesan tidak didukung.",
    "retryable": false
  }
}
```

## Resync

```json
{
  "version": "v1",
  "type": "command",
  "name": "state.resync",
  "request_id": "req_resync_01",
  "room_id": "room_123",
  "payload": { "last_received_sequence": 11 }
}
```

The response contains a personalized snapshot and replacement sequence metadata. MVP resync replaces the client snapshot instead of replaying individual events. It never fills private fields with public placeholders or another player's cards.
