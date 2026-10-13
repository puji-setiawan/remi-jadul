package game

import "errors"

// Deal errors.
var (
	ErrInvalidPlayerCount = errors.New("game: invalid player count")
	ErrDuplicateCard      = errors.New("game: duplicate physical card id")
	ErrInsufficientCards  = errors.New("game: insufficient cards for deal")
)

var playerHandSize = map[int]int{
	2: 11,
	3: 9,
	4: 7,
}

// DealResult captures the outcome of one round-local deal. The physical
// marker card itself is intentionally absent: only its rank is exposed as
// JokerRank. Remaining/ClosedDeck are the cards still in play after dealing.
type DealResult struct {
	// Hand slices, one entry per player.
	Hands         [][]Card
	PlayerHands   [][]Card
	HandsByPlayer [][]Card
	Cards         [][]Card
	Players       [][]Card
	HandsSlice    [][]Card

	// Remaining cards in play after the marker is removed.
	Remaining      []Card
	Stock          []Card
	ClosedDeck     []Card
	DrawPile       []Card
	RemainingCards []Card
	Rest           []Card
	Leftover       []Card
	Deck           []Card

	// Active joker rank derived from the excluded marker.
	JokerRank       Rank
	ActiveJokerRank Rank
	Joker           Rank
	JokerRankValue  Rank
}

func startingHandSize(players int) (int, bool) {
	size, ok := playerHandSize[players]
	return size, ok
}

// Deal deals cards from a caller-provided deck to the requested number of
// players. The deck ordering is trusted as the shuffle input; this function
// does not use time, randomness, or I/O. Invalid player counts, duplicate
// physical cards, or insufficient cards produce an error without mutating the
// supplied deck.
func Deal(deck []Card, players int) (DealResult, error) {
	handSize, ok := startingHandSize(players)
	if !ok {
		return DealResult{}, ErrInvalidPlayerCount
	}
	if err := checkDuplicateCards(deck); err != nil {
		return DealResult{}, err
	}
	dealtCards := players * handSize
	if len(deck) < dealtCards+1 {
		return DealResult{}, ErrInsufficientCards
	}

	hands := make([][]Card, players)
	for i := 0; i < players; i++ {
		start := i * handSize
		end := start + handSize
		hand := make([]Card, handSize)
		copy(hand, deck[start:end])
		hands[i] = hand
	}

	marker := deck[dealtCards]
	jokerRank := marker.Rank

	remaining := make([]Card, len(deck)-dealtCards-1)
	copy(remaining, deck[dealtCards+1:])

	result := DealResult{
		Hands:           hands,
		PlayerHands:     cloneHands(hands),
		HandsByPlayer:   cloneHands(hands),
		Cards:           cloneHands(hands),
		Players:         cloneHands(hands),
		HandsSlice:      cloneHands(hands),
		Remaining:       remaining,
		Stock:           cloneCards(remaining),
		ClosedDeck:      cloneCards(remaining),
		DrawPile:        cloneCards(remaining),
		RemainingCards:  cloneCards(remaining),
		Rest:            cloneCards(remaining),
		Leftover:        cloneCards(remaining),
		Deck:            cloneCards(remaining),
		JokerRank:       jokerRank,
		ActiveJokerRank: jokerRank,
		Joker:           jokerRank,
		JokerRankValue:  jokerRank,
	}
	return result, nil
}

func checkDuplicateCards(deck []Card) error {
	seen := make(map[any]struct{}, len(deck))
	for _, card := range deck {
		if _, exists := seen[card.ID]; exists {
			return ErrDuplicateCard
		}
		seen[card.ID] = struct{}{}
	}
	return nil
}

func cloneCards(cards []Card) []Card {
	out := make([]Card, len(cards))
	copy(out, cards)
	return out
}

func cloneHands(hands [][]Card) [][]Card {
	out := make([][]Card, len(hands))
	for i, hand := range hands {
		out[i] = cloneCards(hand)
	}
	return out
}

func DealCards(deck []Card, players int) (DealResult, error)  { return Deal(deck, players) }
func DealDeck(deck []Card, players int) (DealResult, error)   { return Deal(deck, players) }
func DealHands(deck []Card, players int) (DealResult, error)  { return Deal(deck, players) }
func CreateDeal(deck []Card, players int) (DealResult, error) { return Deal(deck, players) }
func NewDeal(deck []Card, players int) (DealResult, error)    { return Deal(deck, players) }
func DealCardsForPlayers(deck []Card, players int) (DealResult, error) {
	return Deal(deck, players)
}

// DealWithPlayers keeps the result-based API for callers that prefer a
// player count first. The deck remains the externally provided shuffle input.
func DealWithPlayers(players int, deck []Card) (DealResult, error) {
	return Deal(deck, players)
}
