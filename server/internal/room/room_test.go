package room

import (
	"errors"
	"testing"

	"remi-jadul/server/internal/game"
)

func TestRoomLifecycle(t *testing.T) {
	reg := NewRegistry()
	// invalid capacity
	if _, err := reg.CreateRoom("public", 5, "host"); err != ErrInvalidCapacity {
		t.Fatalf("capacity err=%v", err)
	}
	// public create
	room, err := reg.CreatePublicRoom(2, "host")
	if err != nil {
		t.Fatal(err)
	}
	if room.Visibility != "public" {
		t.Fatal("visibility")
	}
	if room.AccessCode != "" {
		t.Fatal("public code leak")
	}
	// private
	proom, err := reg.CreatePrivateRoom(2, "host2", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if proom.AccessCode != "secret" {
		t.Fatal("private code")
	}
	// discovery no code
	for _, r := range reg.PublicRooms() {
		if r.AccessCode != "" {
			t.Fatal("discovery code leak")
		}
	}
	// join public
	if err := reg.Join(room.ID, "p2"); err != nil {
		t.Fatal(err)
	}
	// duplicate
	if err := reg.Join(room.ID, "p2"); err != ErrAlreadyJoined {
		t.Fatalf("dup err=%v", err)
	}
	// capacity
	if err := reg.Join(room.ID, "p3"); err != ErrRoomFull {
		t.Fatalf("full err=%v", err)
	}
	// private wrong code
	if err := reg.Join(proom.ID, "p4", "bad"); err != ErrIncorrectAccessCode {
		t.Fatalf("code err=%v", err)
	}
	// private right code (and discovery excludes full? no join yet)
	if err := reg.Join(proom.ID, "p4", "secret"); err != nil {
		t.Fatal(err)
	}
	// join reorder args test: room full already so duplicate/rejected join
	err = reg.Join("pX", room.ID)
	if err != ErrRoomFull {
		t.Fatalf("reorder join err=%v", err)
	}
	// leave non-host removes room if empty? Host+ p2; remove p2 leaves host only (not empty)
	if err := reg.Leave(room.ID, "p2"); err != nil {
		t.Fatal(err)
	}
	// host leaves closes room
	closedID := room.ID
	if err := reg.Leave(room.ID, "host"); err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.Get(closedID); ok {
		t.Fatal("room not removed after host leave")
	}
	// empty room removed
	solo, _ := reg.CreatePublicRoom(2, "solo")
	if err := reg.Leave(solo.ID, "solo"); err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.Get(solo.ID); ok {
		t.Fatal("empty room not removed")
	}
	// no mutation on rejection (capacity case already didn't mutate full)
}

func TestApplyRoundTransitionRestoresExistingMatchOnCallbackFailure(t *testing.T) {
	reg := NewRegistry()
	rm, err := reg.CreatePublicRoom(2, "host")
	if err != nil {
		t.Fatal(err)
	}
	match := game.NewMatch(2)
	match.RoundNumber = 2
	match.CumulativeScores = []int{7, 11}
	if err := reg.SetMatch(rm.ID, match); err != nil {
		t.Fatal(err)
	}
	oldNext, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("next-round setup failed")
	if err := reg.ApplyRoundTransition(rm.ID, oldNext, func(stored *Room) error {
		stored.Match.RoundNumber = 3
		stored.Match.CumulativeScores[0] = 99
		stored.Match.Ended = true
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("transition error=%v, want %v", err, wantErr)
	}
	got, ok := reg.Get(rm.ID)
	if !ok || got.Match != match {
		t.Fatalf("match pointer changed: got=%p want=%p", got.Match, match)
	}
	if match.RoundNumber != 2 || match.Ended || match.CumulativeScores[0] != 7 {
		t.Fatalf("match was not rolled back: %#v", match)
	}
	if got.Round != nil {
		t.Fatalf("round installed after failed transition: %#v", got.Round)
	}
}
