package auth

import "testing"

func TestGuestCreateLookup(t *testing.T) {
	store := NewGuestStore()
	g, err := store.CreateGuest("  Raka  ")
	if err != nil {
		t.Fatal(err)
	}
	if g.DisplayName != "Raka" {
		t.Fatalf("name=%q", g.DisplayName)
	}
	if g.ID == "" {
		t.Fatal("empty id")
	}
	found, err := store.GetGuest(g.ID)
	if err != nil || found.ID != g.ID {
		t.Fatalf("lookup err=%v", err)
	}
	// unique
	g2, _ := store.CreateGuest("Nina")
	if g2.ID == g.ID {
		t.Fatal("duplicate id")
	}
	// blank
	if _, err := store.CreateGuest("   "); err == nil {
		t.Fatal("expected blank error")
	}
	// unknown
	if _, err := store.GetGuest("nope"); err != ErrGuestNotFound {
		t.Fatalf("err=%v", err)
	}
}
