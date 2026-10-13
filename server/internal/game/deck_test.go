package game

import (
	"reflect"
	"testing"
)

func TestStandardDeckHasFiftyTwoUniqueCards(t *testing.T) {
	deck := NewStandardDeck()
	if len(deck) != 52 {
		t.Fatalf("len(deck) = %d, want 52", len(deck))
	}
	seen := make(map[any]struct{}, len(deck))
	for _, card := range deck {
		if _, exists := seen[card.ID]; exists {
			t.Fatalf("duplicate physical id %v", card.ID)
		}
		seen[card.ID] = struct{}{}
	}
	combos := make(map[string]struct{}, len(deck))
	for _, card := range deck {
		key := card.Rank.String() + card.Suit.String()
		if _, exists := combos[key]; exists {
			t.Fatalf("duplicate card combo %s", key)
		}
		combos[key] = struct{}{}
	}
}

func TestStandardDeckDeterministic(t *testing.T) {
	first := NewStandardDeck()
	second := NewStandardDeck()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("canonical deck is not deterministic")
	}
}

func TestDealValidPlayerCounts(t *testing.T) {
	cases := []struct {
		players   int
		wantHands int
		wantCards int
	}{
		{players: 2, wantHands: 2, wantCards: 11},
		{players: 3, wantHands: 3, wantCards: 9},
		{players: 4, wantHands: 4, wantCards: 7},
	}
	for _, tc := range cases {
		deck := NewStandardDeck()
		before := cloneCards(deck)
		result, err := Deal(deck, tc.players)
		if err != nil {
			t.Fatalf("players=%d: unexpected error: %v", tc.players, err)
		}
		if len(result.Hands) != tc.wantHands {
			t.Fatalf("players=%d: len(Hands) = %d, want %d", tc.players, len(result.Hands), tc.wantHands)
		}
		for i, hand := range result.Hands {
			if len(hand) != tc.wantCards {
				t.Fatalf("players=%d hand %d: len = %d, want %d", tc.players, i, len(hand), tc.wantCards)
			}
		}
		if !reflect.DeepEqual(deck, before) {
			t.Fatalf("players=%d: supplied deck was mutated", tc.players)
		}
	}
}

func TestDealInvalidPlayerCounts(t *testing.T) {
	for _, players := range []int{0, 1, 5, -1, 6, 10} {
		deck := NewStandardDeck()
		before := cloneCards(deck)
		result, err := Deal(deck, players)
		if err != ErrInvalidPlayerCount {
			t.Fatalf("players=%d: err = %v, want %v", players, err, ErrInvalidPlayerCount)
		}
		if len(result.Hands) != 0 {
			t.Fatalf("players=%d: partial hands returned: %d", players, len(result.Hands))
		}
		if !reflect.DeepEqual(deck, before) {
			t.Fatalf("players=%d: supplied deck was mutated", players)
		}
	}
}

func TestDealNoCardInMoreThanOneHand(t *testing.T) {
	for _, players := range []int{2, 3, 4} {
		result, err := Deal(NewStandardDeck(), players)
		if err != nil {
			t.Fatal(err)
		}
		seen := make(map[any]struct{})
		for _, hand := range result.Hands {
			for _, card := range hand {
				if _, exists := seen[card.ID]; exists {
					t.Fatalf("players=%d: duplicate card id %v across hands", players, card.ID)
				}
				seen[card.ID] = struct{}{}
			}
		}
	}
}

func TestDealMarkerExcludedAndJokerRankSet(t *testing.T) {
	for _, players := range []int{2, 3, 4} {
		deck := NewStandardDeck()
		result, err := Deal(deck, players)
		if err != nil {
			t.Fatal(err)
		}
		handSize := playerHandSize[players]
		dealt := players * handSize
		marker := deck[dealt]
		if result.JokerRank != marker.Rank {
			t.Fatalf("joker rank = %v, want marker rank %v", result.JokerRank, marker.Rank)
		}
		if result.ActiveJokerRank != result.JokerRank {
			t.Fatalf("active joker rank = %v, want %v", result.ActiveJokerRank, result.JokerRank)
		}

		// The marker itself must be absent from hands and remaining cards.
		missing := map[any]struct{}{}
		present := make(map[any]struct{})
		for _, hand := range result.Hands {
			for _, card := range hand {
				present[card.ID] = struct{}{}
			}
		}
		for _, card := range result.Remaining {
			present[card.ID] = struct{}{}
		}
		for _, card := range deck {
			if _, ok := present[card.ID]; !ok {
				missing[card.ID] = struct{}{}
			}
		}
		if len(missing) != 1 {
			t.Fatalf("players=%d: expected exactly one missing card, got %d", players, len(missing))
		}
		for id := range missing {
			if id != marker.ID {
				t.Fatalf("players=%d: missing card id %v, want marker %v", players, id, marker.ID)
			}
		}

		// Three cards of the joker rank remain in play.
		count := 0
		for _, hand := range result.Hands {
			for _, card := range hand {
				if card.Rank == result.JokerRank {
					count++
				}
			}
		}
		for _, card := range result.Remaining {
			if card.Rank == result.JokerRank {
				count++
			}
		}
		if count != 3 {
			t.Fatalf("players=%d: joker-rank cards in play = %d, want 3", players, count)
		}
	}
}

func TestDealDuplicateCardInputRejected(t *testing.T) {
	deck := append(cloneCards(NewStandardDeck()[:5]), NewStandardDeck()[0])
	before := cloneCards(deck)
	_, err := Deal(deck, 2)
	if err != ErrDuplicateCard {
		t.Fatalf("err = %v, want %v", err, ErrDuplicateCard)
	}
	if !reflect.DeepEqual(deck, before) {
		t.Fatal("supplied deck was mutated")
	}
}

func TestDealInsufficientCardsRejected(t *testing.T) {
	deck := NewStandardDeck()[:10]
	_, err := Deal(deck, 2)
	if err != ErrInsufficientCards {
		t.Fatalf("err = %v, want %v", err, ErrInsufficientCards)
	}
}

func TestDealRemainingCardCount(t *testing.T) {
	for _, players := range []int{2, 3, 4} {
		handSize := playerHandSize[players]
		result, err := Deal(NewStandardDeck(), players)
		if err != nil {
			t.Fatal(err)
		}
		wantRemaining := 52 - players*handSize - 1
		if len(result.Remaining) != wantRemaining {
			t.Fatalf("players=%d: remaining = %d, want %d", players, len(result.Remaining), wantRemaining)
		}
	}
}
