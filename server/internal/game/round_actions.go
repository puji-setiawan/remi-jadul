package game

func cardsFromArgs(args []interface{}) []Card {
	var out []Card
	for _, arg := range args {
		switch v := arg.(type) {
		case Card:
			out = append(out, v)
		case []Card:
			out = append(out, v...)
		case []interface{}:
			out = append(out, cardsFromArgs(v)...)
		case [][]Card:
			for _, group := range v {
				out = append(out, group...)
			}
		}
	}
	return out
}

func cardInHand(hand []Card, card Card) int {
	for i, c := range hand {
		if c.ID == card.ID {
			return i
		}
	}
	return -1
}

func removeCardsFromHand(hand []Card, cards []Card) ([]Card, bool) {
	hand = clonesCardsForRound(hand)
	for _, target := range cards {
		idx := -1
		for i, c := range hand {
			if c.ID == target.ID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, false
		}
		hand = append(hand[:idx], hand[idx+1:]...)
	}
	return hand, true
}

func cardIDsUnique(cards []Card) bool { return uniqueCardIDs(cards) }

func handContainsAll(hand []Card, cards []Card) bool {
	for _, target := range cards {
		if cardInHand(hand, target) < 0 {
			return false
		}
	}
	return true
}

func layCombinationImpl(r *Round, player int, args []interface{}) error {
	if r.Ended {
		return ErrRoundEnded
	}
	if player < 0 || player >= r.PlayerCount {
		return ErrInvalidPlayer
	}
	if player != r.CurrentPlayer {
		return ErrNotCurrentPlayer
	}
	if !r.HasDrawn {
		return ErrDrawRequired
	}
	cards := cardsFromArgs(args)
	if len(cards) < 3 {
		return ErrTooFewCombinationCards
	}
	if !uniqueCardIDs(cards) {
		return ErrCombinationDuplicateCard
	}
	if !handContainsAll(r.Hands[player], cards) {
		return ErrCardNotInHand
	}
	var kind CombinationKind
	if IsValidSeri(cards, r.JokerRank) {
		kind = CombinationSeri
	} else if IsValidParalel(cards, r.JokerRank) {
		kind = CombinationParalel
	} else {
		return ErrInvalidCombination
	}
	next := r.cloneRound()
	hand, ok := removeCardsFromHand(r.Hands[player], cards)
	if !ok {
		return ErrCardNotInHand
	}
	next.Hands[player] = hand
	combos := append([]LockedCombination{}, next.Table.Combinations...)
	cardCopy := make([]Card, len(cards))
	copy(cardCopy, cards)
	combos = append(combos, LockedCombination{Kind: kind, Cards: cardCopy, Owner: player})
	next.Table = Table{Combinations: combos}
	*r = next
	return nil
}

func (r *Round) LayCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayDown(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) Lay(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayMeld(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LockCombinationFromHand(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) Meld(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) PlaceCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) AddCombinationFromHand(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
