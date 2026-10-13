package game

import (
	"errors"
	"reflect"
)

// Round state errors.
var (
	ErrInvalidRound     = errors.New("game: invalid round state")
	ErrNotCurrentPlayer = errors.New("game: not current player")
	ErrAlreadyDrawn     = errors.New("game: card already drawn")
	ErrDrawRequired     = errors.New("game: draw required before discard")
	ErrClosedDeckEmpty  = errors.New("game: closed deck is empty")
	ErrCardNotInHand    = errors.New("game: card not in hand")
)

// TurnPhase reports whether the current player has already drawn.
type TurnPhase int

const (
	TurnPhaseAwaitingDraw TurnPhase = iota
	TurnPhaseDrawn
)

// Round is the in-memory pure round state owned by the server. It intentionally
// stores only card identities and public turn data; it does not perform I/O.
type Round struct {
	PlayerCount   int
	Hands         [][]Card
	CurrentPlayer int
	ClosedDeck    []Card
	DiscardPile   []Card
	JokerRank     Rank
	Table         Table
	HasDrawn      bool
}

// GameState and Game are public aliases for Round.
type GameState = Round
type Game = Round

func (r *Round) Phase() TurnPhase {
	if r.HasDrawn {
		return TurnPhaseDrawn
	}
	return TurnPhaseAwaitingDraw
}

// Current returns the seat index of the acting player.
func (r *Round) Current() int { return r.CurrentPlayer }

// CurrentPlayerSeat is a compatibility alias for Current.
func (r *Round) CurrentPlayerSeat() int { return r.CurrentPlayer }

func clonesCardsForRound(cards []Card) []Card {
	out := make([]Card, len(cards))
	copy(out, cards)
	return out
}

func cloneHandsForRound(hands [][]Card) [][]Card {
	out := make([][]Card, len(hands))
	for i, hand := range hands {
		out[i] = clonesCardsForRound(hand)
	}
	return out
}

func makeRound(players int, hands [][]Card, closedDeck []Card, jokerRank Rank) (*Round, error) {
	if players < 2 || players > 4 {
		return nil, ErrInvalidPlayerCount
	}
	if len(hands) != players {
		return nil, ErrInvalidRound
	}
	round := &Round{
		PlayerCount:   players,
		Hands:         cloneHandsForRound(hands),
		ClosedDeck:    clonesCardsForRound(closedDeck),
		DiscardPile:   []Card{},
		JokerRank:     jokerRank,
		Table:         NewTable(),
		HasDrawn:      false,
		CurrentPlayer: 0,
	}
	return round, nil
}

// NewRound accepts either (players, hands, closedDeck, jokerRank) or
// (hands, closedDeck, jokerRank).
func NewRound(args ...interface{}) (*Round, error) {
	switch len(args) {
	case 4:
		players, ok1 := args[0].(int)
		if !ok1 {
			return nil, ErrInvalidRound
		}
		var hands [][]Card
		if args[1] == nil {
			hands = nil
		} else {
			value, ok := args[1].([][]Card)
			if !ok {
				return nil, ErrInvalidRound
			}
			hands = value
		}
		var closed []Card
		if args[2] == nil {
			closed = nil
		} else {
			value, ok := args[2].([]Card)
			if !ok {
				return nil, ErrInvalidRound
			}
			closed = value
		}
		joker, ok4 := args[3].(Rank)
		if !ok4 {
			return nil, ErrInvalidRound
		}
		return makeRound(players, hands, closed, joker)
	case 3:
		var hands [][]Card
		if args[0] == nil {
			hands = nil
		} else {
			value, ok := args[0].([][]Card)
			if !ok {
				return nil, ErrInvalidRound
			}
			hands = value
		}
		var closed []Card
		if args[1] == nil {
			closed = nil
		} else {
			value, ok := args[1].([]Card)
			if !ok {
				return nil, ErrInvalidRound
			}
			closed = value
		}
		joker, ok3 := args[2].(Rank)
		if !ok3 {
			return nil, ErrInvalidRound
		}
		if len(hands) == 0 {
			return nil, ErrInvalidRound
		}
		return makeRound(len(hands), hands, closed, joker)
	default:
		return nil, ErrInvalidRound
	}
}

// Compatibility constructors.
func NewRoundState(args ...interface{}) (*Round, error) { return NewRound(args...) }
func NewGameState(args ...interface{}) (*Round, error)  { return NewRound(args...) }
func NewGame(args ...interface{}) (*Round, error)       { return NewRound(args...) }
func MakeRound(args ...interface{}) (*Round, error)     { return NewRound(args...) }

func cardEqualForRound(a, b Card) bool {
	return reflect.DeepEqual(a.ID, b.ID) && a.Rank == b.Rank && a.Suit == b.Suit
}

func removeCardFromHand(hand []Card, target Card) ([]Card, bool) {
	for i, card := range hand {
		if cardEqualForRound(card, target) {
			out := make([]Card, len(hand)-1)
			copy(out, hand[:i])
			copy(out[i:], hand[i+1:])
			return out, true
		}
	}
	return nil, false
}

// Draw moves one card from the closed deck to the acting player's hand.
func (r *Round) Draw(player int) error {
	if player != r.CurrentPlayer {
		return ErrNotCurrentPlayer
	}
	if r.HasDrawn {
		return ErrAlreadyDrawn
	}
	if len(r.ClosedDeck) == 0 {
		return ErrClosedDeckEmpty
	}
	next := *r
	next.Hands = cloneHandsForRound(r.Hands)
	card := r.ClosedDeck[0]
	next.ClosedDeck = clonesCardsForRound(r.ClosedDeck[1:])
	next.Hands[player] = append(clonesCardsForRound(r.Hands[player]), card)
	next.HasDrawn = true
	*r = next
	return nil
}

// DrawCard is an alias for Draw.
func (r *Round) DrawCard(player int) error { return r.Draw(player) }

// Discard ends the current turn by removing one owned card and advancing play.
func (r *Round) Discard(player int, card Card) error {
	if player != r.CurrentPlayer {
		return ErrNotCurrentPlayer
	}
	if !r.HasDrawn {
		return ErrDrawRequired
	}
	hand, ok := removeCardFromHand(r.Hands[player], card)
	if !ok {
		return ErrCardNotInHand
	}
	next := *r
	next.Hands = cloneHandsForRound(r.Hands)
	next.Hands[player] = hand
	next.DiscardPile = append(clonesCardsForRound(r.DiscardPile), card)
	next.HasDrawn = false
	next.CurrentPlayer = (r.CurrentPlayer + 1) % r.PlayerCount
	*r = next
	return nil
}
