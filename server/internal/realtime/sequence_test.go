package realtime

import "testing"

func TestSequenceStreams(t *testing.T) {
	s := NewRecipientStream()
	if s.NextSequence() != 1 {
		t.Fatalf("first=%d", 1)
	}
	if s.NextSequence() != 2 {
		t.Fatalf("second=%d", 2)
	}
	if s.IsStale(1) {
		t.Fatal("new event stale? 1<=2 is stale, ok true actually")
	}
	// stale at/below last received
	s.MarkReceived(2)
	if !s.IsStale(2) {
		t.Fatal("should be stale")
	}
	if s.IsStale(3) {
		t.Fatal("future not stale")
	}
	// gap detection with latest ahead
	s2 := NewRecipientStream()
	_ = s2.NextSequence() //1
	_ = s2.NextSequence() //2
	_ = s2.NextSequence() //3
	s2.MarkReceived(1)
	gap := s2.DetectGap()
	if !gap.HasGap {
		t.Fatal("expected gap")
	}
	if gap.LastReceived != 1 || gap.LatestServerSequence != 3 {
		t.Fatalf("gap=%+v", gap)
	}
	// no gap
	s3 := NewRecipientStream()
	_ = s3.NextSequence()
	s3.MarkReceived(1)
	if s3.DetectGap().HasGap {
		t.Fatal("unexpected gap")
	}
	// snapshot reset
	s3.SnapshotReset(5)
	if s3.LastReceivedSequence() != 5 {
		t.Fatal("reset last")
	}
	// per-recipient independence
	m := NewStreamManager()
	a := m.Get("room", "p1")
	b := m.Get("room", "p2")
	_ = a.NextSequence()
	if b.LatestSequence() != 0 {
		t.Fatal("shared state")
	}
}
