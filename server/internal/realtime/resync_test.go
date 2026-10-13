package realtime

import (
	"testing"

	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"
)

func TestResync(t *testing.T) {
	reg := room.NewRegistry()
	r, _ := reg.CreatePublicRoom(2, "p1")
	_ = reg.Join(r.ID, "p2")
	_ = reg.StartGame(r.ID, "p1", game.StandardDeck())
	rm, _ := reg.Get(r.ID)
	stream := NewRecipientStream()
	seq := stream.NextSequence()
	res, err := Resync(&rm, rm.Round, "p2", seq, stream)
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "snapshot" {
		t.Fatalf("mode=%s", res.Mode)
	}
	if res.ReplacedSinceSequence != seq {
		t.Fatalf("replaced=%d", res.ReplacedSinceSequence)
	}
	if res.LatestSequence != stream.LastReceivedSequence() {
		t.Fatal("latest not reset")
	}
	// equality with live projection
	live, _ := Project(&rm, rm.Round, "p2")
	if res.Snapshot.OwnHand[0].CardID != live.OwnHand[0].CardID {
		t.Fatal("projection mismatch")
	}
	// no mutation to hand length
	if len(rm.Round.Hands[1]) == 0 {
		t.Fatal("mutated")
	}
	// equal sequence handling
	res2, err := Resync(&rm, rm.Round, "p2", res.LatestSequence)
	if err != nil {
		t.Fatal(err)
	}
	if res2.ReplacedSinceSequence != res.LatestSequence {
		t.Fatal("equal seq mishandled")
	}
}
