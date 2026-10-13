# Remi Jadul — Game Rules

This document is the sole implementation reference for MVP game rules. Product scope lives in `PRODUCT_BACKLOG.md`; architecture lives in `ARCHITECTURE.md`. If a rule is unclear, stop and create a decision record; do not choose silently.

## 0. Match and rounds

- One match consists of exactly 3 rounds.
- Each round is scored independently using the rules in this document.
- A player's match score is the sum of their scores from all completed rounds.
- The match ends after the third round. The player with the highest cumulative match score is the winner.
- Public match information includes the current round and a scoreboard of each player's cumulative score from completed rounds.
- Final winner and ranking are presented only after the third round ends.
- A game-over round still contributes its defined score to the match total.

## 1. Players, deck, and physical cards

- A round has 2, 3, or 4 players.
- Use one standard 52-card deck: ranks `A, 2–10, J, Q, K`; suits `♠, ♥, ♦, ♣`.
- There are no printed joker cards.
- Every physical card has a unique ID for the round.
- The server shuffles the deck using a server-side source of randomness.
- Cards are dealt from the shuffled deck; the order in which a player's hand is displayed has no game-rule meaning.
- A player may reorder their own hand for display without changing card ownership, card identity, or game state.

## 2. Deal and joker

| Players | Starting cards | Maximum hand before required resolution |
|---:|---:|---:|
| 2 | 11 | 11 |
| 3 | 9 | 9 |
| 4 | 7 | 7 |

After dealing, the server removes one random remaining card as the joker-rank marker. The marker itself is out of play. Every remaining card with the same rank becomes an active joker, regardless of suit or location.

Example: if `7♥` is the marker, `7♠`, `7♣`, and `7♦` are jokers; `7♥` is not in the round.

The server must never disclose the marker card or hidden deck order to clients.

## 3. Combinations

### 3.1 Seri

- At least 3 cards.
- Same suit.
- Consecutive ranks in ascending order.
- Active jokers may represent missing ranks.
- A joker’s score follows the card it represents.

Example: `3♥, 4♥, 7♠` is valid if `7♠` is an active joker representing `5♥`.

### 3.2 Paralel

- At least 3 cards.
- Same rank.
- Suits must be different.
- Active jokers may represent missing suits/cards.
- A physical card cannot be used twice in one combination or across locked table combinations.

### 3.3 Table behavior

- A player may lay down any valid combination from their hand; they are not required to lay down every available combination.
- The client does not suggest or display valid combinations; the player chooses their actions and the server validates them.
- A laid combination is locked.
- Cards in a locked combination cannot be returned, moved, merged, split, or reused.
- A laid joker cannot be exchanged for its represented natural card.

## 4. Turn lifecycle

Each normal turn is:

1. The server confirms that the player is current.
2. The player draws one card from the closed deck, or performs a valid discard pickup.
3. The player may lay down valid combinations according to the action rules.
4. The player discards exactly one card to end the turn.

The server rejects actions from other players, actions in the wrong phase, cards not owned by the player, duplicate physical cards, and invalid combinations.

## 5. Drawing from the closed deck

- Draw exactly one card from the top of the closed deck.
- The card is visible only to the drawing player until it becomes public through a valid table/discard action.
- A normal turn must end with exactly one discard unless the round ends first.
- If the closed deck is empty, the round ends and scores are calculated for all players unless the round already ended due to game over.

## 6. Taking a discard with a jago

The player may select a discard within the number of cards initially dealt, counted from the bottom of the discard pile. The selected card and every card above it are transferred to the player's hand. The selected card must be used immediately with two hand cards that form the jago basis of one newly laid combination. The jago combination may be extended with additional cards from the player's hand or the transferred cards when the final combination remains valid. The server validates the entire pickup and laydown before committing it.

- The eligible window is the bottom `N` cards of the discard pile, where `N` is the starting-card count for the number of players: 11, 9, or 7.
- The selected discard must be used immediately in a newly laid jago; it may not be kept for later use.
- A jago seri has two hand cards and the selected discard, all of the same suit and consecutive in ascending rank; additional cards may extend the run.
- A jago paralel has two hand cards and the selected discard, all of the same rank with different suits; additional valid cards may extend the parallel.
- Cards above the selected discard are taken into the hand and may be kept or used in other actions on the same turn.
- If validation fails, the discard remains in place and the hand/round state is unchanged.

Example: with two players, a discard at position 9 from the bottom is eligible. If `2♥` is selected and `2♦,2♣` are in hand, the player may lay `2♦,2♣,2♥` as a jago paralel and takes every discard above `2♥` into their hand. If `6♥` is selected with `4♥,5♥` in hand and `3♥` is among the transferred cards, the player may lay the extended seri `3♥,4♥,5♥,6♥`.

## 7. Hand limit and game over

The hand limit is the starting-card count for the number of players: 11, 9, or 7.

If taking a discard makes the player exceed the limit and the player cannot immediately lay a valid combination using that discard, the player enters game over. The round ends immediately.

Discarding an active joker also causes game over immediately. The player who discarded the active joker receives `-50`, other players' scores are not calculated for that round, and the server records the end reason as `game_over`.

Game-over scoring:

- The player causing game over receives `-50`.
- Other players’ scores are not calculated for that round.
- The server records the end reason as `game_over`.

## 8. Closing the round

A player may close when all remaining cards in their hand form valid seri/paralel combinations and the player can discard the closing card. Closing on the first turn is allowed when all conditions are met.

The server validates the complete close atomically before ending the round. The closing player receives the special closing-card value in addition to the values from their valid combinations.

When a player closes, every player is scored normally: valid combinations are positive and cards not belonging to a valid combination are negative.

## 9. Scoring

### 9.1 Combination card values

| Rank | Value |
|---|---:|
| Ace | +3 |
| 2–10 | +1 |
| J, Q, K | +2 |
| Active joker | Value of represented card |

### 9.2 Closing card values

| Rank | Value |
|---|---:|
| Ace | +15 |
| 2–10 | +5 |
| J, Q, K | +10 |

### 9.3 Uncombined cards

Cards not belonging to a valid combination are negative using the base combination value: Ace `-3`, 2–10 `-1`, J/Q/K `-2`. A joker uses the value of its representation.

The engine chooses the valid hand partition with the highest total score. Equal-score partitions use a deterministic physical-card-ID comparator and must not depend on map iteration order. The behavior is recorded in decision `0002`.

### 9.4 Ranking

- Highest score ranks first.
- Equal scores receive the same rank; no hidden tie-break is applied.
- A game-over round has no normal ranking because other player scores are not calculated.

Example: seri `3♥,4♥,5♥` = `+3`, paralel `9♣,9♦,9♠` = `+3`, uncombined `K♠` = `-2`, closing `K♦` = `+10`; total `14`.

## 10. Atomicity and visibility invariants

- Validate before mutate for every action.
- A rejected action produces no partial state change.
- Only the owner sees their hand; other players see hand count only.
- Public state includes turn, the complete discard pile in discard order, every locked table combination from every player, round status, and scores only when rules allow them.
- Locked table combinations are public card information and may be used by players to estimate draw possibilities; the closed-deck order remains hidden.
- The discard pile's last card is the current top discard; pickup eligibility is governed by the discard window and jago rules above.
- Event sequence numbers or equivalent ordering metadata are required for reconnect and stale-event handling.

## 11. Accepted edge behavior

- After a closed-deck draw, the player may discard any card they own; a new combination is not required.
- A discard pickup must create exactly one new combination containing the picked discard.
- A close removes exactly one closing card from the hand; all remaining cards must be partitionable into valid combinations.
- Equal final scores share a rank.
- Scoring and close behavior follow decision `0002`.
