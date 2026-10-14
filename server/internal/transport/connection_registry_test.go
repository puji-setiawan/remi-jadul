package transport

import (
	"sync"
	"testing"

	"remi-jadul/server/internal/realtime"
)

func TestConnectionRegistry(t *testing.T) {
	reg := NewConnectionRegistry()
	c1 := &Connection{GuestID: "g1", Stream: realtime.NewRecipientStream(), Token: "t1"}
	c2 := &Connection{GuestID: "g1", Stream: realtime.NewRecipientStream(), Token: "t2"}
	reg.Register(c1)
	reg.Register(c2)
	if reg.Count() != 1 {
		t.Fatalf("count=%d", reg.Count())
	}
	got, ok := reg.GetByGuest("g1")
	if !ok || got.Token != "t2" {
		t.Fatalf("replacement failed")
	}
	// unregister by token
	if !reg.UnregisterToken("t2") {
		t.Fatal("unregister token")
	}
	if reg.Count() != 0 {
		t.Fatal("not removed")
	}
	// missing recipient broadcast
	reg.Register(&Connection{GuestID: "g2", Token: "t3", Stream: realtime.NewRecipientStream()})
	var called []string
	missing := reg.Broadcast([]string{"g2", "nope"}, func(id string, c *Connection) error {
		called = append(called, id)
		return nil
	})
	if len(called) != 1 || len(missing) != 1 {
		t.Fatalf("called=%v missing=%v", called, missing)
	}
}

func TestConnectionRegistryConcurrent(t *testing.T) {
	reg := NewConnectionRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "g"
			reg.Register(&Connection{GuestID: id, Token: "t", Stream: realtime.NewRecipientStream()})
			reg.GetByGuest(id)
			reg.Count()
		}(i)
	}
	wg.Wait()
}
