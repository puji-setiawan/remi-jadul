package game

import (
	"reflect"
	"testing"
)

func TestTableLockingFirstAndSecondCombination(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	seri := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	paralel := []Card{
		testCard(4, Rank9, SuitSpades),
		testCard(5, Rank9, SuitHearts),
		testCard(6, Rank9, SuitDiamonds),
	}

	next, err := LockCombination(table, seri, jokerRank)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Combinations) != 1 {
		t.Fatalf("after first lock combinations = %d, want 1", len(next.Combinations))
	}
	if len(table.Combinations) != 0 {
		t.Fatal("original table mutated by LockCombination")
	}

	next, err = LockCombination(next, paralel, jokerRank)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Combinations) != 2 {
		t.Fatalf("after second lock combinations = %d, want 2", len(next.Combinations))
	}
}

func TestTableLockingRejectsReuse(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	table, err := LockCombination(table, cards, jokerRank)
	if err != nil {
		t.Fatal(err)
	}

	reuse := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(10, Rank4, SuitHearts),
		testCard(11, Rank5, SuitHearts),
	}
	_, err = LockCombination(table, reuse, jokerRank)
	if err != ErrLockedCardReuse {
		t.Fatalf("err = %v, want %v", err, ErrLockedCardReuse)
	}
	if !table.ContainsCard(1) || !table.HasCard(1) || !IsCardLocked(table, 1) {
		t.Fatal("locked card not reported as locked")
	}
}

func TestTableLockingRejectsInvalidWithoutMutation(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	invalid := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitSpades),
		testCard(3, Rank5, SuitHearts),
	}
	before := table
	_, err := LockCombination(table, invalid, jokerRank)
	if err == nil {
		t.Fatal("expected invalid combination error")
	}
	if !reflect.DeepEqual(table, before) {
		t.Fatal("table mutated on rejected lock")
	}
}

func TestTableLockingDuplicateCardsRejected(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(1, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	_, err := LockCombination(table, cards, jokerRank)
	if err != ErrCombinationDuplicateCard {
		t.Fatalf("err = %v, want %v", err, ErrCombinationDuplicateCard)
	}
}

func TestTableMethodAndAliases(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	if err := table.Lock(cards, jokerRank); err != nil {
		t.Fatal(err)
	}
	if len(table.Combinations) != 1 {
		t.Fatalf("method lock combinations = %d, want 1", len(table.Combinations))
	}
	if table.ContainsLockedCard([]Card{testCard(1, Rank3, SuitHearts)}) != true {
		t.Fatal("contains locked card check failed")
	}
}

func TestTableLockByKindAndSliceHelpers(t *testing.T) {
	jokerRank := Seven
	table := NewTable()
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	next, err := LockCombinationByKind(table, CombinationSeri, cards, jokerRank)
	if err != nil {
		t.Fatal(err)
	}
	slice, err := LockCombinationSlice(next.Combinations, []Card{
		testCard(4, Rank9, SuitSpades),
		testCard(5, Rank9, SuitHearts),
		testCard(6, Rank9, SuitDiamonds),
	}, jokerRank)
	if err != nil {
		t.Fatal(err)
	}
	if len(slice) != 2 {
		t.Fatalf("slice length = %d, want 2", len(slice))
	}
	if err := table.Lock(cards, jokerRank); err != nil {
		t.Fatal(err)
	}
	if LockByKind(&table, CombinationSeri, cards, jokerRank) == nil {
		t.Fatal("expected reuse rejection on second lock")
	}
}

func TestTableAddCombination(t *testing.T) {
	table := NewTable()
	cards := []Card{
		testCard(1, Rank3, SuitHearts),
		testCard(2, Rank4, SuitHearts),
		testCard(3, Rank5, SuitHearts),
	}
	comb := Combination{Kind: CombinationSeri, Cards: cards}
	next, err := AddCombination(table, comb)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Combinations) != 1 {
		t.Fatal("add combination failed")
	}
	// Locked cards reject an overlapping meld.
	if _, err := AddCombination(next, comb); err != ErrLockedCardReuse {
		t.Fatalf("err = %v, want %v", err, ErrLockedCardReuse)
	}
}
