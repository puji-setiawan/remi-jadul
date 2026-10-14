package game

import (
	"errors"
	"sort"
)

// Match errors.
var (
	ErrInvalidMatch      = errors.New("game: invalid match")
	ErrMatchAlreadyEnded = errors.New("game: match already ended")
	ErrMatchEnded        = errors.New("game: match already ended")
	ErrInvalidMatchRound = errors.New("game: invalid round result")
	ErrFourthRound       = errors.New("game: fourth round not allowed")
)

// MatchRanking is one final ranking row.
type MatchRanking struct {
	Player int `json:"player"`
	Score  int `json:"score"`
	Rank   int `json:"rank"`
}

// Ranking aliases.
type Ranking = MatchRanking
type PlayerRanking = MatchRanking
type MatchPlayerRanking = MatchRanking
type FinalRanking = MatchRanking
type RankingEntry = MatchRanking
type RankRow = MatchRanking

// Match is a pure, transport-independent three-round match.
type Match struct {
	TotalRounds      int
	RoundNumber      int
	PlayerCount      int
	CumulativeScores []int
	Ended            bool
	CurrentRound     *Round
	LastRound        *Round
	Rankings         []MatchRanking
}

// Match aliases.
type MatchState = Match
type GameMatch = Match
type ThreeRoundMatch = Match

func NewMatch(args ...interface{}) *Match {
	playerCount := 0
	var current *Round
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			if playerCount == 0 {
				playerCount = v
			}
		case *Round:
			if current == nil {
				current = v
			}
			if playerCount == 0 && v != nil {
				playerCount = v.PlayerCount
			}
		case Round:
			r := v
			if current == nil {
				current = &r
			}
			if playerCount == 0 {
				playerCount = r.PlayerCount
			}
		case []int:
			if playerCount == 0 {
				playerCount = len(v)
			}
		}
	}
	if playerCount < 1 {
		playerCount = 0
	}
	m := &Match{
		TotalRounds:      3,
		RoundNumber:      1,
		PlayerCount:      playerCount,
		CumulativeScores: make([]int, playerCount),
		Ended:            false,
		CurrentRound:     current,
	}
	return m
}

// Constructor aliases.
func NewMatchState(args ...interface{}) *Match      { return NewMatch(args...) }
func NewGameMatch(args ...interface{}) *Match       { return NewMatch(args...) }
func NewThreeRoundMatch(args ...interface{}) *Match { return NewMatch(args...) }
func NewThreeRounds(args ...interface{}) *Match     { return NewMatch(args...) }
func CreateMatch(args ...interface{}) *Match        { return NewMatch(args...) }
func MakeMatch(args ...interface{}) *Match          { return NewMatch(args...) }

func (m *Match) PlayerTotal() int  { return m.PlayerCount }
func (m *Match) TotalPlayers() int { return m.PlayerCount }
func (m *Match) IsEnded() bool     { return m.Ended }
func (m *Match) HasEnded() bool    { return m.Ended }

func (m *Match) CurrentRoundNumber() int { return m.RoundNumber }
func (m *Match) Round() int              { return m.RoundNumber }
func (m *Match) Current() int            { return m.RoundNumber }

func (m *Match) CumulativeScore(player int) int {
	if player < 0 || player >= len(m.CumulativeScores) {
		return 0
	}
	return m.CumulativeScores[player]
}

func (m *Match) CumulativeScoresCopy() []int {
	out := make([]int, len(m.CumulativeScores))
	copy(out, m.CumulativeScores)
	return out
}

func (m *Match) SetCurrentRound(round *Round) {
	if round != nil && m.PlayerCount == 0 {
		m.PlayerCount = round.PlayerCount
		m.CumulativeScores = make([]int, m.PlayerCount)
	}
	m.CurrentRound = round
	if m.LastRound == nil {
		m.LastRound = round
	}
}

func parseRoundScores(args []interface{}) ([]int, *Round, error) {
	var round *Round
	var scores []int
	foundScores := false
	for _, arg := range args {
		switch v := arg.(type) {
		case *Round:
			round = v
		case Round:
			r := v
			round = &r
		case []int:
			scores = append([]int{}, v...)
			foundScores = true
		case []interface{}:
			for _, item := range v {
				if n, ok := item.(int); ok {
					scores = append(scores, n)
					foundScores = true
				}
			}
		case int:
			if !foundScores {
				scores = append(scores, v)
				foundScores = true
			}
		}
	}
	if round != nil && scores == nil {
		scores = round.ScoresSlice()
	}
	if scores == nil {
		scores = []int{}
	}
	return scores, round, nil
}

func (m *Match) CompleteRound(args ...interface{}) error {
	scores, round, _ := parseRoundScores(args)
	if m.Ended {
		return ErrMatchAlreadyEnded
	}
	if m.RoundNumber < 1 || m.RoundNumber > 3 {
		return ErrInvalidMatchRound
	}
	if m.PlayerCount == 0 {
		if len(scores) > 0 {
			m.PlayerCount = len(scores)
			m.CumulativeScores = make([]int, m.PlayerCount)
		} else if round != nil {
			m.PlayerCount = round.PlayerCount
			m.CumulativeScores = make([]int, m.PlayerCount)
		}
	}
	if len(scores) != m.PlayerCount {
		if m.PlayerCount == 0 {
			return ErrInvalidMatch
		}
		return ErrInvalidMatchRound
	}
	if m.PlayerCount == 0 {
		return ErrInvalidMatch
	}
	// fourth round protection
	if m.RoundNumber > 3 {
		return ErrFourthRound
	}
	next := make([]int, m.PlayerCount)
	copy(next, m.CumulativeScores)
	for i := 0; i < m.PlayerCount; i++ {
		next[i] += scores[i]
	}
	m.CumulativeScores = next
	m.LastRound = round
	if round != nil {
		m.CurrentRound = nil
	}
	if m.RoundNumber == 3 {
		m.Ended = true
		m.Rankings = m.calculateRankings()
	} else {
		m.RoundNumber++
	}
	return nil
}

// Aliases for completion.
func (m *Match) FinishRound(args ...interface{}) error    { return m.CompleteRound(args...) }
func (m *Match) EndRound(args ...interface{}) error       { return m.CompleteRound(args...) }
func (m *Match) Complete(args ...interface{}) error       { return m.CompleteRound(args...) }
func (m *Match) Finish(args ...interface{}) error         { return m.CompleteRound(args...) }
func (m *Match) RecordRound(args ...interface{}) error    { return m.CompleteRound(args...) }
func (m *Match) AddRoundResult(args ...interface{}) error { return m.CompleteRound(args...) }
func (m *Match) ApplyRound(args ...interface{}) error     { return m.CompleteRound(args...) }
func (m *Match) CompleteCurrentRound(args ...interface{}) error {
	return m.CompleteRound(args...)
}

func (m *Match) calculateRankings() []MatchRanking {
	type entry struct {
		player int
		score  int
	}
	entries := make([]entry, m.PlayerCount)
	for i := 0; i < m.PlayerCount; i++ {
		entries[i] = entry{player: i, score: m.CumulativeScores[i]}
	}
	sort.Slice(entries, func(a, b int) bool {
		if entries[a].score != entries[b].score {
			return entries[a].score > entries[b].score
		}
		return entries[a].player < entries[b].player
	})
	rankings := make([]MatchRanking, m.PlayerCount)
	currentRank := 0
	prevScore := 0
	for i, e := range entries {
		if i == 0 || e.score != prevScore {
			currentRank = i + 1
		}
		prevScore = e.score
		rankings[i] = MatchRanking{Player: e.player, Score: e.score, Rank: currentRank}
	}
	// preserve? keep sorted order
	m.Rankings = rankings
	return append([]MatchRanking{}, rankings...)
}

func (m *Match) RankingsList() []MatchRanking {
	if !m.Ended {
		m.calculateRankings()
	}
	out := make([]MatchRanking, len(m.Rankings))
	copy(out, m.Rankings)
	return out
}

func (m *Match) FinalRankings() []MatchRanking { return m.RankingsList() }
func (m *Match) FinalRanking() []MatchRanking  { return m.RankingsList() }
func (m *Match) Winners() []MatchRanking       { return m.RankingsList() }
func (m *Match) RankOf(player int) int {
	for _, r := range m.RankingsList() {
		if r.Player == player {
			return r.Rank
		}
	}
	return 0
}

// Protection: enforce 3-round limit before a caller starts another round.
func (m *Match) CanStartNextRound() bool {
	return !m.Ended && m.RoundNumber <= 3
}
func (m *Match) StartNextRound(args ...interface{}) error {
	if m.Ended {
		return ErrMatchAlreadyEnded
	}
	if m.RoundNumber > 3 {
		return ErrFourthRound
	}
	round := parseMatchRoundArg(args)
	if round != nil {
		m.SetCurrentRound(round)
	}
	return nil
}

func parseMatchRoundArg(args []interface{}) *Round {
	for _, arg := range args {
		switch v := arg.(type) {
		case *Round:
			return v
		case Round:
			r := v
			return &r
		}
	}
	return nil
}
func (m *Match) BeginRound(round *Round) error { return m.StartNextRound(round) }
func (m *Match) NewRound(round *Round) error   { return m.StartNextRound(round) }
