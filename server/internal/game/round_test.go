package game

import (
	"reflect"
	"testing"
)

func TestRoundDrawAndDiscardAdvancesTurn(t *testing.T) {
	hands := [][]Card{{{ID: 1, Rank: Rank3, Suit: SuitHearts}}, {{ID: 2, Rank: Rank4, Suit: SuitHearts}}}
	closed := []Card{{ID: 9, Rank: Rank5, Suit: SuitSpades}}
	round, err := NewRound(2, hands, closed, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if round.Current() != 0 || round.HasDrawn {
		t.Fatal("unexpected initial state")
	}
	if err := round.Draw(0); err != nil {
		t.Fatal(err)
	}
	if !round.HasDrawn || len(round.Hands[0]) != 2 || len(round.ClosedDeck) != 0 {
		t.Fatalf("draw state wrong: hands=%d closed=%d drawn=%v", len(round.Hands[0]), len(round.ClosedDeck), round.HasDrawn)
	}
	discardCard := round.Hands[0][1]
	if err := round.Discard(0, discardCard); err != nil {
		t.Fatal(err)
	}
	if round.HasDrawn || round.Current() != 1 || len(round.DiscardPile) != 1 {
		t.Fatalf("discard state wrong: drawn=%v current=%d discard=%d", round.HasDrawn, round.Current(), len(round.DiscardPile))
	}
}

func TestRoundWrongPlayerRejected(t *testing.T) {
	round, err := NewRound(2, [][]Card{{{ID: 1}}, {{ID: 2}}}, []Card{{ID: 9}}, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if err := round.Draw(1); err != ErrNotCurrentPlayer {
		t.Fatalf("draw err = %v, want %v", err, ErrNotCurrentPlayer)
	}
	if err := round.Draw(0); err != nil {
		t.Fatal(err)
	}
	if err := round.Discard(1, Card{ID: 9}); err != ErrNotCurrentPlayer {
		t.Fatalf("discard err = %v, want %v", err, ErrNotCurrentPlayer)
	}
}

func TestRoundDuplicateDrawRejected(t *testing.T) {
	round, err := NewRound(2, [][]Card{{{ID: 1}}, {{ID: 2}}}, []Card{{ID: 9}, {ID: 10}}, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if err := round.Draw(0); err != nil {
		t.Fatal(err)
	}
	if err := round.Draw(0); err != ErrAlreadyDrawn {
		t.Fatalf("err = %v, want %v", err, ErrAlreadyDrawn)
	}
}

func TestRoundDiscardRequiresDraw(t *testing.T) {
	round, err := NewRound(2, [][]Card{{{ID: 1}}, {{ID: 2}}}, []Card{{ID: 9}}, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if err := round.Discard(0, Card{ID: 1}); err != ErrDrawRequired {
		t.Fatalf("err = %v, want %v", err, ErrDrawRequired)
	}
}

func TestRoundDiscardMissingCardRejected(t *testing.T) {
	round, err := NewRound(2, [][]Card{{{ID: 1}}, {{ID: 2}}}, []Card{{ID: 9}}, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if err := round.Draw(0); err != nil {
		t.Fatal(err)
	}
	beforeDeck := clonesCardsForRound(round.ClosedDeck)
	beforeHand := cloneHandsForRound(round.Hands)
	beforeDiscard := cloneHandsForRound([][]Card{round.DiscardPile}) // keeps shape
	_ = beforeDeck
	_ = beforeHand
	_ = beforeDiscard
	if err := round.Discard(0, Card{ID: 999}); err != ErrCardNotInHand {
		t.Fatalf("err = %v, want %v", err, ErrCardNotInHand)
	}
}

func TestRoundEmptyDeckDrawRejected(t *testing.T) {
	round, err := NewRound(2, [][]Card{{{ID: 1}}, {{ID: 2}}}, nil, Seven)
	if err != nil {
		t.Fatal(err)
	}
	if err := round.Draw(0); err != ErrClosedDeckEmpty {
		t.Fatalf("err = %v, want %v", err, ErrClosedDeckEmpty)
	}
}

func TestRoundImmutabilityOnRejectedDraw(t *testing.T) {
	hands := [][]Card{{{ID: 1}}, {{ID: 2}}}
	closed := []Card{{ID: 9}}
	round, err := NewRound(2, hands, closed, Seven)
	if err != nil {
		t.Fatal(err)
	}
	before := *round
	before.Hands = cloneHandsForRound(before.Hands)
	before.ClosedDeck = clonesCardsForRound(before.ClosedDeck)
	before.DiscardPile = clonesCardsForRound(before.DiscardPile)
	_ = round.Draw(1)
	if !reflect.DeepEqual(*round, before) {
		t.Fatal("rejected draw mutated round")
	}
}

func TestRoundConstructorPlayerCountValidation(t *testing.T) {
	for _, players := range []int{0, 1, 5} {
		if _, err := NewRound(players, nil, nil, Seven); err != ErrInvalidPlayerCount {
			t.Fatalf("players=%d err=%v, want %v", players, err, ErrInvalidPlayerCount)
		}
	}
	// Inferred player count form.
	round, err := NewRound([][]Card{{{ID: 1}}, {{ID: 2}}}, []Card{{ID: 9}}, Seven)
	if err != nil || round.PlayerCount != 2 {
		t.Fatalf("inferred constructor failed: %v count=%d", err, round.PlayerCount)
	}
}
