package game

import (
	"fmt"
	"sort"
)

// Scoring values from GAME_RULES.md.
func rankCombinationValue(rank Rank) int {
	switch rank {
	case RankAce:
		return 3
	case RankJack, RankQueen, RankKing:
		return 2
	default:
		return 1
	}
}

func rankCloseValue(rank Rank) int {
	switch rank {
	case RankAce:
		return 15
	case RankJack, RankQueen, RankKing:
		return 10
	default:
		return 5
	}
}

func rankUncombinedValue(rank Rank) int {
	return -rankCombinationValue(rank)
}

// Card score helpers.
func CombinationCardValue(card Card, jokerRank Rank) int {
	if IsJoker(card, jokerRank) {
		return rankCombinationValue(jokerRank)
	}
	return rankCombinationValue(card.Rank)
}

func UncombinedCardValue(card Card, jokerRank Rank) int {
	if IsJoker(card, jokerRank) {
		return rankUncombinedValue(jokerRank)
	}
	return rankUncombinedValue(card.Rank)
}

func CloseCardValue(card Card) int {
	return rankCloseValue(card.Rank)
}

// Aliases commonly expected by callers.
func CardCombinationValue(card Card, jokerRank Rank) int {
	return CombinationCardValue(card, jokerRank)
}
func CardUncombinedValue(card Card, jokerRank Rank) int { return UncombinedCardValue(card, jokerRank) }
func CardScore(card Card, jokerRank Rank) int           { return CombinationCardValue(card, jokerRank) }

// Partition is one deterministic partition used for scoring.
type Partition struct {
	Combinations []Combination
	Uncombined   []Card
	Score        int
}

// HandScore and aliases.
type HandScore = Partition
type ScoredPartition = Partition
type ScoringResult = Partition

func bestIntervalScore(indices []int, cards []Card, jokerRank Rank, kind CombinationKind) (int, bool) {
	n := len(indices)
	if n < 3 {
		return 0, false
	}
	cardList := make([]Card, n)
	for i, idx := range indices {
		cardList[i] = cards[idx]
	}
	if kind == CombinationSeri {
		natural := make([]Card, 0, n)
		for _, c := range cardList {
			if !IsJoker(c, jokerRank) {
				natural = append(natural, c)
			}
		}
		if len(natural) == 0 {
			// All jokers: choose the highest-scoring consecutive run.
			best := 0
			found := false
			for start := int(RankAce); start+(n-1) <= int(RankKing); start++ {
				score := 0
				for i := 0; i < n; i++ {
					score += rankCombinationValue(Rank(start + i))
				}
				if !found || score > best {
					best = score
					found = true
				}
			}
			if !found {
				return 0, false
			}
			return best, true
		}
		// Natural cards must share suit.
		for i := 1; i < len(natural); i++ {
			if natural[i].Suit != natural[0].Suit {
				return 0, false
			}
		}
		seen := make(map[Rank]struct{}, len(natural))
		for _, c := range natural {
			if _, ok := seen[c.Rank]; ok {
				return 0, false
			}
			seen[c.Rank] = struct{}{}
		}
		minRank, maxRank := natural[0].Rank, natural[0].Rank
		for _, c := range natural[1:] {
			if c.Rank < minRank {
				minRank = c.Rank
			}
			if c.Rank > maxRank {
				maxRank = c.Rank
			}
		}
		if int(maxRank-minRank) > n-1 {
			return 0, false
		}
		lowestStart := maxRank - Rank(n-1)
		if lowestStart < RankAce {
			lowestStart = RankAce
		}
		highestStart := minRank
		if highestStart > RankKing-Rank(n-1) {
			highestStart = RankKing - Rank(n-1)
		}
		best := 0
		found := false
		for start := lowestStart; start <= highestStart; start++ {
			naturalSet := make(map[Rank]struct{}, len(natural))
			for _, c := range natural {
				naturalSet[c.Rank] = struct{}{}
			}
			score := 0
			for i := range n {
				rank := Rank(int(start) + i)
				if _, ok := naturalSet[rank]; ok {
					// find corresponding card value (rank value, same)
					score += rankCombinationValue(rank)
				} else {
					score += rankCombinationValue(rank)
				}
			}
			if !found || score > best {
				best = score
				found = true
			}
		}
		if !found {
			return 0, false
		}
		return best, true
	}
	// Paralel
	natural := make([]Card, 0, n)
	for _, c := range cardList {
		if !IsJoker(c, jokerRank) {
			natural = append(natural, c)
		}
	}
	if len(natural) == 0 {
		// All jokers can represent at most four distinct suits at one rank.
		if n > 4 {
			return 0, false
		}
		best := rankCombinationValue(RankAce) * n
		for r := RankAce; r <= RankKing; r++ {
			value := rankCombinationValue(r) * n
			if value > best {
				best = value
			}
		}
		return best, true
	}
	commonRank := natural[0].Rank
	suits := make(map[Suit]struct{})
	for _, c := range natural {
		if c.Rank != commonRank {
			return 0, false
		}
		if _, ok := suits[c.Suit]; ok {
			return 0, false
		}
		suits[c.Suit] = struct{}{}
	}
	if len(natural) > 4 || n > 4 {
		return 0, false
	}
	return rankCombinationValue(commonRank) * n, true
}

type candidate struct {
	mask  uint
	score int
	combo Combination
}

func candidatesForPartition(cards []Card, jokerRank Rank) []candidate {
	n := len(cards)
	var result []candidate
	for mask := 1; mask < (1 << n); mask++ {
		if bitsSet(mask) < 3 {
			continue
		}
		indices := make([]int, 0, bitsSet(mask))
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				indices = append(indices, i)
			}
		}
		subsetCards := make([]Card, len(indices))
		for i, idx := range indices {
			subsetCards[i] = cards[idx]
		}
		var bestKind CombinationKind
		bestScore := 0
		found := false
		for _, kind := range []CombinationKind{CombinationSeri, CombinationParalel} {
			score, ok := bestIntervalScore(indices, cards, jokerRank, kind)
			if ok && (!found || score > bestScore) {
				bestScore = score
				bestKind = kind
				found = true
			}
		}
		// ensure validation rules agree (especially duplicate ids)
		if !uniqueCardIDs(subsetCards) {
			continue
		}
		if found {
			combo := Combination{Kind: bestKind, Cards: subsetCards}
			result = append(result, candidate{mask: uint(mask), score: bestScore, combo: combo})
		}
	}
	return result
}

func bitsSet(v int) int {
	count := 0
	for v != 0 {
		count += v & 1
		v >>= 1
	}
	return count
}

func uncombinedValue(cards []Card, used uint, jokerRank Rank) int {
	total := 0
	for i, card := range cards {
		if used&(1<<i) != 0 {
			continue
		}
		total += UncombinedCardValue(card, jokerRank)
	}
	return total
}

func partitionKey(combinations []Combination, uncombined []Card) string {
	key := fmt.Sprintf("c=%d", len(combinations))
	type sig struct {
		kind CombinationKind
		ids  []int
	}
	sigs := make([]sig, len(combinations))
	for i, combo := range combinations {
		ids := make([]int, len(combo.Cards))
		for j, c := range combo.Cards {
			ids[j] = cardIDToIntForSort(c.ID)
		}
		sort.Ints(ids)
		sigs[i] = sig{kind: combo.Kind, ids: ids}
	}
	sort.Slice(sigs, func(a, b int) bool {
		if sigs[a].kind != sigs[b].kind {
			return sigs[a].kind < sigs[b].kind
		}
		for j := 0; j < len(sigs[a].ids) && j < len(sigs[b].ids); j++ {
			if sigs[a].ids[j] != sigs[b].ids[j] {
				return sigs[a].ids[j] < sigs[b].ids[j]
			}
		}
		return len(sigs[a].ids) < len(sigs[b].ids)
	})
	for _, s := range sigs {
		key += fmt.Sprintf("|k=%d", s.kind)
		for _, id := range s.ids {
			key += fmt.Sprintf(",%d", id)
		}
	}
	uncombinedIDs := make([]int, len(uncombined))
	for i, c := range uncombined {
		uncombinedIDs[i] = cardIDToIntForSort(c.ID)
	}
	sort.Ints(uncombinedIDs)
	key += "|u"
	for _, id := range uncombinedIDs {
		key += fmt.Sprintf(",%d", id)
	}
	return key
}

func cardIDToIntForSort(id any) int {
	switch v := id.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// BestPartition returns the highest-scoring partition for a hand.
func BestPartition(cards []Card, jokerRank Rank) Partition {
	bestScore := -1 << 60
	var bestCombos []Combination
	bestKey := ""
	var bestUncombined []Card
	cands := candidatesForPartition(cards, jokerRank)

	var search func(start int, used uint, score int, combos []Combination)
	search = func(start int, used uint, score int, combos []Combination) {
		finalScore := score + uncombinedValue(cards, used, jokerRank)
		finalCombos := make([]Combination, len(combos))
		copy(finalCombos, combos)
		uncombined := make([]Card, 0)
		for i, card := range cards {
			if used&(1<<i) == 0 {
				uncombined = append(uncombined, card)
			}
		}
		choiceKey := partitionKey(finalCombos, uncombined)
		if finalScore > bestScore || (finalScore == bestScore && lessKey(choiceKey, bestKey)) {
			bestScore = finalScore
			bestCombos = finalCombos
			bestKey = choiceKey
			bestUncombined = uncombined
		}
		for i := start; i < len(cands); i++ {
			if cands[i].mask&used != 0 {
				continue
			}
			newCombos := append([]Combination(nil), combos...)
			newCombos = append(newCombos, cands[i].combo)
			search(i+1, used|cands[i].mask, score+cands[i].score, newCombos)
		}
	}

	search(0, 0, 0, nil)
	uncombined := bestUncombined
	if uncombined == nil {
		uncombined = []Card{}
	}
	if bestCombos == nil {
		bestCombos = []Combination{}
	}
	return Partition{
		Combinations: bestCombos,
		Uncombined:   uncombined,
		Score:        bestScore,
	}
}

func lessKey(a, b string) bool {
	if b == "" {
		return true
	}
	return a < b
}

// ScoreHand returns the highest possible score for a hand.
func ScoreHand(cards []Card, jokerRank Rank) int {
	return BestPartition(cards, jokerRank).Score
}

// BestHandPartition is an alias.
func BestHandPartition(cards []Card, jokerRank Rank) Partition {
	return BestPartition(cards, jokerRank)
}

// ScoreHandResult aliases.
func ScoreHandResult(cards []Card, jokerRank Rank) Partition { return BestPartition(cards, jokerRank) }
func ScoreHandPartition(cards []Card, jokerRank Rank) Partition {
	return BestPartition(cards, jokerRank)
}
func BestScoringPartition(cards []Card, jokerRank Rank) Partition {
	return BestPartition(cards, jokerRank)
}

// ScoreHandWithClose scores a hand whose closing card is removed and gets the
// closing-card value.
func ScoreHandWithClose(cards []Card, closeCard Card, jokerRank Rank) int {
	return BestPartition(cards, jokerRank).Score + CloseCardValue(closeCard)
}

// Aliases.
func ScoreCloseHand(cards []Card, closeCard Card, jokerRank Rank) int {
	return ScoreHandWithClose(cards, closeCard, jokerRank)
}
func ScoreHandClosing(cards []Card, closeCard Card, jokerRank Rank) int {
	return ScoreHandWithClose(cards, closeCard, jokerRank)
}

// Internal score of a listed partition (validation not required).
func ScorePartition(combinations []Combination, uncombined []Card, jokerRank Rank) int {
	total := 0
	for _, combo := range combinations {
		indices := make([]int, len(combo.Cards))
		for i := range combo.Cards {
			indices[i] = i
		}
		score, ok := bestIntervalScore(indices, combo.Cards, jokerRank, combo.Kind)
		if !ok {
			for _, c := range combo.Cards {
				total += CombinationCardValue(c, jokerRank)
			}
			continue
		}
		total += score
	}
	for _, c := range uncombined {
		total += UncombinedCardValue(c, jokerRank)
	}
	return total
}

func indicesOfCards(cards []Card, all []Card) []int {
	// Helper not used for score fallback; returns empty.
	return nil
}
