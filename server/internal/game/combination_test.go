package game

import (
	"reflect"
	"testing"
)

func testCard(id int, rank Rank, suit Suit) Card {
	return Card{ID: id, Rank: rank, Suit: suit}
}

func TestSeriValidation(t *testing.T) {
	jokerRank := Seven
	cases := []struct {
		name    string
		cards   []Card
		wantErr error
	}{
		{
			name: "valid consecutive natural seri",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(2, Rank4, SuitHearts),
				testCard(3, Rank5, SuitHearts),
			},
			wantErr: nil,
		},
		{
			name: "joker fills one missing rank",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(2, Rank5, SuitHearts),
				testCard(3, Seven, SuitSpades),
			},
			wantErr: nil,
		},
		{
			name:    "all jokers",
			cards:   []Card{testCard(1, Seven, SuitSpades), testCard(2, Seven, SuitHearts), testCard(3, Seven, SuitDiamonds)},
			wantErr: nil,
		},
		{
			name: "minimum three with endpoint joker",
			cards: []Card{
				testCard(1, Four, SuitHearts),
				testCard(2, Five, SuitHearts),
				testCard(3, Seven, SuitClubs),
			},
			wantErr: nil,
		},
		{
			name: "mixed suits",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(2, Rank4, SuitSpades),
				testCard(3, Rank5, SuitHearts),
			},
			wantErr: ErrInvalidSeri,
		},
		{
			name: "gap without enough joker",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(2, Rank5, SuitHearts),
				testCard(3, Rank6, SuitHearts),
			},
			wantErr: ErrInvalidSeri,
		},
		{
			name: "too few cards",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(2, Rank4, SuitHearts),
			},
			wantErr: ErrTooFewCombinationCards,
		},
		{
			name: "duplicate physical card",
			cards: []Card{
				testCard(1, Rank3, SuitHearts),
				testCard(1, Rank4, SuitHearts),
				testCard(3, Rank5, SuitHearts),
			},
			wantErr: ErrCombinationDuplicateCard,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSeri(tc.cards, jokerRank)
			if err != tc.wantErr {
				t.Fatalf("ValidateSeri() = %v, want %v", err, tc.wantErr)
			}
			if (err == nil) != IsValidSeri(tc.cards, jokerRank) {
				t.Fatal("bool helper disagrees with error helper")
			}
		})
	}
}

func TestParalelValidation(t *testing.T) {
	jokerRank := Seven
	cases := []struct {
		name    string
		cards   []Card
		wantErr error
	}{
		{
			name: "valid three distinct suits",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
				testCard(3, Rank9, SuitDiamonds),
			},
			wantErr: nil,
		},
		{
			name: "joker represents missing suit",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
				testCard(3, Seven, SuitClubs),
			},
			wantErr: nil,
		},
		{
			name:    "all jokers",
			cards:   []Card{testCard(1, Seven, SuitSpades), testCard(2, Seven, SuitHearts), testCard(3, Seven, SuitDiamonds)},
			wantErr: nil,
		},
		{
			name: "four natural suits",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
				testCard(3, Rank9, SuitDiamonds),
				testCard(4, Rank9, SuitClubs),
			},
			wantErr: nil,
		},
		{
			name: "duplicate suit",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitSpades),
				testCard(3, Rank9, SuitHearts),
			},
			wantErr: ErrInvalidParalel,
		},
		{
			name: "different rank",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
				testCard(3, Rank10, SuitDiamonds),
			},
			wantErr: ErrInvalidParalel,
		},
		{
			name: "too many cards for four suits",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
				testCard(3, Rank9, SuitDiamonds),
				testCard(4, Rank9, SuitClubs),
				testCard(5, Seven, SuitSpades),
			},
			wantErr: ErrInvalidParalel,
		},
		{
			name: "too few cards",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(2, Rank9, SuitHearts),
			},
			wantErr: ErrTooFewCombinationCards,
		},
		{
			name: "duplicate physical card",
			cards: []Card{
				testCard(1, Rank9, SuitSpades),
				testCard(1, Rank9, SuitHearts),
				testCard(3, Rank9, SuitDiamonds),
			},
			wantErr: ErrCombinationDuplicateCard,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParalel(tc.cards, jokerRank)
			if err != tc.wantErr {
				t.Fatalf("ValidateParalel() = %v, want %v", err, tc.wantErr)
			}
			if (err == nil) != IsValidParalel(tc.cards, jokerRank) {
				t.Fatal("bool helper disagrees with error helper")
			}
			// Common English spelling alias must behave identically.
			if ValidateParallel(tc.cards, jokerRank) != ValidateParalel(tc.cards, jokerRank) {
				t.Fatal("parallel alias differs")
			}
		})
	}
}

func TestCombinationKindDispatcher(t *testing.T) {
	jokerRank := Seven
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	if !IsValidCombination(cards, jokerRank, CombinationSeri) {
		t.Fatal("dispatcher rejected valid seri")
	}
	if ValidateCombination(cards, jokerRank, CombinationParalel) == nil {
		t.Fatal("dispatcher accepted seri as paralel")
	}
	if ValidateCombination(cards, jokerRank, CombinationKind(999)) == nil {
		t.Fatal("dispatcher accepted unknown kind")
	}
}

func TestIsJokerByRank(t *testing.T) {
	jokerRank := Seven
	if !IsJoker(testCard(1, Seven, SuitSpades), jokerRank) {
		t.Fatal("active joker not detected")
	}
	if IsJoker(testCard(2, Rank3, SuitHearts), jokerRank) {
		t.Fatal("natural card misclassified as joker")
	}
}

func TestCombinationValidationDoesNotDependOnMapOrder(t *testing.T) {
	jokerRank := Seven
	cards := []Card{
		testCard(3, Rank5, SuitHearts),
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
	}
	before := make([]Card, len(cards))
	copy(before, cards)
	if !IsValidSeri(cards, jokerRank) {
		t.Fatal("valid seri order rejected")
	}
	if !reflect.DeepEqual(cards, before) {
		t.Fatal("input mutated")
	}
}
