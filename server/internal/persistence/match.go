package persistence

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// err helpers reused.
var ErrCompletedMatchNotFound = errors.New("persistence: completed match not found")

// Domain result objects.
type CompletedMatchPlayer struct {
	MatchID         string
	PlayerID        string
	DisplayName     string
	CumulativeScore int
	Rank            int
}

type CompletedRoundScore struct {
	MatchID         string
	RoundNumber     int
	EndReason       string
	PlayerID        string
	ScoreStatus     string
	RoundScore      *int
	CumulativeScore int
}

type CompletedMatch struct {
	MatchID     string
	RoundCount  int
	CompletedAt time.Time
	Players     []CompletedMatchPlayer
	RoundScores []CompletedRoundScore
}

// Aliases.
type CompletedMatchResult = CompletedMatch
type MatchResult = CompletedMatch
type RoundResult = CompletedRoundScore
type CompletedRoundResult = CompletedRoundScore
type MatchScore = CompletedMatchPlayer
type PlayerMatchScore = CompletedMatchPlayer

// GORM models.
type completedMatchModel struct {
	MatchID     string    `gorm:"column:match_id;type:text;primaryKey"`
	RoundCount  int       `gorm:"column:round_count;not null"`
	CompletedAt time.Time `gorm:"column:completed_at;not null;default:now()"`
}

func (completedMatchModel) TableName() string { return "completed_matches" }

type completedMatchPlayerModel struct {
	MatchID         string `gorm:"column:match_id;type:text;not null"`
	PlayerID        string `gorm:"column:player_id;type:text;not null"`
	DisplayName     string `gorm:"column:display_name;not null"`
	CumulativeScore int    `gorm:"column:cumulative_score;not null"`
	Rank            int    `gorm:"column:rank;not null"`
}

func (completedMatchPlayerModel) TableName() string { return "completed_match_players" }

type completedRoundScoreModel struct {
	MatchID         string `gorm:"column:match_id;type:text;not null"`
	RoundNumber     int    `gorm:"column:round_number;not null"`
	EndReason       string `gorm:"column:end_reason;not null"`
	PlayerID        string `gorm:"column:player_id;type:text;not null"`
	ScoreStatus     string `gorm:"column:score_status;not null"`
	RoundScore      *int   `gorm:"column:round_score"`
	CumulativeScore int    `gorm:"column:cumulative_score;not null"`
}

func (completedRoundScoreModel) TableName() string { return "completed_round_scores" }

type MatchRepository struct {
	db *gorm.DB
}

// Type aliases.
type MatchResultRepository = MatchRepository
type CompletedMatchRepository = MatchRepository
type CompletedMatchStore = MatchRepository
type ResultRepository = MatchRepository

func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func NewCompletedMatchRepository(db *gorm.DB) *MatchRepository { return NewMatchRepository(db) }
func NewMatchResultRepository(db *gorm.DB) *MatchRepository    { return NewMatchRepository(db) }
func NewResultRepository(db *gorm.DB) *MatchRepository         { return NewMatchRepository(db) }

func (r *MatchRepository) DB() *gorm.DB { return r.db }

func (r *MatchRepository) Save(match CompletedMatch) error {
	now := match.CompletedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		mm := completedMatchModel{
			MatchID:     match.MatchID,
			RoundCount:  match.RoundCount,
			CompletedAt: now,
		}
		if err := tx.Create(&mm).Error; err != nil {
			return fmtDBError(err)
		}
		for _, p := range match.Players {
			pm := completedMatchPlayerModel{
				MatchID:         match.MatchID,
				PlayerID:        p.PlayerID,
				DisplayName:     p.DisplayName,
				CumulativeScore: p.CumulativeScore,
				Rank:            p.Rank,
			}
			if err := tx.Create(&pm).Error; err != nil {
				return fmtDBError(err)
			}
		}
		for _, rs := range match.RoundScores {
			rr := completedRoundScoreModel{
				MatchID:         match.MatchID,
				RoundNumber:     rs.RoundNumber,
				EndReason:       rs.EndReason,
				PlayerID:        rs.PlayerID,
				ScoreStatus:     rs.ScoreStatus,
				RoundScore:      rs.RoundScore,
				CumulativeScore: rs.CumulativeScore,
			}
			if err := tx.Create(&rr).Error; err != nil {
				return fmtDBError(err)
			}
		}
		return nil
	})
}

// SaveCompletedMatch aliases.
func (r *MatchRepository) SaveCompletedMatch(match CompletedMatch) error   { return r.Save(match) }
func (r *MatchRepository) Create(match CompletedMatch) error               { return r.Save(match) }
func (r *MatchRepository) CreateCompletedMatch(match CompletedMatch) error { return r.Save(match) }
func (r *MatchRepository) Persist(match CompletedMatch) error              { return r.Save(match) }
func (r *MatchRepository) Store(match CompletedMatch) error                { return r.Save(match) }

func (r *MatchRepository) Load(matchID string) (CompletedMatch, error) {
	var mm completedMatchModel
	if err := r.db.Where("match_id = ?", matchID).First(&mm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CompletedMatch{}, ErrCompletedMatchNotFound
		}
		return CompletedMatch{}, fmtDBError(err)
	}
	var players []completedMatchPlayerModel
	if err := r.db.Where("match_id = ?", matchID).Find(&players).Error; err != nil {
		return CompletedMatch{}, fmtDBError(err)
	}
	var rounds []completedRoundScoreModel
	if err := r.db.Where("match_id = ?", matchID).Find(&rounds).Error; err != nil {
		return CompletedMatch{}, fmtDBError(err)
	}
	result := CompletedMatch{
		MatchID:     mm.MatchID,
		RoundCount:  mm.RoundCount,
		CompletedAt: mm.CompletedAt,
	}
	for _, p := range players {
		result.Players = append(result.Players, CompletedMatchPlayer{
			MatchID:         p.MatchID,
			PlayerID:        p.PlayerID,
			DisplayName:     p.DisplayName,
			CumulativeScore: p.CumulativeScore,
			Rank:            p.Rank,
		})
	}
	for _, rs := range rounds {
		score := rs.RoundScore
		result.RoundScores = append(result.RoundScores, CompletedRoundScore{
			MatchID:         rs.MatchID,
			RoundNumber:     rs.RoundNumber,
			EndReason:       rs.EndReason,
			PlayerID:        rs.PlayerID,
			ScoreStatus:     rs.ScoreStatus,
			RoundScore:      score,
			CumulativeScore: rs.CumulativeScore,
		})
	}
	return result, nil
}

// Load aliases.
func (r *MatchRepository) LoadCompletedMatch(matchID string) (CompletedMatch, error) {
	return r.Load(matchID)
}
func (r *MatchRepository) Get(matchID string) (CompletedMatch, error) { return r.Load(matchID) }
func (r *MatchRepository) GetCompletedMatch(matchID string) (CompletedMatch, error) {
	return r.Load(matchID)
}
func (r *MatchRepository) Find(matchID string) (CompletedMatch, error) { return r.Load(matchID) }
func (r *MatchRepository) FindCompletedMatch(matchID string) (CompletedMatch, error) {
	return r.Load(matchID)
}
func (r *MatchRepository) LoadMatch(matchID string) (CompletedMatch, error) {
	return r.Load(matchID)
}
func (r *MatchRepository) GetMatch(matchID string) (CompletedMatch, error) { return r.Load(matchID) }

func (r *MatchRepository) Delete(matchID string) error {
	res := r.db.Delete(&completedMatchModel{}, "match_id = ?", matchID)
	if res.Error != nil {
		return fmtDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrCompletedMatchNotFound
	}
	return nil
}

func fmtDBError(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrDatabase, err)
}
