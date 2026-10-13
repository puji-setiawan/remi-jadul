package room

import (
	"testing"

	"remi-jadul/server/internal/game"
)

func TestHostStartRound(t *testing.T) {
	reg := NewRegistry()
	room, err := reg.CreatePublicRoom(2, "host")
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Join(room.ID, "p2"); err != nil {
		t.Fatal(err)
	}
	deck := game.StandardDeck()
	if err := reg.StartGame(room.ID, "host", deck); err != nil {
		t.Fatalf("start err=%v", err)
	}
	got, ok := reg.Get(room.ID)
	if !ok {
		t.Fatal("room gone")
	}
	if !got.Started {
		t.Fatal("not started")
	}
	if got.Round == nil {
		t.Fatal("round nil")
	}
	if got.Round.PlayerCount != 2 {
		t.Fatalf("players=%d", got.Round.PlayerCount)
	}
	if len(got.Round.Hands[0]) != 11 {
		t.Fatalf("hand=%d", len(got.Round.Hands[0]))
	}
	if got.Round.JokerRank == 0 {
		t.Fatal("joker not set")
	}
	// non-host
	room2, _ := reg.CreatePublicRoom(2, "host2")
	if err := reg.Join(room2.ID, "q2"); err != nil {
		t.Fatal(err)
	}
	if err := reg.StartGame(room2.ID, "not-host", deck); err != ErrNotHost {
		t.Fatalf("err=%v", err)
	}
	// already started
	if err := reg.StartGame(room.ID, "host", deck); err != ErrRoomAlreadyStarted {
		t.Fatalf("err=%v", err)
	}
	// insufficient players
	room3, _ := reg.CreatePublicRoom(2, "host3")
	if err := reg.StartGame(room3.ID, "host3", deck); err != ErrInsufficientPlayers {
		t.Fatalf("err=%v", err)
	}
	// no mutation on rejection
	got3, _ := reg.Get(room3.ID)
	if got3.Started {
		t.Fatal("insufficient start mutated")
	}
	// round-returning variant
	rr, err := reg.StartGameWithRound(room3.ID, "host3", deck)
	if err == nil {
		t.Fatal("expected insufficient err")
	}
	_ = rr
}
