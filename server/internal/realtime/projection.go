package realtime

import (
	"fmt"
	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"
)

// Card projection.
type CardProjection struct {
	CardID        string `json:"card_id"`
	ID            string `json:"-"`
	Rank          string `json:"rank"`
	Suit          string `json:"suit"`
	IsActiveJoker bool   `json:"is_active_joker"`
}

// Table combination projection.
type TableCombinationProjection struct {
	CombinationID    string              `json:"combination_id"`
	CombinationId    string              `json:"-"`
	ID               string              `json:"-"`
	OwnerPlayerID    string              `json:"owner_player_id"`
	Type             string              `json:"type"`
	CardIDs          []string            `json:"card_ids"`
	CardIds          []string            `json:"-"`
	JokerAssignments []map[string]string `json:"joker_assignments"`
}

// Opponent projection.
type OpponentProjection struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	HandSize    int    `json:"hand_size"`
	Connection  string `json:"connection"`
}

type PlayerProjection = OpponentProjection

// Scoreboard entry.
type ScoreboardEntry struct {
	PlayerID            string `json:"player_id"`
	DisplayName         string `json:"display_name"`
	CompletedRoundScore *int   `json:"completed_round_score,omitempty"`
	CumulativeScore     int    `json:"cumulative_score"`
	RoundScore          *int   `json:"round_score,omitempty"`
	ScoreStatus         string `json:"score_status,omitempty"`
}

// Personalized game state projection.
type GameStateProjection struct {
	MatchID                 string                       `json:"match_id"`
	RoundNumber             int                          `json:"round_number"`
	RoundCount              int                          `json:"round_count"`
	RoundStatus             string                       `json:"round_status"`
	EndReason               *string                      `json:"end_reason"`
	CurrentPlayerID         string                       `json:"current_player_id"`
	Phase                   string                       `json:"phase"`
	ActiveJokerRank         string                       `json:"active_joker_rank"`
	OwnHand                 []CardProjection             `json:"own_hand"`
	Opponents               []OpponentProjection         `json:"opponents"`
	DiscardPile             []CardProjection             `json:"discard_pile"`
	DiscardPickupWindowSize int                          `json:"discard_pickup_window_size"`
	TableCombinations       []TableCombinationProjection `json:"table_combinations"`
	ClosedDeckCount         int                          `json:"closed_deck_count"`
	Scoreboard              []ScoreboardEntry            `json:"scoreboard"`

	// Convenience aliases.
	State      *GameStateProjection `json:"-"`
	Projection *GameStateProjection `json:"-"`
}

// PersonalizedState / aliases.
type PersonalizedState = GameStateProjection
type GameState = GameStateProjection
type StateProjection = GameStateProjection
type ProjectedState = GameStateProjection
type GameProjection = GameStateProjection

// Projector type.
type Projector struct{}

func NewProjector() *Projector { return &Projector{} }

func (p *Projector) Project(args ...interface{}) (GameStateProjection, error) {
	return Project(args...)
}

func (p *Projector) ProjectGameState(args ...interface{}) (GameStateProjection, error) {
	return Project(args...)
}

func (p *Projector) Build(args ...interface{}) (GameStateProjection, error) {
	return Project(args...)
}

func (p *Projector) ProjectState(args ...interface{}) (GameStateProjection, error) {
	return Project(args...)
}

func Project(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func ProjectGameState(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func NewGameStateProjection(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func BuildGameState(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func BuildGameStateProjection(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func BuildProjection(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func ProjectState(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func PersonalizeGameState(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func ProjectRoomState(args ...interface{}) (GameStateProjection, error) {
	return buildProjection(args...)
}

func buildProjection(args ...interface{}) (GameStateProjection, error) {
	var r *room.Room
	var g *game.Round
	var playerID string
	var displayNames map[string]string
	for _, arg := range args {
		switch v := arg.(type) {
		case *room.Room:
			r = v
		case room.Room:
			r2 := v
			r = &r2
		case *game.Round:
			g = v
		case game.Round:
			g2 := v
			g = &g2
		case string:
			if playerID == "" {
				playerID = v
			}
		case map[string]string:
			displayNames = v
		}
	}
	if r == nil {
		return GameStateProjection{}, fmt.Errorf("realtime: room is required")
	}
	if g == nil {
		g = r.Round
	}
	if g == nil {
		return GameStateProjection{}, fmt.Errorf("realtime: round is required")
	}
	if playerID == "" {
		if len(r.Members) > 0 {
			playerID = r.Members[0]
		} else {
			return GameStateProjection{}, fmt.Errorf("realtime: requesting player is required")
		}
	}

	// requesting player index
	reqIdx := -1
	for i, m := range r.Members {
		if m == playerID {
			reqIdx = i
			break
		}
	}
	if reqIdx < 0 {
		return GameStateProjection{}, fmt.Errorf("realtime: requesting player is not in room")
	}

	jokerRank := g.JokerRank
	window := g.HandLimit()
	if window == 0 {
		n := g.PlayerCount
		if n == 2 {
			window = 11
		} else if n == 3 {
			window = 9
		} else if n == 4 {
			window = 7
		}
	}

	// own hand
	ownHand := make([]CardProjection, 0)
	if reqIdx < len(g.Hands) {
		for _, c := range g.Hands[reqIdx] {
			ownHand = append(ownHand, cardProjection(c, jokerRank))
		}
	}

	// discard pile
	discard := make([]CardProjection, 0, len(g.DiscardPile))
	for _, c := range g.DiscardPile {
		discard = append(discard, cardProjection(c, jokerRank))
	}

	// opponents
	opponents := make([]OpponentProjection, 0)
	for i, member := range r.Members {
		if i == reqIdx {
			continue
		}
		handSize := 0
		if i < len(g.Hands) {
			handSize = len(g.Hands[i])
		}
		name := member
		if displayNames != nil {
			if n, ok := displayNames[member]; ok && n != "" {
				name = n
			}
		}
		opponents = append(opponents, OpponentProjection{
			PlayerID:    member,
			DisplayName: name,
			HandSize:    handSize,
			Connection:  "online",
		})
	}

	// table combinations
	tableCombos := make([]TableCombinationProjection, 0, len(g.Table.Combinations))
	for i, combo := range g.Table.Combinations {
		cardIDs := make([]string, 0, len(combo.Cards))
		for _, c := range combo.Cards {
			cardIDs = append(cardIDs, fmt.Sprint(c.ID))
		}
		ownerID := ""
		if combo.Owner >= 0 && combo.Owner < len(r.Members) {
			ownerID = r.Members[combo.Owner]
		} else if len(r.Members) > 0 {
			ownerID = r.Members[0]
		}
		tname := "seri"
		if combo.Kind == game.CombinationParalel {
			tname = "paralel"
		}
		tableCombos = append(tableCombos, TableCombinationProjection{
			CombinationID:    fmt.Sprintf("combo-%d", i+1),
			CombinationId:    fmt.Sprintf("combo-%d", i+1),
			ID:               fmt.Sprintf("combo-%d", i+1),
			OwnerPlayerID:    ownerID,
			Type:             tname,
			CardIDs:          cardIDs,
			CardIds:          cardIDs,
			JokerAssignments: []map[string]string{},
		})
	}

	// end reason
	var endReason *string
	status := g.Status
	if status == "" {
		if g.Ended {
			status = "ended"
		} else {
			status = "active"
		}
	}
	if g.Ended && (g.EndReason != "" || g.Reason != "") {
		reason := g.EndReason
		if reason == "" {
			reason = g.Reason
		}
		endReason = &reason
	}

	phase := "draw_required"
	if g.HasDrawn {
		phase = "discard_required"
	}

	activeJokerRank := ""
	if jokerRank != 0 {
		activeJokerRank = jokerRank.String()
	}

	// scoreboard
	scoreboard := make([]ScoreboardEntry, 0)
	if g.Ended {
		scores := g.ScoresSlice()
		for i, member := range r.Members {
			entry := ScoreboardEntry{
				PlayerID:        member,
				DisplayName:     member,
				CumulativeScore: 0,
			}
			if r.Match != nil && i < len(r.Match.CumulativeScores) {
				entry.CumulativeScore = r.Match.CumulativeScores[i]
			}
			if displayNames != nil {
				if n, ok := displayNames[member]; ok && n != "" {
					entry.DisplayName = n
				}
			}
			if g.GameOver {
				if i == g.ClosePlayer || i < len(scores) && scores[i] == -50 {
					// causing game over player only -50; others not calculated
				}
			}
			if i < len(scores) {
				score := scores[i]
				if g.GameOver {
					// only causing player calculated? We store scores already -50 for causing only.
					if score == 0 && !(i >= 0) {
						// not calculated
					}
				}
				// For game over, all players except causing have 0 scores in our model; mark status.
				if g.GameOver {
					if score == -50 {
						rs := score
						entry.RoundScore = &rs
						entry.CompletedRoundScore = &rs
						entry.ScoreStatus = "calculated"
					} else {
						entry.ScoreStatus = "not_calculated"
					}
				} else {
					rs := score
					entry.RoundScore = &rs
					entry.CompletedRoundScore = &rs
					entry.ScoreStatus = "calculated"
				}
			}
			scoreboard = append(scoreboard, entry)
		}
	}

	proj := GameStateProjection{
		MatchID:                 r.ID,
		RoundNumber:             roundNumber(r),
		RoundCount:              3,
		RoundStatus:             status,
		EndReason:               endReason,
		CurrentPlayerID:         currentPlayerID(r, g),
		Phase:                   phase,
		ActiveJokerRank:         activeJokerRank,
		OwnHand:                 ownHand,
		Opponents:               opponents,
		DiscardPile:             discard,
		DiscardPickupWindowSize: window,
		TableCombinations:       tableCombos,
		ClosedDeckCount:         len(g.ClosedDeck),
		Scoreboard:              scoreboard,
	}
	proj.State = &proj
	proj.Projection = &proj
	return proj, nil
}

func currentPlayerID(r *room.Room, g *game.Round) string {
	idx := g.CurrentPlayer
	if idx >= 0 && idx < len(r.Members) {
		return r.Members[idx]
	}
	if len(r.Members) > 0 {
		return r.Members[0]
	}
	return ""
}

func roundNumber(r *room.Room) int {
	if r.Match != nil && r.Match.RoundNumber > 0 {
		return r.Match.RoundNumber
	}
	return 1
}

func cardProjection(c game.Card, jokerRank game.Rank) CardProjection {
	cp := CardProjection{
		CardID:        fmt.Sprint(c.ID),
		ID:            fmt.Sprint(c.ID),
		Rank:          c.Rank.String(),
		Suit:          suitName(c.Suit),
		IsActiveJoker: c.Rank == jokerRank,
	}
	return cp
}

func suitName(s game.Suit) string {
	switch s {
	case game.SuitSpades:
		return "spades"
	case game.SuitHearts:
		return "hearts"
	case game.SuitDiamonds:
		return "diamonds"
	case game.SuitClubs:
		return "clubs"
	default:
		return "unknown"
	}
}
