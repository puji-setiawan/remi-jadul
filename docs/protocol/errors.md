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

Errors must not disclose another player’s cards, the marker card, closed-deck order, or server-only validation details.
