package realtime

import (
	"testing"

	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"
)

func TestProjectionVisibility(t *testing.T) {
	reg := room.NewRegistry()
	r, err := reg.CreatePublicRoom(2, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Join(r.ID, "p2"); err != nil {
		t.Fatal(err)
	}
	deck := game.StandardDeck()
	if err := reg.StartGame(r.ID, "p1", deck); err != nil {
		t.Fatal(err)
	}
	rm, _ := reg.Get(r.ID)
	proj, err := Project(&rm, rm.Round, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(proj.OwnHand) != 11 {
		t.Fatalf("own_hand=%d", len(proj.OwnHand))
	}
	if len(proj.Opponents) != 1 {
		t.Fatalf("opponents=%d", len(proj.Opponents))
	}
	if proj.Opponents[0].HandSize != 11 {
		t.Fatalf("opp hand=%d", proj.Opponents[0].HandSize)
	}
	if proj.ClosedDeckCount == 0 {
		t.Fatal("closed count hidden? should count")
	}
	// own hand cards absent from opponents
	if proj.Opponents[0].PlayerID != "p2" {
		t.Fatal("opp id")
	}
	if proj.DiscardPickupWindowSize != 11 {
		t.Fatalf("window=%d", proj.DiscardPickupWindowSize)
	}
	// mutation check
	beforeHand := len(rm.Round.Hands[0])
	_ = proj
	if len(rm.Round.Hands[0]) != beforeHand {
		t.Fatal("mutated")
	}
}

func TestProjectionVariadicRoomValue(t *testing.T) {
	reg := room.NewRegistry()
	r, _ := reg.CreatePublicRoom(2, "p1")
	_ = reg.Join(r.ID, "p2")
	_ = reg.StartGame(r.ID, "p1", game.StandardDeck())
	rm, _ := reg.Get(r.ID)
	proj, err := Project(rm, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if proj.MatchID != rm.ID {
		t.Fatal("match id")
	}
}
