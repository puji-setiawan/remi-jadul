package game

import "errors"

var (
	ErrRoundAlreadyEnded = errors.New("game: round already ended")
	ErrRoundEnded        = errors.New("game: round already ended")
	ErrInvalidPlayer     = errors.New("game: invalid player")
	ErrRoundNotEnded     = errors.New("game: round not ended")
	ErrGameOver          = errors.New("game: game over")
)

func (r *Round) cloneRound() Round {
	next := *r
	next.Hands = cloneHandsForRound(r.Hands)
	next.ClosedDeck = clonesCardsForRound(r.ClosedDeck)
	next.DiscardPile = clonesCardsForRound(r.DiscardPile)
	next.Table = cloneTable(r.Table)
	next.Scores = cloneInts(r.Scores)
	next.RoundScores = cloneInts(r.RoundScores)
	next.PlayerScores = cloneInts(r.PlayerScores)
	return next
}

func cloneInts(values []int) []int {
	out := make([]int, len(values))
	copy(out, values)
	return out
}

func cloneTable(t Table) Table {
	out := Table{Combinations: make([]LockedCombination, len(t.Combinations))}
	copy(out.Combinations, t.Combinations)
	for i := range out.Combinations {
		cards := make([]Card, len(out.Combinations[i].Cards))
		copy(cards, out.Combinations[i].Cards)
		out.Combinations[i].Cards = cards
	}
	return out
}

func (r *Round) startingHandSizeOrZero() int {
	size, ok := startingHandSize(r.PlayerCount)
	if !ok || size <= 0 {
		return 0
	}
	return size
}

func (r *Round) HandLimit() int { return r.startingHandSizeOrZero() }

func (r *Round) PlayerHandSize() int { return r.startingHandSizeOrZero() }

func (r *Round) scorePlayers(closer int, closeCard *Card, gameover bool, gameoverPlayer int) []int {
	scores := make([]int, r.PlayerCount)
	if gameover {
		if gameoverPlayer >= 0 && gameoverPlayer < r.PlayerCount {
			scores[gameoverPlayer] = -50
		}
		return scores
	}
	for p := 0; p < r.PlayerCount; p++ {
		handScore := 0
		if p < len(r.Hands) {
			handScore = BestPartition(r.Hands[p], r.JokerRank).Score
		}
		lockedScore := 0
		for _, combo := range r.Table.Combinations {
			owner := combo.Owner
			if owner == p || (owner == 0 && p == 0) {
				lockedScore += ScorePartition([]Combination{combo}, nil, r.JokerRank)
			}
		}
		scores[p] = handScore + lockedScore
		if closer >= 0 && p == closer && closeCard != nil {
			scores[p] += CloseCardValue(*closeCard)
		}
	}
	return scores
}

func (r *Round) IsEnded() bool    { return r.Ended }
func (r *Round) IsGameOver() bool { return r.GameOver }
func (r *Round) IsClosed() bool   { return r.Closed }
func (r *Round) EndReasonValue() string {
	if r.EndReason != "" {
		return r.EndReason
	}
	return r.Reason
}
func (r *Round) ReasonValue() string { return r.EndReasonValue() }
func (r *Round) SetEndedReason(reason string) {
	r.Ended = true
	r.Status = "ended"
	r.EndReason = reason
	r.Reason = reason
}

func (r *Round) ScoresSlice() []int {
	if len(r.Scores) > 0 {
		return cloneInts(r.Scores)
	}
	if len(r.RoundScores) > 0 {
		return cloneInts(r.RoundScores)
	}
	return cloneInts(r.PlayerScores)
}
func (r *Round) GetScores() []int { return r.ScoresSlice() }
func (r *Round) PlayerScore(player int) int {
	scores := r.ScoresSlice()
	if player < 0 || player >= len(scores) {
		return 0
	}
	return scores[player]
}
func (r *Round) RoundScore(player int) int { return r.PlayerScore(player) }
func (r *Round) ScoreFor(player int) int   { return r.PlayerScore(player) }
func (r *Round) Score(player int) int      { return r.PlayerScore(player) }
func (r *Round) GetScore(player int) int   { return r.PlayerScore(player) }

func (r *Round) ScoreRound() []int {
	next := r.cloneRound()
	scores := next.scorePlayers(-1, nil, false, -1)
	next.Scores = scores
	next.RoundScores = cloneInts(scores)
	next.PlayerScores = cloneInts(scores)
	*r = next
	return cloneInts(scores)
}
func (r *Round) CalculateScores() []int   { return r.ScoreRound() }
func (r *Round) ComputeScores() []int     { return r.ScoreRound() }
func (r *Round) ScorePlayers() []int      { return r.ScoreRound() }
func (r *Round) ReScore() []int           { return r.ScoreRound() }
func (r *Round) RecalculateScores() []int { return r.ScoreRound() }

func (r *Round) End() error { return r.EndRound("completed") }
func (r *Round) EndRound(reason string) error {
	if r.Ended {
		return ErrRoundAlreadyEnded
	}
	next := r.cloneRound()
	next.Ended = true
	next.Status = "ended"
	next.EndReason = reason
	next.Reason = reason
	next.Scores = next.scorePlayers(-1, nil, false, -1)
	next.RoundScores = cloneInts(next.Scores)
	next.PlayerScores = cloneInts(next.Scores)
	*r = next
	return nil
}
func (r *Round) FinishRound(reason string) error   { return r.EndRound(reason) }
func (r *Round) CompleteRound(reason string) error { return r.EndRound(reason) }
func (r *Round) EndRoundNormal() error             { return r.EndRound("completed") }
func (r *Round) EndRoundDeckEmpty() error {
	next := r.cloneRound()
	next.Ended = true
	next.Status = "ended"
	next.EndReason = "deck_empty"
	next.Reason = next.EndReason
	next.Scores = next.scorePlayers(-1, nil, false, -1)
	next.RoundScores = cloneInts(next.Scores)
	next.PlayerScores = cloneInts(next.Scores)
	*r = next
	return nil
}
func (r *Round) EndDeckEmpty() error   { return r.EndRoundDeckEmpty() }
func (r *Round) EndByDeckEmpty() error { return r.EndRoundDeckEmpty() }
func (r *Round) EndEmptyDeck() error   { return r.EndRoundDeckEmpty() }
func (r *Round) EndRoundGameOver(player int) error {
	next := r.cloneRound()
	next.Ended = true
	next.Status = "ended"
	next.GameOver = true
	next.EndReason = "game_over"
	next.Reason = next.EndReason
	scores := make([]int, r.PlayerCount)
	if player >= 0 && player < r.PlayerCount {
		scores[player] = -50
	}
	next.Scores = scores
	next.RoundScores = cloneInts(scores)
	next.PlayerScores = cloneInts(scores)
	*r = next
	return nil
}
func (r *Round) EndGameOver(player int) error     { return r.EndRoundGameOver(player) }
func (r *Round) GameOverEnd(player int) error     { return r.EndRoundGameOver(player) }
func (r *Round) ApplyGameOver(player int) error   { return r.EndRoundGameOver(player) }
func (r *Round) TriggerGameOver(player int) error { return r.EndRoundGameOver(player) }
func (r *Round) GameOverPlayer(player int)        { _ = r.EndRoundGameOver(player) }
func (r *Round) SetGameOver(player int)           { _ = r.EndRoundGameOver(player) }
func (r *Round) MarkGameOver(player int)          { _ = r.EndRoundGameOver(player) }
