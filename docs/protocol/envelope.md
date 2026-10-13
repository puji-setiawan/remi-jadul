# WebSocket Envelope v1

All WebSocket messages use one JSON envelope. Fields not applicable to a message are omitted, except `version`, `type`, and `name`.

```json
{
  "version": "v1",
  "type": "command",
  "name": "game.draw",
  "request_id": "req_01J...",
  "room_id": "room_123",
  "sequence": 42,
  "payload": {}
}
```

## Fields

- `version`: protocol version, currently `v1`.
- `type`: `command`, `event`, or `error`.
- `name`: stable namespaced message name.
- `request_id`: client-generated for commands; echoed by correlated responses/errors.
- `room_id`: room scope when applicable.
- `sequence`: server event ordering value; absent on client commands.
- `payload`: typed message body.

Unknown fields must be ignored for forward compatibility. Unknown message names must produce a typed error without mutating state.
