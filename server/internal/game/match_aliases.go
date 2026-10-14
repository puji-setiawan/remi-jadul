package game

// Additional match errors.
var (
	ErrMatchFinished         = ErrMatchAlreadyEnded
	ErrMatchComplete         = ErrMatchAlreadyEnded
	ErrMatchFullyPlayed      = ErrMatchAlreadyEnded
	ErrTooManyRounds         = ErrFourthRound
	ErrFourthRoundNotAllowed = ErrFourthRound
	ErrMaxRoundsReached      = ErrFourthRound
	ErrRoundLimitReached     = ErrFourthRound
)

// Capability checks and lifecycle aliases.
func (m *Match) CanStartRound() bool    { return m.CanStartNextRound() }
func (m *Match) CanStartNext() bool     { return m.CanStartNextRound() }
func (m *Match) CanBeginRound() bool    { return m.CanStartNextRound() }
func (m *Match) CanPlayNextRound() bool { return m.CanStartNextRound() }

func (m *Match) IsFinished() bool { return m.Ended }
func (m *Match) IsComplete() bool { return m.Ended }
func (m *Match) Completed() bool  { return m.Ended }
func (m *Match) Done() bool       { return m.Ended }

func (m *Match) CurrentRoundNumberValue() int { return m.RoundNumber }
func (m *Match) CurrentRoundNo() int          { return m.RoundNumber }
func (m *Match) RoundNo() int                 { return m.RoundNumber }
func (m *Match) RoundCount() int              { return m.TotalRounds }
func (m *Match) TotalRoundCount() int         { return m.TotalRounds }

func (m *Match) CompleteMatch(args ...interface{}) error { return m.CompleteRound(args...) }
func (m *Match) FinishMatch(args ...interface{}) error   { return m.CompleteRound(args...) }
func (m *Match) EndMatch(args ...interface{}) error      { return m.CompleteRound(args...) }
func (m *Match) CloseMatch(args ...interface{}) error    { return m.CompleteRound(args...) }

func (m *Match) StartRound(args ...interface{}) error       { return m.StartNextRound(args...) }
func (m *Match) BeginNextRound(args ...interface{}) error   { return m.StartNextRound(args...) }
func (m *Match) BeginNext(args ...interface{}) error        { return m.StartNextRound(args...) }
func (m *Match) NextRound(args ...interface{}) error        { return m.StartNextRound(args...) }
func (m *Match) OpenRound(args ...interface{}) error        { return m.StartNextRound(args...) }
func (m *Match) OpenNextRound(args ...interface{}) error    { return m.StartNextRound(args...) }
func (m *Match) StartFourthRound(args ...interface{}) error { return m.StartNextRound(args...) }
func (m *Match) StartRound4(args ...interface{}) error      { return m.StartNextRound(args...) }

func (m *Match) CompleteCurrentMatch(args ...interface{}) error { return m.CompleteRound(args...) }
func (m *Match) RecordScore(args ...interface{}) error          { return m.CompleteRound(args...) }
func (m *Match) ApplyRoundResult(args ...interface{}) error     { return m.CompleteRound(args...) }
func (m *Match) AddRound(args ...interface{}) error             { return m.CompleteRound(args...) }
func (m *Match) AddScores(args ...interface{}) error            { return m.CompleteRound(args...) }
func (m *Match) AddRoundScores(args ...interface{}) error       { return m.CompleteRound(args...) }

func (m *Match) ScoreOf(player int) int        { return m.CumulativeScore(player) }
func (m *Match) MatchScore(player int) int     { return m.CumulativeScore(player) }
func (m *Match) PlayerScore(player int) int    { return m.CumulativeScore(player) }
func (m *Match) GetScore(player int) int       { return m.CumulativeScore(player) }
func (m *Match) GetCumulative(player int) int  { return m.CumulativeScore(player) }
func (m *Match) CumulativeScoresMethod() []int { return m.CumulativeScoresCopy() }

func (m *Match) Cumulative() []int  { return m.CumulativeScoresCopy() }
func (m *Match) Cumulatives() []int { return m.CumulativeScoresCopy() }
func (m *Match) Scores() []int      { return m.CumulativeScoresCopy() }
func (m *Match) MatchScores() []int { return m.CumulativeScoresCopy() }
func (m *Match) TotalScores() []int { return m.CumulativeScoresCopy() }
func (m *Match) Scoreboard() []int  { return m.CumulativeScoresCopy() }

func (m *Match) Ranking() []MatchRanking            { return m.RankingsList() }
func (m *Match) RankPlayers() []MatchRanking        { return m.RankingsList() }
func (m *Match) PlayersRanking() []MatchRanking     { return m.RankingsList() }
func (m *Match) FinalScoreRanking() []MatchRanking  { return m.RankingsList() }
func (m *Match) FinalScoresRanking() []MatchRanking { return m.RankingsList() }

func (m *Match) Winner() MatchRanking {
	r := m.RankingsList()
	if len(r) == 0 {
		return MatchRanking{}
	}
	return r[0]
}
