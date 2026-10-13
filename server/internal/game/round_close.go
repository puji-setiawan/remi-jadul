package game

import "errors"

type closeArgs struct {
	closeCard    Card
	hasClose     bool
	combinations [][]Card // groups
	combos       []Combination
}

func parseCloseArgs(args []interface{}) (closeArgs, error) {
	res := closeArgs{}
	var positional []Card
	for _, arg := range args {
		switch v := arg.(type) {
		case Card:
			if !res.hasClose {
				res.closeCard = v
				res.hasClose = true
			} else {
				positional = append(positional, v)
			}
		case []Card:
			if !res.hasClose && len(v) == 1 {
				res.closeCard = v[0]
				res.hasClose = true
			} else {
				res.combinations = append(res.combinations, append([]Card{}, v...))
			}
		case [][]Card:
			res.combinations = append(res.combinations, v...)
		case Combination:
			res.combos = append(res.combos, v)
		case []Combination:
			res.combos = append(res.combos, v...)
		case []interface{}:
			nested, _ := parseCloseArgs(v)
			if nested.hasClose && !res.hasClose {
				res.closeCard = nested.closeCard
				res.hasClose = true
			}
			res.combinations = append(res.combinations, nested.combinations...)
			res.combos = append(res.combos, nested.combos...)
		}
	}
	if !res.hasClose {
		// if positional cards include a close card? assume first positional
		if len(positional) > 0 {
			res.closeCard = positional[0]
			res.hasClose = true
			positional = positional[1:]
		}
	}
	if !res.hasClose {
		return res, errors.New("game: missing close card")
	}
	if len(positional) > 0 {
		res.combinations = append(res.combinations, append([]Card{}, positional...))
	}
	return res, nil
}

func inferCombinationKind(cards []Card, jokerRank Rank) (CombinationKind, bool) {
	if IsValidSeri(cards, jokerRank) && !IsValidParalel(cards, jokerRank) {
		return CombinationSeri, true
	}
	if IsValidParalel(cards, jokerRank) {
		return CombinationParalel, true
	}
	if IsValidSeri(cards, jokerRank) {
		return CombinationSeri, true
	}
	return 0, false
}

func combinationsFromCloseArgs(parsed closeArgs, jokerRank Rank) ([]Combination, error) {
	var combos []Combination
	// explicit combos
	for _, c := range parsed.combos {
		combos = append(combos, c)
	}
	// card groups
	for _, group := range parsed.combinations {
		if len(group) == 0 {
			continue
		}
		kind, ok := inferCombinationKind(group, jokerRank)
		if !ok {
			return nil, ErrInvalidCombination
		}
		combos = append(combos, Combination{Kind: kind, Cards: group})
	}
	return combos, nil
}

func (r *Round) validateClose(player int, args []interface{}) (Card, []Combination, error) {
	if r.Ended {
		return Card{}, nil, ErrRoundEnded
	}
	if player < 0 || player >= r.PlayerCount {
		return Card{}, nil, ErrInvalidPlayer
	}
	if player != r.CurrentPlayer {
		return Card{}, nil, ErrNotCurrentPlayer
	}
	parsed, err := parseCloseArgs(args)
	if err != nil {
		return Card{}, nil, err
	}
	closeCard := parsed.closeCard
	if cardInHand(r.Hands[player], closeCard) < 0 {
		return Card{}, nil, ErrCardNotInHand
	}
	// remaining hand
	remainingHand, ok := removeCardsFromHand(r.Hands[player], []Card{closeCard})
	if !ok {
		return Card{}, nil, ErrCardNotInHand
	}
	combos, err := combinationsFromCloseArgs(parsed, r.JokerRank)
	if err != nil {
		return Card{}, nil, err
	}
	// If explicit combos provided, validate partition exactly.
	if len(combos) > 0 {
		// validate each
		for _, combo := range combos {
			if err := ValidateCombination(combo.Cards, r.JokerRank, combo.Kind); err != nil {
				return Card{}, nil, err
			}
			if !uniqueCardIDs(combo.Cards) {
				return Card{}, nil, ErrCombinationDuplicateCard
			}
		}
		// all remaining cards used exactly once
		used := make(map[any]struct{})
		for _, combo := range combos {
			for _, c := range combo.Cards {
				if _, exists := used[c.ID]; exists {
					return Card{}, nil, ErrCombinationDuplicateCard
				}
				used[c.ID] = struct{}{}
			}
		}
		for _, c := range remainingHand {
			if _, exists := used[c.ID]; !exists {
				return Card{}, nil, ErrInvalidCombination
			}
		}
		if len(used) != len(remainingHand) {
			return Card{}, nil, ErrInvalidCombination
		}
		return closeCard, combos, nil
	}
	// Auto: remaining hand must partition completely
	if len(remainingHand) == 0 {
		return closeCard, []Combination{}, nil
	}
	part := BestPartition(remainingHand, r.JokerRank)
	if len(part.Uncombined) != 0 {
		return Card{}, nil, ErrInvalidCombination
	}
	return closeCard, part.Combinations, nil
}

func closeImpl(r *Round, player int, args []interface{}) error {
	closeCard, combos, err := r.validateClose(player, args)
	if err != nil {
		return err
	}
	next := r.cloneRound()
	// remove close card and combo cards from hand
	hand := clonesCardsForRound(r.Hands[player])
	hand, _ = removeCardsFromHand(hand, []Card{closeCard})
	allComboCards := []Card{}
	for _, combo := range combos {
		allComboCards = append(allComboCards, combo.Cards...)
	}
	hand, _ = removeCardsFromHand(hand, allComboCards)
	next.Hands[player] = hand
	// lock combos into table with owner
	combosCopy := make([]LockedCombination, len(combos))
	for i, combo := range combos {
		cards := make([]Card, len(combo.Cards))
		copy(cards, combo.Cards)
		combosCopy[i] = LockedCombination{Kind: combo.Kind, Cards: cards, Owner: player}
	}
	tableCombos := append([]LockedCombination{}, next.Table.Combinations...)
	tableCombos = append(tableCombos, combosCopy...)
	next.Table = Table{Combinations: tableCombos}
	// ended
	next.Ended = true
	next.Status = "ended"
	next.EndReason = "close"
	next.Reason = next.EndReason
	next.Closed = true
	next.HasClosed = true
	next.ClosePlayer = player
	next.Closer = player
	next.ClosedBy = player
	cc := closeCard
	next.Scores = next.scorePlayers(player, &cc, false, -1)
	next.RoundScores = cloneInts(next.Scores)
	next.PlayerScores = cloneInts(next.Scores)
	next.HasDrawn = false
	*r = next
	return nil
}

func (r *Round) Close(player int, args ...interface{}) error {
	return closeImpl(r, player, args)
}
func (r *Round) CloseRound(player int, args ...interface{}) error {
	return closeImpl(r, player, args)
}
func (r *Round) CloseHand(player int, args ...interface{}) error {
	return closeImpl(r, player, args)
}
func (r *Round) TryClose(player int, args ...interface{}) error {
	return closeImpl(r, player, args)
}
func (r *Round) ValidateClose(player int, args ...interface{}) error {
	_, _, err := r.validateClose(player, args)
	return err
}
func (r *Round) CanClose(player int, args ...interface{}) bool {
	err := r.ValidateClose(player, args)
	return err == nil
}
func (r *Round) IsClosable(player int, args ...interface{}) bool { return r.CanClose(player, args) }
func (r *Round) CheckClose(player int, args ...interface{}) error {
	return r.ValidateClose(player, args)
}

// Free functions / helpers for close
func ValidateCloseCards(hand []Card, closeCard Card, jokerRank Rank) error {
	if cardInHand(hand, closeCard) < 0 {
		return ErrCardNotInHand
	}
	remaining, _ := removeCardsFromHand(hand, []Card{closeCard})
	if len(remaining) == 0 {
		return nil
	}
	part := BestPartition(remaining, jokerRank)
	if len(part.Uncombined) != 0 {
		return ErrInvalidCombination
	}
	return nil
}
func CanCloseCards(hand []Card, closeCard Card, jokerRank Rank) bool {
	return ValidateCloseCards(hand, closeCard, jokerRank) == nil
}
func CanCloseHand(hand []Card, closeCard Card, jokerRank Rank) bool {
	return CanCloseCards(hand, closeCard, jokerRank)
}

func CanClose(hand []Card, closeCard Card, jokerRank Rank) bool {
	return CanCloseCards(hand, closeCard, jokerRank)
}
func CloseValid(hand []Card, closeCard Card, jokerRank Rank) bool {
	return CanCloseCards(hand, closeCard, jokerRank)
}
func ValidateClose(hand []Card, closeCard Card, jokerRank Rank) error {
	return ValidateCloseCards(hand, closeCard, jokerRank)
}
