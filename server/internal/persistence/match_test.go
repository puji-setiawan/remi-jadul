package persistence

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCompletedMatchRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL completed match repository integration test")
	}
	dbrepo := mustOpen(t, dsn)

	matchID := fmt.Sprintf("match-test-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_ = dbrepo.Delete(matchID)
	})

	r1p1 := 45
	r1p2 := -15
	r2p1 := -50
	r3p1 := 20
	r3p2 := 30
	completedAt := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)

	match := CompletedMatch{
		MatchID:     matchID,
		RoundCount:  3,
		CompletedAt: completedAt,
		Players: []CompletedMatchPlayer{
			{PlayerID: "p1", DisplayName: "Raka", CumulativeScore: 15, Rank: 1},
			{PlayerID: "p2", DisplayName: "Sinta", CumulativeScore: 15, Rank: 1},
		},
		RoundScores: []CompletedRoundScore{
			{RoundNumber: 1, EndReason: "close", PlayerID: "p1", ScoreStatus: "calculated", RoundScore: &r1p1, CumulativeScore: 45},
			{RoundNumber: 1, EndReason: "close", PlayerID: "p2", ScoreStatus: "calculated", RoundScore: &r1p2, CumulativeScore: -15},
			{RoundNumber: 2, EndReason: "game_over", PlayerID: "p1", ScoreStatus: "game_over_penalty", RoundScore: &r2p1, CumulativeScore: -5},
			{RoundNumber: 2, EndReason: "game_over", PlayerID: "p2", ScoreStatus: "not_calculated", RoundScore: nil, CumulativeScore: -15},
			{RoundNumber: 3, EndReason: "deck_exhausted", PlayerID: "p1", ScoreStatus: "calculated", RoundScore: &r3p1, CumulativeScore: 15},
			{RoundNumber: 3, EndReason: "deck_exhausted", PlayerID: "p2", ScoreStatus: "calculated", RoundScore: &r3p2, CumulativeScore: 15},
		},
	}

	if err := dbrepo.Save(match); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := dbrepo.Load(match.MatchID)
	if err != nil {
		t.Fatalf("Load(%s): %v", match.MatchID, err)
	}
	if got.MatchID != match.MatchID || got.RoundCount != 3 {
		t.Fatalf("unexpected match header: %+v", got)
	}
	if !got.CompletedAt.Equal(completedAt) {
		t.Fatalf("CompletedAt = %v, want %v", got.CompletedAt, completedAt)
	}
	if len(got.Players) != 2 {
		t.Fatalf("len(Players) = %d, want 2", len(got.Players))
	}

	playersByID := make(map[string]CompletedMatchPlayer, len(got.Players))
	for _, p := range got.Players {
		if p.MatchID != match.MatchID {
			t.Fatalf("player MatchID = %s, want %s", p.MatchID, match.MatchID)
		}
		playersByID[p.PlayerID] = p
	}
	if p1, ok := playersByID["p1"]; !ok || p1.DisplayName != "Raka" || p1.CumulativeScore != 15 || p1.Rank != 1 {
		t.Fatalf("unexpected player p1: %+v", p1)
	}
	if p2, ok := playersByID["p2"]; !ok || p2.DisplayName != "Sinta" || p2.CumulativeScore != 15 || p2.Rank != 1 {
		t.Fatalf("unexpected player p2: %+v", p2)
	}

	if len(got.RoundScores) != 6 {
		t.Fatalf("len(RoundScores) = %d, want 6", len(got.RoundScores))
	}

	type roundKey struct {
		round    int
		playerID string
	}
	scoresByKey := make(map[roundKey]CompletedRoundScore, len(got.RoundScores))
	for _, rs := range got.RoundScores {
		if rs.MatchID != match.MatchID {
			t.Fatalf("round score MatchID = %s, want %s", rs.MatchID, match.MatchID)
		}
		scoresByKey[roundKey{round: rs.RoundNumber, playerID: rs.PlayerID}] = rs
	}

	r1p1Got, ok := scoresByKey[roundKey{round: 1, playerID: "p1"}]
	if !ok || r1p1Got.EndReason != "close" || r1p1Got.ScoreStatus != "calculated" || r1p1Got.RoundScore == nil || *r1p1Got.RoundScore != 45 || r1p1Got.CumulativeScore != 45 {
		t.Fatalf("unexpected round 1 p1 score: %+v", r1p1Got)
	}

	r2p1Got, ok := scoresByKey[roundKey{round: 2, playerID: "p1"}]
	if !ok || r2p1Got.EndReason != "game_over" || r2p1Got.ScoreStatus != "game_over_penalty" || r2p1Got.RoundScore == nil || *r2p1Got.RoundScore != -50 || r2p1Got.CumulativeScore != -5 {
		t.Fatalf("unexpected round 2 p1 game-over penalty score: %+v", r2p1Got)
	}

	r2p2Got, ok := scoresByKey[roundKey{round: 2, playerID: "p2"}]
	if !ok || r2p2Got.EndReason != "game_over" || r2p2Got.ScoreStatus != "not_calculated" || r2p2Got.RoundScore != nil || r2p2Got.CumulativeScore != -15 {
		t.Fatalf("unexpected round 2 p2 not-calculated score: %+v", r2p2Got)
	}

	if _, err := dbrepo.Load("match-does-not-exist"); !errors.Is(err, ErrCompletedMatchNotFound) {
		t.Fatalf("Load(unknown) err = %v, want %v", err, ErrCompletedMatchNotFound)
	}
	if err := dbrepo.Delete("match-does-not-exist"); !errors.Is(err, ErrCompletedMatchNotFound) {
		t.Fatalf("Delete(unknown) err = %v, want %v", err, ErrCompletedMatchNotFound)
	}

	// Atomicity check: round score referencing unknown player must roll back the entire match save.
	atomicMatchID := matchID + "-atomic-fail"
	t.Cleanup(func() {
		_ = dbrepo.Delete(atomicMatchID)
	})
	invalidMatch := CompletedMatch{
		MatchID:    atomicMatchID,
		RoundCount: 3,
		Players: []CompletedMatchPlayer{
			{PlayerID: "p1", DisplayName: "Raka", CumulativeScore: 10, Rank: 1},
		},
		RoundScores: []CompletedRoundScore{
			{RoundNumber: 1, EndReason: "close", PlayerID: "unknown-player", ScoreStatus: "calculated", RoundScore: &r1p1, CumulativeScore: 10},
		},
	}
	if err := dbrepo.Save(invalidMatch); err == nil {
		t.Fatal("expected Save with foreign-key violation to fail")
	}
	if _, err := dbrepo.Load(atomicMatchID); !errors.Is(err, ErrCompletedMatchNotFound) {
		t.Fatalf("Load after rolled-back Save err = %v, want %v", err, ErrCompletedMatchNotFound)
	}

	if err := dbrepo.Delete(match.MatchID); err != nil {
		t.Fatalf("Delete(%s): %v", match.MatchID, err)
	}
	if _, err := dbrepo.Load(match.MatchID); !errors.Is(err, ErrCompletedMatchNotFound) {
		t.Fatalf("Load after Delete err = %v, want %v", err, ErrCompletedMatchNotFound)
	}
}

func mustOpen(t *testing.T, dsn string) *MatchRepository {
	t.Helper()
	repo, err := NewPostgresGuestRepository(dsn)
	if err != nil {
		t.Fatalf("open match repository db: %v", err)
	}
	return NewMatchRepository(repo.DB())
}
