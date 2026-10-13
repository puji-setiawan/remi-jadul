package game

import "errors"

// Table locking and card reuse errors.
var (
	ErrLockedCardReuse          = errors.New("game: locked card reused")
	ErrInvalidLockedCombination = errors.New("game: invalid combination for table locking")
)

// Combination represents one proposed or locked seri/paralel.
type Combination struct {
	Kind  CombinationKind
	Cards []Card
}

// LockedCombination, TableCombination, and TableMeld are aliases for public
// table-combination data.
type LockedCombination = Combination
type TableCombination = Combination
type TableMeld = Combination

// Table holds locked combinations. Once a combination is added, its cards are
// permanently unavailable for another combination or table mutation.
type Table struct {
	Combinations []LockedCombination
}

func NewTable() Table {
	return Table{Combinations: []LockedCombination{}}
}

func EmptyTable() Table {
	return NewTable()
}

// NewLockedTable is an alias for NewTable.
func NewLockedTable() Table { return NewTable() }

// ContainsCard reports whether id is present in a locked combination.
func (t Table) ContainsCard(id any) bool {
	for _, combination := range t.Combinations {
		for _, card := range combination.Cards {
			if card.ID == id {
				return true
			}
		}
	}
	return false
}

// HasCard is an alias for ContainsCard.
func (t Table) HasCard(id any) bool { return t.ContainsCard(id) }

// IsCardLocked is a free-function form of ContainsCard.
func IsCardLocked(t Table, id any) bool { return t.ContainsCard(id) }

// CardLocked is an alias for IsCardLocked.
func CardLocked(t Table, id any) bool { return IsCardLocked(t, id) }

// ContainsLockedCard reports whether any card in cards is already locked.
func (t Table) ContainsLockedCard(cards []Card) bool {
	for _, card := range cards {
		if t.ContainsCard(card.ID) {
			return true
		}
	}
	return false
}

// HasLockedCard is an alias for ContainsLockedCard.
func (t Table) HasLockedCard(cards []Card) bool { return t.ContainsLockedCard(cards) }

// ReusesLockedCard is a free-function form of ContainsLockedCard.
func ReusesLockedCard(t Table, cards []Card) bool {
	return t.ContainsLockedCard(cards)
}

// LockedCardIDs returns the physical IDs currently locked in table order.
func (t Table) LockedCardIDs() []any {
	ids := make([]any, 0)
	for _, combination := range t.Combinations {
		for _, card := range combination.Cards {
			ids = append(ids, card.ID)
		}
	}
	return ids
}

// CanLockCombination validates a combination and ensures none of its cards are
// already locked. It does not mutate table.
func CanLockCombination(t Table, cards []Card, jokerRank Rank) error {
	if containsDuplicateCard(cards) {
		return ErrCombinationDuplicateCard
	}
	if _, err := validLockedCombinationKind(cards, jokerRank); err != nil {
		return err
	}
	if t.ContainsLockedCard(cards) {
		return ErrLockedCardReuse
	}
	return nil
}

// LockCombination returns a new immutable table with the combination locked.
func LockCombination(t Table, cards []Card, jokerRank Rank) (Table, error) {
	if err := CanLockCombination(t, cards, jokerRank); err != nil {
		return t, err
	}
	kind, err := validLockedCombinationKind(cards, jokerRank)
	if err != nil {
		return t, err
	}
	next := Table{Combinations: make([]LockedCombination, len(t.Combinations), len(t.Combinations)+1)}
	copy(next.Combinations, t.Combinations)
	cardsCopy := make([]Card, len(cards))
	copy(cardsCopy, cards)
	next.Combinations = append(next.Combinations, LockedCombination{
		Kind:  kind,
		Cards: cardsCopy,
	})
	return next, nil
}

// Lock validates and locks a combination on the same table value.
func (t *Table) Lock(cards []Card, jokerRank Rank) error {
	next, err := LockCombination(*t, cards, jokerRank)
	if err != nil {
		return err
	}
	*t = next
	return nil
}

// LockCombinationByKind validates and locks a combination whose kind is given by
// the caller. The kind must agree with actual validation, except all-joker
// combinations may be recorded as the requested valid kind.
func LockCombinationByKind(t Table, kind CombinationKind, cards []Card, jokerRank Rank) (Table, error) {
	if containsDuplicateCard(cards) {
		return t, ErrCombinationDuplicateCard
	}
	if err := validateCombinationForKind(kind, cards, jokerRank); err != nil {
		return t, err
	}
	if t.ContainsLockedCard(cards) {
		return t, ErrLockedCardReuse
	}
	next := Table{Combinations: make([]LockedCombination, len(t.Combinations), len(t.Combinations)+1)}
	copy(next.Combinations, t.Combinations)
	cardsCopy := make([]Card, len(cards))
	copy(cardsCopy, cards)
	next.Combinations = append(next.Combinations, LockedCombination{
		Kind:  kind,
		Cards: cardsCopy,
	})
	return next, nil
}

// LockByKind locks a combination on t using the supplied kind.
func (t *Table) LockByKind(kind CombinationKind, cards []Card, jokerRank Rank) error {
	next, err := LockCombinationByKind(*t, kind, cards, jokerRank)
	if err != nil {
		return err
	}
	*t = next
	return nil
}

// LockByKindFunc provides package-level locking for callers that prefer an explicit table pointer.
func LockByKind(t *Table, kind CombinationKind, cards []Card, jokerRank Rank) error {
	if t == nil {
		return ErrInvalidLockedCombination
	}
	next, err := LockCombinationByKind(*t, kind, cards, jokerRank)
	if err != nil {
		return err
	}
	*t = next
	return nil
}

// AddCombination appends an already validated combination to a copy of table.
func AddCombination(t Table, combination Combination) (Table, error) {
	if len(combination.Cards) == 0 {
		return t, ErrInvalidLockedCombination
	}
	if combination.Kind != CombinationSeri && combination.Kind != CombinationParalel {
		return t, ErrInvalidLockedCombination
	}
	if containsDuplicateCard(combination.Cards) {
		return t, ErrCombinationDuplicateCard
	}
	if t.ContainsLockedCard(combination.Cards) {
		return t, ErrLockedCardReuse
	}
	next := Table{Combinations: make([]LockedCombination, len(t.Combinations), len(t.Combinations)+1)}
	copy(next.Combinations, t.Combinations)
	cardsCopy := make([]Card, len(combination.Cards))
	copy(cardsCopy, combination.Cards)
	combCopy := combination
	combCopy.Cards = cardsCopy
	next.Combinations = append(next.Combinations, combCopy)
	return next, nil
}

// AppendLockedCombination is an alias for AddCombination.
func AppendLockedCombination(t Table, combination Combination) (Table, error) {
	return AddCombination(t, combination)
}

// Table-level free-function aliases for callers that prefer slice results.
func LockCombinationSlice(existing []Combination, cards []Card, jokerRank Rank) ([]Combination, error) {
	table := Table{Combinations: existing}
	next, err := LockCombination(table, cards, jokerRank)
	if err != nil {
		return existing, err
	}
	return append([]Combination(nil), next.Combinations...), nil
}

func lockedCombinationKindOrInvalid(cards []Card, jokerRank Rank) (CombinationKind, error) {
	if IsValidSeri(cards, jokerRank) && IsValidParalel(cards, jokerRank) {
		return CombinationSeri, nil
	}
	if IsValidSeri(cards, jokerRank) {
		return CombinationSeri, nil
	}
	if IsValidParalel(cards, jokerRank) {
		return CombinationParalel, nil
	}
	return 0, ErrInvalidLockedCombination
}

func validLockedCombinationKind(cards []Card, jokerRank Rank) (CombinationKind, error) {
	return lockedCombinationKindOrInvalid(cards, jokerRank)
}

func validateCombinationForKind(kind CombinationKind, cards []Card, jokerRank Rank) error {
	switch kind {
	case CombinationSeri:
		return ValidateSeri(cards, jokerRank)
	case CombinationParalel:
		return ValidateParalel(cards, jokerRank)
	default:
		return ErrInvalidCombination
	}
}

func containsDuplicateCard(cards []Card) bool {
	return !uniqueCardIDs(cards)
}
