package game

import "errors"

// Combination validation errors.
var (
	ErrInvalidCombination       = errors.New("game: invalid combination")
	ErrTooFewCombinationCards   = errors.New("game: combination requires at least 3 cards")
	ErrCombinationDuplicateCard = errors.New("game: duplicate physical card in combination")
	ErrInvalidSeri              = errors.New("game: invalid seri")
	ErrInvalidParalel           = errors.New("game: invalid paralel")
)

// CombinationKind identifies which combination type to validate.
type CombinationKind int

const (
	CombinationSeri CombinationKind = iota + 1
	CombinationParalel
)

const CombinationParallel = CombinationParalel

// IsJoker reports whether card has the round's active joker rank.
func IsJoker(card Card, jokerRank Rank) bool {
	return card.Rank == jokerRank
}

// IsValidSeri reports whether cards form one valid Seri, allowing active
// jokers to represent missing ranks in one suit.
func IsValidSeri(cards []Card, jokerRank Rank) bool {
	return validateSeri(cards, jokerRank) == nil
}

// ValidateSeri returns nil when cards form one valid Seri.
func ValidateSeri(cards []Card, jokerRank Rank) error {
	return validateSeri(cards, jokerRank)
}

func IsSeri(cards []Card, jokerRank Rank) bool {
	return IsValidSeri(cards, jokerRank)
}

func CheckSeri(cards []Card, jokerRank Rank) error {
	return ValidateSeri(cards, jokerRank)
}

func validateSeri(cards []Card, jokerRank Rank) error {
	if len(cards) < 3 {
		return ErrTooFewCombinationCards
	}
	if !uniqueCardIDs(cards) {
		return ErrCombinationDuplicateCard
	}

	natural := make([]Card, 0, len(cards))
	for _, card := range cards {
		if IsJoker(card, jokerRank) {
			continue
		}
		natural = append(natural, card)
	}

	// An all-joker combination can represent any consecutive run in one suit.
	if len(natural) == 0 {
		return nil
	}
	for i := 1; i < len(natural); i++ {
		if natural[i].Suit != natural[0].Suit {
			return ErrInvalidSeri
		}
	}

	seenRanks := make(map[Rank]struct{}, len(natural))
	for _, card := range natural {
		if _, exists := seenRanks[card.Rank]; exists {
			return ErrInvalidSeri
		}
		seenRanks[card.Rank] = struct{}{}
	}

	minRank, maxRank := natural[0].Rank, natural[0].Rank
	for _, card := range natural[1:] {
		if card.Rank < minRank {
			minRank = card.Rank
		}
		if card.Rank > maxRank {
			maxRank = card.Rank
		}
	}
	if int(maxRank-minRank) > len(cards)-1 {
		return ErrInvalidSeri
	}
	return nil
}

// IsValidParalel reports whether cards form one valid Paralel.
func IsValidParalel(cards []Card, jokerRank Rank) bool {
	return validateParalel(cards, jokerRank) == nil
}

// ValidateParalel returns nil when cards form one valid Paralel.
func ValidateParalel(cards []Card, jokerRank Rank) error {
	return validateParalel(cards, jokerRank)
}

func IsParalel(cards []Card, jokerRank Rank) bool {
	return IsValidParalel(cards, jokerRank)
}

func CheckParalel(cards []Card, jokerRank Rank) error {
	return ValidateParalel(cards, jokerRank)
}

// Common English spelling aliases.
func IsValidParallel(cards []Card, jokerRank Rank) bool   { return IsValidParalel(cards, jokerRank) }
func ValidateParallel(cards []Card, jokerRank Rank) error { return ValidateParalel(cards, jokerRank) }
func IsParallel(cards []Card, jokerRank Rank) bool        { return IsValidParalel(cards, jokerRank) }
func CheckParallel(cards []Card, jokerRank Rank) error    { return ValidateParalel(cards, jokerRank) }

func validateParalel(cards []Card, jokerRank Rank) error {
	if len(cards) < 3 {
		return ErrTooFewCombinationCards
	}
	// There are only four suits, so a paralel cannot have more cards.
	if len(cards) > 4 {
		return ErrInvalidParalel
	}
	if !uniqueCardIDs(cards) {
		return ErrCombinationDuplicateCard
	}

	natural := make([]Card, 0, len(cards))
	for _, card := range cards {
		if IsJoker(card, jokerRank) {
			continue
		}
		natural = append(natural, card)
	}

	if len(natural) == 0 {
		return nil
	}
	commonRank := natural[0].Rank
	seenSuits := make(map[Suit]struct{}, len(natural))
	for _, card := range natural {
		if card.Rank != commonRank {
			return ErrInvalidParalel
		}
		if _, exists := seenSuits[card.Suit]; exists {
			return ErrInvalidParalel
		}
		seenSuits[card.Suit] = struct{}{}
	}
	return nil
}

// IsValidCombination validates combinations selected by kind.
func IsValidCombination(cards []Card, jokerRank Rank, kind CombinationKind) bool {
	return ValidateCombination(cards, jokerRank, kind) == nil
}

// ValidateCombination validates combinations selected by kind.
func ValidateCombination(cards []Card, jokerRank Rank, kind CombinationKind) error {
	switch kind {
	case CombinationSeri:
		return ValidateSeri(cards, jokerRank)
	case CombinationParalel:
		return ValidateParalel(cards, jokerRank)
	default:
		return ErrInvalidCombination
	}
}

func uniqueCardIDs(cards []Card) bool {
	seen := make(map[any]struct{}, len(cards))
	for _, card := range cards {
		if _, exists := seen[card.ID]; exists {
			return false
		}
		seen[card.ID] = struct{}{}
	}
	return true
}
