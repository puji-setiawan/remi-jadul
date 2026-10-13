package game

import "testing"

func TestScoreHandSimpleCombination(t *testing.T) {
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
		testCard(4, RankKing, SuitSpades),
	}
	if score := ScoreHand(cards, Seven); score != 1 {
		t.Fatalf("score = %d, want 1", score)
	}
	part := BestPartition(cards, Seven)
	if len(part.Combinations) != 1 || len(part.Uncombined) != 1 {
		t.Fatalf("partition wrong: combos=%d uncombined=%d", len(part.Combinations), len(part.Uncombined))
	}
}

func TestScoreHandJokerRepresentsMissingRank(t *testing.T) {
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank5, SuitHearts),
		testCard(3, Seven, SuitSpades),
	}
	if score := ScoreHand(cards, Seven); score != 3 {
		t.Fatalf("score = %d, want 3", score)
	}
}

func TestScoreHandUncombinedJoker(t *testing.T) {
	cards := []Card{{ID: 1, Rank: Seven, Suit: Spades}}
	if score := ScoreHand(cards, Seven); score != -1 {
		t.Fatalf("score = %d, want -1", score)
	}
}

func TestScoreHandWithClose(t *testing.T) {
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	closeCard := testCard(9, RankKing, SuitSpades)
	if score := ScoreHandWithClose(cards, closeCard, Seven); score != 13 {
		t.Fatalf("score = %d, want 13", score)
	}
}

func TestScoringPrefersHighestCombination(t *testing.T) {
	// A valid seri exists and should be selected rather than left uncombined.
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	if score := ScoreHand(cards, Seven); score != 3 {
		t.Fatalf("score = %d, want 3", score)
	}
}

func TestCombinationCardValueHelpers(t *testing.T) {
	if got := CombinationCardValue(testCard(1, RankAce, SuitSpades), Seven); got != 3 {
		t.Fatalf("ace combination value = %d, want 3", got)
	}
	if got := UncombinedCardValue(testCard(1, RankAce, SuitSpades), Seven); got != -3 {
		t.Fatalf("ace uncombined value = %d, want -3", got)
	}
	if got := CloseCardValue(testCard(1, RankQueen, SuitSpades)); got != 10 {
		t.Fatalf("queen close value = %d, want 10", got)
	}
}
