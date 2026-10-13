package realtime

import (
	"encoding/json"
	"testing"

	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"
)

func TestContractVisibilityAndPublicState(t *testing.T) {
	reg := room.NewRegistry()
	r, _ := reg.CreatePublicRoom(3, "p1")
	_ = reg.Join(r.ID, "p2")
	_ = reg.Join(r.ID, "p3")
	_ = reg.StartGame(r.ID, "p1", game.StandardDeck())
	rm, _ := reg.Get(r.ID)
	proj, err := Project(&rm, rm.Round, "p1")
	if err != nil {
		t.Fatal(err)
	}

	// own_hand only requesting player's cards
	if len(proj.OwnHand) != len(rm.Round.Hands[0]) {
		t.Fatal("own hand count")
	}

	// opponents hand sizes only (no card details in JSON)
	raw, _ := json.Marshal(proj.Opponents)
	var opps []map[string]interface{}
	_ = json.Unmarshal(raw, &opps)
	for _, o := range opps {
		if _, ok := o["hand_size"]; !ok {
			t.Fatal("missing hand_size")
		}
		for _, forbidden := range []string{"card_ids", "cards", "rank", "suit", "is_active_joker", "hand"} {
			if _, ok := o[forbidden]; ok {
				t.Fatalf("forbidden field %s in opponent", forbidden)
			}
		}
	}
	// public fields available
	if proj.ActiveJokerRank == "" {
		t.Fatal("missing joker rank")
	}
	if proj.CurrentPlayerID == "" {
		t.Fatal("missing current player")
	}
	if proj.Phase == "" {
		t.Fatal("missing phase")
	}
	if proj.ClosedDeckCount == 0 {
		t.Fatal("missing closed count")
	}
	if proj.DiscardPickupWindowSize == 0 {
		t.Fatal("missing window")
	}
	// opponent count
	if len(proj.Opponents) != 2 {
		t.Fatalf("opponents=%d", len(proj.Opponents))
	}
}

func TestContractResyncMetadata(t *testing.T) {
	reg := room.NewRegistry()
	r, _ := reg.CreatePublicRoom(2, "p1")
	_ = reg.Join(r.ID, "p2")
	_ = reg.StartGame(r.ID, "p1", game.StandardDeck())
	rm, _ := reg.Get(r.ID)
	stream := NewRecipientStream()
	seq := stream.NextSequence()
	res, err := Resync(&rm, rm.Round, "p1", seq, stream)
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "snapshot" {
		t.Fatal("mode")
	}
	if res.LatestSequence == 0 {
		t.Fatal("latest")
	}
	if res.ReplacedSinceSequence != seq {
		t.Fatal("replaced")
	}
	// projection equivalence
	live, _ := Project(&rm, rm.Round, "p1")
	liveRaw, _ := json.Marshal(live)
	resRaw, _ := json.Marshal(res.Snapshot)
	if string(liveRaw) != string(resRaw) {
		t.Fatal("projection mismatch")
	}
}
