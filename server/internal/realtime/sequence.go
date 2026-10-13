package realtime

import "sync"

// Initial sequence value for a new recipient stream.
const InitialSequence = 0

// SequenceGap describes a detected missing sequence range.
type SequenceGap struct {
	LastReceived         uint64 `json:"last_received_sequence"`
	LatestServerSequence uint64 `json:"latest_server_sequence"`
	HasGap               bool   `json:"gap"`
}

// GapResult alias.
type GapResult = SequenceGap
type GapInfo = SequenceGap

type RecipientStream struct {
	mu           sync.Mutex
	lastReceived uint64
	latest       uint64
}

// RecipientStream aliases.
type RecipientSequenceStream = RecipientStream
type EventSequenceStream = RecipientStream
type SequenceStream = RecipientStream
type RecipientEventStream = RecipientStream
type Stream = RecipientStream

func NewRecipientStream() *RecipientStream {
	return &RecipientStream{lastReceived: InitialSequence, latest: InitialSequence}
}

// Constructor aliases.
func NewSequenceStream() *RecipientStream          { return NewRecipientStream() }
func NewEventStream() *RecipientStream             { return NewRecipientStream() }
func NewRecipientSequenceStream() *RecipientStream { return NewRecipientStream() }
func NewRecipientEventStream() *RecipientStream    { return NewRecipientStream() }
func NewStream() *RecipientStream                  { return NewRecipientStream() }
func NewSequence() *RecipientStream                { return NewRecipientStream() }

func (s *RecipientStream) LastReceivedSequence() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastReceived
}

func (s *RecipientStream) LatestSequence() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.latest
}

func (s *RecipientStream) Last() uint64         { return s.LastReceivedSequence() }
func (s *RecipientStream) Latest() uint64       { return s.LatestSequence() }
func (s *RecipientStream) Current() uint64      { return s.LatestSequence() }
func (s *RecipientStream) ServerLatest() uint64 { return s.LatestSequence() }

// NextSequence allocates the next monotonic sequence in this recipient stream.
func (s *RecipientStream) NextSequence() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest++
	return s.latest
}

// Aliases.
func (s *RecipientStream) Next() uint64         { return s.NextSequence() }
func (s *RecipientStream) Deliver() uint64      { return s.NextSequence() }
func (s *RecipientStream) DeliverEvent() uint64 { return s.NextSequence() }
func (s *RecipientStream) Assign() uint64       { return s.NextSequence() }
func (s *RecipientStream) Allocate() uint64     { return s.NextSequence() }
func (s *RecipientStream) NewSequence() uint64  { return s.NextSequence() }
func (s *RecipientStream) Advance() uint64      { return s.NextSequence() }
func (s *RecipientStream) Increment() uint64    { return s.NextSequence() }

// MarkReceived records a recipient's last received sequence without gap detection.
func (s *RecipientStream) MarkReceived(seq uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq > s.lastReceived {
		s.lastReceived = seq
	}
}

func (s *RecipientStream) Receive(seq uint64)            { s.MarkReceived(seq) }
func (s *RecipientStream) Ack(seq uint64)                { s.MarkReceived(seq) }
func (s *RecipientStream) Acknowledge(seq uint64)        { s.MarkReceived(seq) }
func (s *RecipientStream) UpdateLastReceived(seq uint64) { s.MarkReceived(seq) }
func (s *RecipientStream) SetLastReceived(seq uint64)    { s.MarkReceived(seq) }

// IsStale reports whether a sequence is at or below the last received sequence.
func (s *RecipientStream) IsStale(seq uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return seq <= s.lastReceived
}

func IsStaleSequence(lastReceived, seq uint64) bool { return seq <= lastReceived }

// DetectGap reports whether latest server sequence is ahead of last received by
// more than one position. It does not mutate state.
func (s *RecipientStream) DetectGap() SequenceGap {
	s.mu.Lock()
	defer s.mu.Unlock()
	gap := s.latest > s.lastReceived+1
	return SequenceGap{
		LastReceived:         s.lastReceived,
		LatestServerSequence: s.latest,
		HasGap:               gap,
	}
}

// Gap aliases.
func (s *RecipientStream) Gap() SequenceGap         { return s.DetectGap() }
func (s *RecipientStream) GapInfo() SequenceGap     { return s.DetectGap() }
func (s *RecipientStream) CheckGap() SequenceGap    { return s.DetectGap() }
func (s *RecipientStream) SequenceGap() SequenceGap { return s.DetectGap() }
func (s *RecipientStream) HasGap() bool             { return s.DetectGap().HasGap }

// DetectGapValues reports a gap from explicit values without a stream.
func DetectGapValues(lastReceived, latest uint64) SequenceGap {
	gap := latest > lastReceived+1
	return SequenceGap{LastReceived: lastReceived, LatestServerSequence: latest, HasGap: gap}
}

// SnapshotReset resets the recipient stream to the server's latest sequence.
func (s *RecipientStream) SnapshotReset(latest uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastReceived = latest
	if latest > s.latest {
		s.latest = latest
	}
}

// Aliases.
func (s *RecipientStream) Reset(latest uint64)             { s.SnapshotReset(latest) }
func (s *RecipientStream) ResetFromSnapshot(latest uint64) { s.SnapshotReset(latest) }
func (s *RecipientStream) Resync(latest uint64)            { s.SnapshotReset(latest) }
func (s *RecipientStream) ApplySnapshot(latest uint64)     { s.SnapshotReset(latest) }
func (s *RecipientStream) Snapshot(latest uint64)          { s.SnapshotReset(latest) }

// Stream manager for per room/recipient.
type StreamManager struct {
	mu      sync.Mutex
	streams map[string]*RecipientStream
}

func NewStreamManager() *StreamManager {
	return &StreamManager{streams: make(map[string]*RecipientStream)}
}

func (m *StreamManager) StreamKey(roomID, recipientID string) string {
	return roomID + ":" + recipientID
}

func (m *StreamManager) Get(roomID, recipientID string) *RecipientStream {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.StreamKey(roomID, recipientID)
	if s, ok := m.streams[key]; ok {
		return s
	}
	s := NewRecipientStream()
	m.streams[key] = s
	return s
}

func (m *StreamManager) Stream(roomID, recipientID string) *RecipientStream {
	return m.Get(roomID, recipientID)
}

func (m *StreamManager) NewStream(roomID, recipientID string) *RecipientStream {
	return m.Get(roomID, recipientID)
}

// RegisterManager aliases type.
type RecipientStreamManager = StreamManager
type SequenceManager = StreamManager
type EventSequenceManager = StreamManager
