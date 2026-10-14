package game

import (
	"reflect"
	"testing"
)

func TestMatchThreeRoundsCumulative(t *testing.T) {
	m := NewMatch(2)
	if m.TotalRounds != 3 || m.RoundNumber != 1 || m.Ended {
		t.Fatalf("initial match wrong: %+v", m)
	}
	if len(m.CumulativeScores) != 2 || m.CumulativeScores[0] != 0 || m.CumulativeScores[1] != 0 {
		t.Fatalf("cumulative wrong: %v", m.CumulativeScores)
	}
	if err := m.CompleteRound([]int{4, -1}); err != nil {
		t.Fatal(err)
	}
	if m.RoundNumber != 2 || m.Ended {
		t.Fatalf("after round1: round=%d ended=%v", m.RoundNumber, m.Ended)
	}
	if m.CumulativeScores[0] != 4 || m.CumulativeScores[1] != -1 {
		t.Fatalf("cumulative after round1: %v", m.CumulativeScores)
	}
	if err := m.CompleteRound([]int{1, 1}); err != nil {
		t.Fatal(err)
	}
	if m.RoundNumber != 3 || m.Ended {
		t.Fatalf("after round2: round=%d ended=%v", m.RoundNumber, m.Ended)
	}
	if m.CumulativeScores[0] != 5 || m.CumulativeScores[1] != 0 {
		t.Fatalf("cumulative after round2: %v", m.CumulativeScores)
	}
	if err := m.CompleteRound([]int{10, 3}); err != nil {
		t.Fatal(err)
	}
	if !m.Ended || m.RoundNumber != 3 {
		t.Fatalf("after round3: ended=%v round=%d", m.Ended, m.RoundNumber)
	}
	if m.CumulativeScores[0] != 15 || m.CumulativeScores[1] != 3 {
		t.Fatalf("final cumulative: %v", m.CumulativeScores)
	}
}

func TestMatchFourthRoundRejectedNoMutation(t *testing.T) {
	m := NewMatch(2)
	for i := 0; i < 3; i++ {
		if err := m.CompleteRound([]int{1, 2}); err != nil {
			t.Fatal(err)
		}
	}
	before := *m
	before.CumulativeScores = cloneInts(before.CumulativeScores)
	if err := m.StartNextRound(nil); err == nil {
		t.Fatal("fourth round accepted")
	}
	if !reflect.DeepEqual(*m, before) {
		t.Fatal("fourth-round rejection mutated match")
	}
	if err := m.CompleteRound([]int{9, 9}); err == nil {
		t.Fatal("fourth complete accepted")
	}
}

func TestMatchGameOverRoundScore(t *testing.T) {
	m := NewMatch(3)
	// two normal-ish rounds (score only)
	if err := m.CompleteRound([]int{5, 4, 6}); err != nil {
		t.Fatal(err)
	}
	roundScores := []int{-50, 0, 0}
	if err := m.CompleteRound(roundScores); err != nil {
		t.Fatal(err)
	}
	if m.CumulativeScores[0] != 5-50 || m.CumulativeScores[1] != 4 || m.CumulativeScores[2] != 6 {
		t.Fatalf("game-over cumulative: %v", m.CumulativeScores)
	}
	if err := m.CompleteRound([]int{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	rankings := m.FinalRankings()
	if len(rankings) != 3 {
		t.Fatalf("rankings len=%d", len(rankings))
	}
	// player1 or player2 highest
	if rankings[0].Rank != 1 {
		t.Fatalf("top rank=%d", rankings[0].Rank)
	}
}

func TestMatchEqualScoreRanking(t *testing.T) {
	m := NewMatch(2)
	m.CompleteRound([]int{10, 10})
	m.CompleteRound([]int{5, 5})
	m.CompleteRound([]int{0, 0})
	r := m.FinalRankings()
	if r[0].Score != r[1].Score || r[0].Rank != 1 || r[1].Rank != 1 {
		t.Fatalf("tie ranks wrong: %+v", r)
	}
}

func TestMatchRoundConstructorTotal(t *testing.T) {
	hands := make([][]Card, 3)
	for i := range hands {
		hands[i] = []Card{{ID: i + 1}}
	}
	round, err := NewRound(3, hands, []Card{{ID: 9}}, Seven)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatch(round)
	if m.PlayerCount != 3 || m.RoundNumber != 1 {
		t.Fatalf("round constructor match wrong: %+v", m)
	}
}
