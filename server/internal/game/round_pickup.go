package game

import "errors"

var ErrInvalidJago = errors.New("game: invalid jago")
var ErrInvalidDiscardWindow = errors.New("game: discard outside eligible window")
var ErrPickupSuffix = errors.New("game: invalid pickup suffix")
var ErrInvalidExtension = errors.New("game: invalid extension card")

type pickupArgs struct {
	discard    Card
	basis      []Card
	extensions []Card
	hasDiscard bool
}

func parsePickupArgs(args []interface{}) (pickupArgs, error) {
	res := pickupArgs{}
	var leadingCards []Card
	var slices [][]Card
	for _, arg := range args {
		switch v := arg.(type) {
		case Card:
			leadingCards = append(leadingCards, v)
		case []Card:
			slices = append(slices, append([]Card{}, v...))
		case []interface{}:
			nested, _ := parsePickupArgs(v)
			if nested.hasDiscard {
				if !res.hasDiscard {
					res.discard = nested.discard
					res.hasDiscard = true
				}
				res.basis = append(res.basis, nested.basis...)
				res.extensions = append(res.extensions, nested.extensions...)
			} else {
				leadingCards = append(leadingCards, cardsFromArgs(v)...)
			}
		}
	}
	// discard
	if !res.hasDiscard {
		if len(leadingCards) > 0 {
			res.discard = leadingCards[0]
			res.hasDiscard = true
			leadingCards = leadingCards[1:]
		} else if len(slices) > 0 && len(slices[0]) == 1 {
			res.discard = slices[0][0]
			res.hasDiscard = true
			slices = slices[1:]
		}
	}
	// basis
	if len(res.basis) == 0 {
		if len(leadingCards) >= 2 {
			res.basis = append(res.basis, leadingCards[0], leadingCards[1])
			leadingCards = leadingCards[2:]
		} else if len(slices) > 0 && len(slices[0]) >= 2 {
			res.basis = append(res.basis, slices[0]...)
			slices = slices[1:]
		} else if len(leadingCards) == 1 {
			res.basis = append(res.basis, leadingCards[0])
			leadingCards = leadingCards[1:]
		}
	}
	// extensions
	res.extensions = append(res.extensions, leadingCards...)
	for _, sl := range slices {
		res.extensions = append(res.extensions, sl...)
	}

	if !res.hasDiscard {
		return res, errors.New("game: missing discard card")
	}
	if len(res.basis) != 2 {
		return res, errors.New("game: jago requires two basis cards")
	}
	return res, nil
}

func (r *Round) eligibleDiscardIndex(discard Card) (int, bool) {
	for i, c := range r.DiscardPile {
		if c.ID == discard.ID {
			// eligible if position from bottom (i) < N
			if i < r.startingHandSizeOrZero() {
				return i, true
			}
			return i, false
		}
	}
	return -1, false
}

func pickupImpl(r *Round, player int, args []interface{}) error {
	if r.Ended {
		return ErrRoundEnded
	}
	if player < 0 || player >= r.PlayerCount {
		return ErrInvalidPlayer
	}
	if player != r.CurrentPlayer {
		return ErrNotCurrentPlayer
	}
	if r.HasDrawn {
		return ErrAlreadyDrawn
	}
	if len(r.DiscardPile) == 0 {
		return ErrCardNotInHand
	}
	parsed, err := parsePickupArgs(args)
	if err != nil {
		return err
	}
	discardIdx, ok := r.eligibleDiscardIndex(parsed.discard)
	if !ok {
		// card may exist but outside window
		found := false
		for _, c := range r.DiscardPile {
			if c.ID == parsed.discard.ID {
				found = true
				break
			}
		}
		if !found {
			return ErrCardNotInHand
		}
		return ErrInvalidDiscardWindow
	}
	// basis in hand
	if !handContainsAll(r.Hands[player], parsed.basis) {
		return ErrCardNotInHand
	}
	// duplicates
	all := []Card{parsed.discard}
	all = append(all, parsed.basis...)
	all = append(all, parsed.extensions...)
	if !uniqueCardIDs(all) {
		return ErrCombinationDuplicateCard
	}
	// transferred suffix
	suffix := r.DiscardPile[discardIdx:]
	transferredIDs := make(map[any]struct{})
	for _, c := range suffix {
		transferredIDs[c.ID] = struct{}{}
	}
	// extensions may be hand or transferred (excluding selected? selected already)
	for _, ext := range parsed.extensions {
		if cardInHand(r.Hands[player], ext) < 0 {
			if _, ok := transferredIDs[ext.ID]; !ok {
				return ErrCardNotInHand
			}
		}
	}
	// final combo
	finalCards := []Card{parsed.discard}
	finalCards = append(finalCards, parsed.basis...)
	finalCards = append(finalCards, parsed.extensions...)
	if !(IsValidSeri(finalCards, r.JokerRank) || IsValidParalel(finalCards, r.JokerRank)) {
		limit := r.HandLimit()
		if limit > 0 && len(r.Hands[player])+len(suffix) > limit {
			_ = r.EndRoundGameOver(player)
			return nil
		}
		return ErrInvalidJago
	}
	// Mutate
	next := r.cloneRound()
	// remove basis (+ hand ext) from hand
	removeList := append([]Card{}, parsed.basis...)
	for _, ext := range parsed.extensions {
		if cardInHand(r.Hands[player], ext) >= 0 {
			removeList = append(removeList, ext)
		}
	}
	hand, ok := removeCardsFromHand(r.Hands[player], removeList)
	if !ok {
		return ErrCardNotInHand
	}
	// transferred cards except combo cards go to hand
	for _, c := range suffix {
		// if c in final combo, skip (locked)
		if containsCardID(finalCards, c.ID) {
			continue
		}
		hand = append(hand, c)
	}
	next.Hands[player] = hand
	// remove suffix from discard pile
	next.DiscardPile = clonesCardsForRound(r.DiscardPile[:discardIdx])
	// lock combo
	kind := CombinationSeri
	if IsValidParalel(finalCards, r.JokerRank) && !IsValidSeri(finalCards, r.JokerRank) {
		kind = CombinationParalel
	}
	if IsValidSeri(finalCards, r.JokerRank) && IsValidParalel(finalCards, r.JokerRank) {
		kind = CombinationSeri
	}
	cardCopy := make([]Card, len(finalCards))
	copy(cardCopy, finalCards)
	combos := append([]LockedCombination{}, next.Table.Combinations...)
	combos = append(combos, LockedCombination{Kind: kind, Cards: cardCopy, Owner: player})
	next.Table = Table{Combinations: combos}
	next.HasDrawn = true
	*r = next
	return nil
}

func containsCardID(cards []Card, id any) bool {
	for _, c := range cards {
		if c.ID == id {
			return true
		}
	}
	return false
}

func (r *Round) TakeDiscard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) TakeDiscardCard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) TakeDiscardPile(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) TakeFromDiscard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) PickDiscard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) PickupDiscard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) Pickup(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) PickUp(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) DiscardPickup(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) PickDiscardCard(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) TakeDiscardSuffix(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) PickDiscardSuffix(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
