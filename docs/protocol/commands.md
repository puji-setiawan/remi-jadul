# Commands v1

Commands express player intent. They never contain an authoritative hand, score, turn, or full state replacement.

Exact payload shapes and retry behavior are defined in [`contract.md`](contract.md); examples are in [`examples.md`](examples.md).

| Name | Payload | Result |
|---|---|---|
| `room.create` | visibility, player count | room state |
| `room.join` | room ID or private code | room state |
| `room.leave` | none | room state |
| `game.start` | none | game started/state |
| `game.draw` | source `closed_deck` | private draw + public state |
| `game.take_discard` | discard card ID, two-card jago basis, optional extensions | accepted state/error |
| `game.laydown` | selected physical card IDs and intended combination | accepted state/error |
| `game.discard` | one physical card ID | accepted state/error or game over |
| `game.close` | closing card ID and intended remaining combinations | round result/error |
| `state.resync` | last received sequence | ordered state replay/snapshot |

The server validates ownership, turn, phase, physical card IDs, combination rules, visibility, and atomicity for every command.
