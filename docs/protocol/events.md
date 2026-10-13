# Events v1

Server events are ordered by `sequence` within a room stream.

| Name | Visibility | Purpose |
|---|---|---|
| `guest.identified` | private | Confirms guest identity.
| `room.updated` | room | Membership, host, room status, and capacity.
| `game.started` | personalized | Starts a round with private hand projection.
| `game.state` | personalized | Accepted state projection after an action.
| `game.drawn` | personalized | Draw result; only the drawing player receives the card identity.
| `game.action_rejected` | private | Stable error for the requesting command.
| `round.ended` | room | Round result, reason, and allowed scores.
| `match.updated` | room | Current round and cumulative public scoreboard.
| `match.ended` | room | Final cumulative ranking after round three.
| `state.resynced` | personalized | Snapshot/replay after reconnect.

Every projection must preserve hidden-card and deck-order secrecy.
