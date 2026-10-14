package persistence

import (
	"os"
	"testing"
)

func TestCompletedMatchRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	dbrepo := mustOpen(t, dsn)
	score := 5
	match := CompletedMatch{
		MatchID:    "match-test-1",
		RoundCount: 3,
		Players: []CompletedMatchPlayer{
			{PlayerID: "p1", DisplayName: "Raka", CumulativeScore: 10, Rank: 1},
		},
		RoundScores: []CompletedRoundScore{
			{RoundNumber: 1, EndReason: "close", PlayerID: "p1", ScoreStatus: "calculated", RoundScore: &score, CumulativeScore: 10},
			{RoundNumber: 2, EndReason: "game_over", PlayerID: "p1", ScoreStatus: "not_calculated", RoundScore: nil, CumulativeScore: 10},
		},
	}
	if err := dbrepo.Save(match); err != nil {
		t.Fatal(err)
	}
	got, err := dbrepo.Load(match.MatchID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RoundCount != 3 || len(got.Players) != 1 || len(got.RoundScores) != 2 {
		t.Fatalf("got=%+v", got)
	}
	if got.RoundScores[1].RoundScore != nil {
		t.Fatal("game-over score should be nil")
	}
	// cleanup
	_ = dbrepo.Delete(match.MatchID)
}

func mustOpen(t *testing.T, dsn string) *MatchRepository {
	_ = t
	repo, err := NewPostgresGuestRepository(dsn)
	if err != nil {
		t.Fatal(err)
	}
	return NewMatchRepository(repo.DB())
}
