package game

// Rank is a card rank. It uses an integer representation so equal ranks
// compare without string normalization.
type Rank int

const (
	RankAce Rank = iota + 1
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJack
	RankQueen
	RankKing
)

// Common short rank aliases.
const (
	Ace   = RankAce
	Two   = Rank2
	Three = Rank3
	Four  = Rank4
	Five  = Rank5
	Six   = Rank6
	Seven = Rank7
	Eight = Rank8
	Nine  = Rank9
	Ten   = Rank10
	Jack  = RankJack
	Queen = RankQueen
	King  = RankKing
)

func (r Rank) String() string {
	switch r {
	case RankAce:
		return "A"
	case Rank2:
		return "2"
	case Rank3:
		return "3"
	case Rank4:
		return "4"
	case Rank5:
		return "5"
	case Rank6:
		return "6"
	case Rank7:
		return "7"
	case Rank8:
		return "8"
	case Rank9:
		return "9"
	case Rank10:
		return "10"
	case RankJack:
		return "J"
	case RankQueen:
		return "Q"
	case RankKing:
		return "K"
	default:
		return "?"
	}
}

// Suit is a card suit.
type Suit int

const (
	SuitSpades Suit = iota
	SuitHearts
	SuitDiamonds
	SuitClubs
)

// Singular and plural suit aliases.
const (
	Spade    = SuitSpades
	Spades   = SuitSpades
	Heart    = SuitHearts
	Hearts   = SuitHearts
	Diamond  = SuitDiamonds
	Diamonds = SuitDiamonds
	Club     = SuitClubs
	Clubs    = SuitClubs
)

func (s Suit) String() string {
	switch s {
	case SuitSpades:
		return "♠"
	case SuitHearts:
		return "♥"
	case SuitDiamonds:
		return "♦"
	case SuitClubs:
		return "♣"
	default:
		return "?"
	}
}

// Card is one physical card for the duration of a round. Its ID is
// round-local and must be unique within a round.
type Card struct {
	ID   any
	Rank Rank
	Suit Suit
}
