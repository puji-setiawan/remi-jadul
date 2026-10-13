# Errors v1

Errors use stable machine-readable codes and human-readable messages safe for the requesting player.

Required initial codes:

- `unauthorized`
- `invalid_request`
- `not_found`
- `room_full`
- `room_private`
- `not_host`
- `game_not_started`
- `not_your_turn`
- `invalid_phase`
- `card_not_owned`
- `unknown_card`
- `invalid_combination`
- `invalid_discard_window`
- `invalid_jago`
- `stale_request`
- `sequence_gap`
- `game_over`
- `unknown_message`

Errors must not disclose another player’s cards, the marker card, closed-deck order, or server-only validation details.

An unknown command or event name uses `protocol.error` with code `unknown_message`. It includes the received name only when that name is safe to echo and never mutates room state.
