# Remi Jadul: Gameplay UI Plan

## Status

Draft implementation plan for the Flutter Android client. Game rules remain authoritative in [`GAME_RULES.md`](GAME_RULES.md), and the server remains authoritative for card ownership, validation, turns, scoring, and round completion.

## Product intent

The gameplay screen should feel like a warm tabletop card game: readable at phone size, calm during planning, and clear when an action changes the state. The player decides which combinations to make. The UI must not suggest valid combinations or decide outcomes.

Design direction comes from [`DESIGN.md`](../DESIGN.md): felt-green table, warm card paper, restrained suit colors, clear card geometry, one-handed Android use, and purposeful motion only for card movement and turn changes.

Design dials: ENERGY 2, RHYTHM 2, MOTION 1.

## Card asset source

The card direction is an original, scalable vector treatment designed for Remi Jadul:

- Local sample: `app/assets/cards/remi-modern/card-faces.svg`
- Style: modern tabletop premium, warm paper, deep felt, restrained red and black suits, and a gold structural accent.
- The sample is original project artwork and does not depend on an external asset license.

The Flutter implementation should turn the sample into rank and suit components while keeping the server's physical card ID separate from the visual asset. The visual layer must not own deck state, joker selection, validation, or scoring. The Kenney files remain only as an unselected reference and must not be used by the final UI.

## Gameplay screen hierarchy

### 1. Match header

The gameplay header is integrated into the right-side information panel. Show:

- Room or game title.
- Match status and connection state.
- Current round, for example `Ronde 2 dari 3`.
- Current turn and phase.

The cumulative scoreboard sits below this information in the same right-side panel. There is no separate full-width top header on the gameplay screen. The final winner and ranking appear only after round three. Do not show an early winner during an active match.

### 2. Public table

Show every locked combination from every player:

- Own seri and paralel.
- Opponent seri and paralel.
- Owner label for each group.
- Locked state that prevents moving, splitting, merging, or reusing cards.

All table cards are public information. The opponent's unplayed hand remains private.

### Player information around the table

Each seat shows:

- Player name or seat label.
- Hand size as a row of face-down card icons, not a numeric count.
- All locked seri and paralel owned by that player in the nearby meld zone.
- Connection status with both an icon and text such as `ONLINE` or `OFFLINE`; status must not rely on color alone.
- Turn state when that player is active.

The opponent's card faces remain hidden. The icon count communicates hand size without exposing rank or suit.

### 3. Discard pile

Show the complete discard pile in discard order, from the bottom card to the top card.

- Mark the current top card.
- Mark the eligible pickup window based on player count: 11 cards for 2 players, 9 for 3 players, and 7 for 4 players.
- Make cards outside the eligible window visibly unavailable for pickup.
- When a middle card is selected, show that the selected card and every card above it will be transferred to the player's hand.
- Do not expose discarded cards as possible closed-deck order. They are only public history.

The discard pile must remain usable on a narrow screen through horizontal scrolling or a compact stacked layout without hiding card identity.

### 4. Joker information

Show the active joker rank and visually mark physical cards that are active jokers.

- Do not reveal the physical marker card.
- Do not reveal the closed-deck order.
- Make the active joker warning visible before a discard is submitted.
- A rejected active-joker discard must explain that it ends the round with the defined penalty, when the server reports that rule error or game-over result.

### 5. Own hand

Show every card owned by the current player.

- Allow drag reordering for presentation only.
- Use physical card IDs for actions, never hand positions.
- Preserve the selected state when the hand wraps or scrolls.
- Do not highlight, group, or recommend possible seri or paralel combinations.
- Provide semantic labels containing rank, suit, joker status, and selection state.

### 6. Opponent hand

Show only:

- Player name or seat.
- Number of cards in hand.
- Turn state.

Never show hidden opponent cards, inferred card identities, or hidden deck information.

### 7. Action area

Actions must reflect the server phase and current player's turn:

- `Ambil dari buangan`
- `Turunkan`
- `Tutup`

Drawing is performed by tapping the deck. Discarding is performed by dragging a hand card to the discard target, with a tap-selection fallback for accessibility. Each remaining enabled control must perform a real command. Disabled controls must explain the current reason where practical. The UI sends player intent and the server validates the complete action atomically.

On wider gameplay layouts, the table shifts left to make room for the scoreboard panel on the right. The player's hand remains anchored at the bottom and spans the available gameplay width below the table.

## Primary card interactions and animation

### Draw from the deck

1. The current player taps the deck in the center of the table.
2. The deck enters a brief pressed state and the action is sent to the server.
3. After the accepted draw event arrives, a card animates from the deck to the player's hand.
4. The hand reflows and the new card receives a short focus treatment.

The client must not reveal the card identity before the accepted server event. Draw animation duration should be about 200 milliseconds and must respect reduced-motion settings.

### Discard from the hand

1. The player presses and drags a card from their hand.
2. The discard area becomes a clear drop target while the card is held.
3. On release over the discard area, the client submits the discard intent.
4. After server acceptance, the card moves into the discard pile and the hand reflows.
5. If rejected, the card returns to its original hand position and an actionable error is shown.

Discard animation duration should be about 200 milliseconds. The UI must not remove the card permanently before server acceptance.

Tap fallback for discard:

- Tap a card to select it.
- Tap the discard target.
- The same server validation and accepted-event animation apply.

The drag interaction is the primary physical-table gesture. The tap fallback keeps the action accessible and reliable on narrow screens.

Persistent `Draw` and `Buang` buttons are not shown in the action area.

## Discard pickup flow

1. Player taps exactly two cards in their hand as the jago basis.
2. Player taps one eligible discard card.
3. UI displays the selected discard and the suffix of cards above it that will enter the hand.
4. Player may add cards to extend the resulting seri or paralel when the final combination remains valid.
5. The third selection, the discard card, submits the pickup and laydown command immediately. No persistent `Ambil` button is shown.
6. On success, the table, hand, discard pile, and turn update from the server event.
7. On rejection, all visible state returns to the last accepted server state and an actionable error is shown.

The UI must not decide whether a jago, extension, or close is valid.

Before the third selection, the player can cancel by tapping a selected card again. After the third selection, the UI shows a brief submitting state while the server validates the complete action.

## Turn and round flows

### Normal turn

1. Server confirms the player turn.
2. Player draws from the closed deck or selects an eligible discard pickup.
3. Player optionally lays down selected cards.
4. Player discards exactly one card unless the round ends first.
5. Server broadcasts the accepted state and advances the turn.

### Close

Show a confirmation step only after the player explicitly chooses to close. The server performs the complete close validation. On success, show the round result and updated cumulative scoreboard.

### Game over

Show a prominent but calm result state containing:

- End reason.
- Player penalty or defined score.
- Round number.
- Updated match scoreboard when available.

Do not allow another player action after a game-over event.

### Match completion

After round three:

- Show final cumulative scores.
- Show ranking and winner.
- Preserve the public table and discard history as round result context when available.
- Provide only real actions supported by the product, such as returning to the room or starting a defined new match.

## Public and private information matrix

| Information | Current player | Opponents | Public table |
|---|---:|---:|---:|
| Own hand | Full cards | No | No |
| Opponent hand | Count only | Own hand only | No |
| Complete discard pile | Yes | Yes | Yes |
| Locked table combinations | Yes | Yes | Yes |
| Active joker rank | Yes | Yes | Yes |
| Physical joker marker | No | No | No |
| Closed-deck order | No | No | No |
| Turn and phase | Yes | Yes | Yes |
| Cumulative scoreboard | Yes | Yes | Yes |

## Required UI states

- Loading room state.
- Waiting for players.
- Waiting for host start.
- Player turn.
- Opponent turn.
- Draw pending.
- Pickup selection pending.
- Laydown selection pending.
- Discard required.
- Action submitted.
- Invalid action with server error.
- Rejected pickup with unchanged state.
- Reconnecting.
- Reconnected with stale-event recovery.
- Closed deck empty.
- Round close result.
- Game-over result.
- Match final result.
- Empty table combinations.
- Empty eligible discard window when the pile is too short for a meaningful selection.

## Accessibility and mobile behavior

- Design narrow Android screens first.
- Keep primary action and turn state visible without opening a menu.
- Use at least 44dp touch targets for cards and controls.
- Do not communicate suit, selection, turn, joker status, or errors with color alone.
- Use rank and suit text or semantic labels as a second signal.
- Maintain readable contrast on the green table and ivory cards.
- Support scalable text without hiding critical controls.
- Respect reduced-motion preferences.
- Keep drag reorder usable with an accessible alternative, such as select card then choose a move command.

## Motion rules

- Tapping the deck is the draw gesture and animates an accepted card from the deck to the hand.
- Dragging a hand card to the discard target is the discard gesture and animates an accepted card into the discard pile.
- Animate draw from the deck to the hand to explain card movement.
- Animate discard from the hand to the discard pile.
- Animate accepted laydown to the public table.
- Animate turn changes briefly.
- Do not use looping decoration, celebratory motion during ordinary turns, or animation that delays an action.

## Contract and state requirements

The UI implementation depends on versioned server projections containing at least:

- Match and round number.
- Turn, phase, round status, and end reason.
- Personalized own hand with physical card IDs.
- Opponent hand counts.
- Active joker rank without the marker card.
- Complete discard pile in order.
- All locked table combinations with owner identity.
- Cumulative scoreboard.
- Event sequence metadata for reconnect and stale-event handling.

The Flutter client must not create authoritative card identities, validate combinations, calculate scores, choose a winner, or infer hidden cards from UI state.

## Acceptance criteria

- The player can see the full discard pile and identify the current top card.
- The player can identify the eligible discard window for the current player count.
- The player can see the active joker rank without seeing the marker card.
- The player can reorder their own hand without changing ownership or card identity.
- Every own and opponent locked seri or paralel is visible on the table.
- The UI never suggests a valid combination.
- The opponent's hand contents and closed-deck order never appear.
- Invalid actions show server errors without partial local mutation.
- A successful pickup shows transferred cards and the newly locked jago or extended combination.
- A game-over event blocks further actions.
- The scoreboard updates after completed rounds and final ranking appears only after round three.
- Loading, empty, error, reconnect, and final states have explicit UI treatment.
- Mobile touch targets, semantic labels, contrast, text scaling, and reduced motion are verified.

## Verification plan

Before delivery, run the Flutter formatter, analyzer, unit/widget tests, and Android emulator acceptance flow. Test valid, invalid, and boundary cases for draw, discard, pickup window, jago extension, hand reorder, joker discard game over, visibility, reconnect, and final ranking.
