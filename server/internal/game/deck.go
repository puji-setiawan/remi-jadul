package game

// suitsInCanonicalOrder defines the deterministic order used by
// NewStandardDeck. It is not shuffled and depends on no I/O or entropy.
var suitsInCanonicalOrder = []Suit{SuitSpades, SuitHearts, SuitDiamonds, SuitClubs}

var ranksInCanonicalOrder = []Rank{
	RankAce, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9,
	Rank10, RankJack, RankQueen, RankKing,
}

// NewStandardDeck returns the canonical 52-card deck in a deterministic order.
// Callers are responsible for providing whatever shuffled order they need;
// this function only establishes physical card identities.
func NewStandardDeck() []Card {
	deck := make([]Card, 0, 52)
	id := 0
	for _, suit := range suitsInCanonicalOrder {
		for _, rank := range ranksInCanonicalOrder {
			deck = append(deck, Card{ID: id, Rank: rank, Suit: suit})
			id++
		}
	}
	return deck
}

// StandardDeck is an alias for NewStandardDeck.
func StandardDeck() []Card { return NewStandardDeck() }

// CanonicalDeck is an alias for NewStandardDeck.
func CanonicalDeck() []Card { return NewStandardDeck() }

// NewDeck is an alias for NewStandardDeck.
func NewDeck() []Card { return NewStandardDeck() }

// FullDeck is an alias for NewStandardDeck.
func FullDeck() []Card { return NewStandardDeck() }
