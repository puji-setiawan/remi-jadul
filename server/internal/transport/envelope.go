package transport

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

// Protocol version.
const ProtocolVersion = "v1"

var (
	ErrInvalidEnvelope  = errors.New("transport: invalid envelope")
	ErrInvalidCommand   = errors.New("transport: invalid command")
	ErrUnknownCommand   = errors.New("transport: unknown command")
	ErrUnknownMessage   = errors.New("transport: unknown message")
	ErrMissingRoomID    = errors.New("transport: missing room id")
	ErrMissingRequestID = errors.New("transport: missing request id")
)

// Known v1 command names.
var KnownCommands = map[string]bool{
	"room.create":       true,
	"room.join":         true,
	"room.leave":        true,
	"game.start":        true,
	"game.draw":         true,
	"game.take_discard": true,
	"game.laydown":      true,
	"game.discard":      true,
	"game.close":        true,
	"state.resync":      true,
}

var RoomScopedCommands = map[string]bool{
	"room.join":         true,
	"room.leave":        true,
	"game.start":        true,
	"game.draw":         true,
	"game.take_discard": true,
	"game.laydown":      true,
	"game.discard":      true,
	"game.close":        true,
	"state.resync":      true,
}

// ActionError is a safe typed action error.
type ActionError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func (e ActionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func (e ActionError) Is(target error) bool {
	if te, ok := target.(ActionError); ok {
		return te.Code == e.Code
	}
	return false
}

func NewActionError(code string) ActionError {
	return ActionError{Code: code, Message: safeMessage(code), Retryable: retryable(code)}
}

func safeMessage(code string) string {
	switch code {
	case "invalid_command", "unknown_message":
		return "Unknown message type or command."
	case "invalid_request":
		return "Request is invalid."
	case "missing_version":
		return "Protocol version is required."
	case "missing_type":
		return "Message type is required."
	case "missing_name":
		return "Message name is required."
	case "missing_request_id":
		return "Request ID is required."
	case "missing_room_id":
		return "Room ID is required."
	default:
		return "Action could not be processed."
	}
}

func retryable(code string) bool {
	switch code {
	case "stale_request":
		return true
	default:
		return false
	}
}

// Envelope is the versioned WebSocket command envelope.
type CommandEnvelope struct {
	Version   string          `json:"version"`
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	RequestID string          `json:"request_id"`
	RoomID    string          `json:"room_id,omitempty"`
	Sequence  uint64          `json:"sequence,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// Aliases.
type Envelope = CommandEnvelope
type Command = CommandEnvelope
type CommandMessage = CommandEnvelope
type ClientCommand = CommandEnvelope
type WSEnvelope = CommandEnvelope
type WebSocketEnvelope = CommandEnvelope

// Known aliases for type.
const (
	CommandType = "command"
	EventType   = "event"
	ErrorType   = "error"
)

func (e CommandEnvelope) IsCommand() bool { return e.Type == CommandType }
func (e CommandEnvelope) HasRoomID() bool { return e.RoomID != "" }

// DecodeCommand parses and validates a command envelope.
func DecodeCommand(data []byte) (CommandEnvelope, ActionError) {
	var env CommandEnvelope
	dec := json.NewDecoder(bytesReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		// Unknown fields must be ignored, so fall back to lenient decode.
		_ = err
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return CommandEnvelope{}, NewActionError("invalid_request")
	}
	if ae := ValidateCommand(env); ae.Code != "" {
		return env, ae
	}
	return env, ActionError{}
}

// ValidateCommand validates envelope shape and request ID.
func ValidateCommand(env CommandEnvelope) ActionError {
	if env.Version == "" {
		return NewActionError("missing_version")
	}
	if env.Version != ProtocolVersion {
		return NewActionError("invalid_request")
	}
	if env.Type == "" {
		return NewActionError("missing_type")
	}
	if env.Type != "command" {
		return NewActionError("invalid_request")
	}
	if env.Name == "" {
		return NewActionError("missing_name")
	}
	if !KnownCommands[env.Name] {
		return NewActionError("invalid_command")
	}
	if env.RequestID == "" {
		return NewActionError("missing_request_id")
	}
	if RoomScopedCommands[env.Name] && env.RoomID == "" {
		return NewActionError("missing_room_id")
	}
	return ActionError{}
}

// Alias
func ValidateEnvelope(env CommandEnvelope) ActionError { return ValidateCommand(env) }
func Validate(env CommandEnvelope) ActionError         { return ValidateCommand(env) }
func IsValidCommand(env CommandEnvelope) bool          { return ValidateCommand(env).Code == "" }

func IsKnownCommand(name string) bool { return KnownCommands[name] }
func IsRoomScoped(name string) bool   { return RoomScopedCommands[name] }

// Idempotency helpers.
type RequestTracker struct {
	mu    sync.Mutex
	seen  map[string]struct{}
	order []string
}

func NewRequestTracker() *RequestTracker {
	return &RequestTracker{seen: make(map[string]struct{})}
}

func (t *RequestTracker) Seen(requestID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.seen[requestID]
	return ok
}

func (t *RequestTracker) Remember(requestID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.seen[requestID]; !ok {
		t.seen[requestID] = struct{}{}
		t.order = append(t.order, requestID)
	}
}

func (t *RequestTracker) ForgetRequest(requestID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.seen, requestID)
}

func (t *RequestTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.seen = make(map[string]struct{})
	t.order = nil
}

// RequestID type.
type RequestID = string

func ValidRequestID(id string) bool { return id != "" }

// Unknown fields ignored: json.Unmarshal already ignores.
// Helper to ensure no hidden data in action error.
func (e ActionError) Safe() bool {
	msg := strings.ToLower(e.Message)
	for _, forbidden := range []string{"card", "hand", "marker", "deck order", "hand size"} {
		if containsSubstring(msg, forbidden) {
			return false
		}
	}
	return true
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	if sub == "" {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Stable code constants.
const (
	CodeUnauthorized         = "unauthorized"
	CodeInvalidRequest       = "invalid_request"
	CodeNotFound             = "not_found"
	CodeRoomFull             = "room_full"
	CodeRoomPrivate          = "room_private"
	CodeNotHost              = "not_host"
	CodeGameNotStarted       = "game_not_started"
	CodeNotYourTurn          = "not_your_turn"
	CodeInvalidPhase         = "invalid_phase"
	CodeCardNotOwned         = "card_not_owned"
	CodeUnknownCard          = "unknown_card"
	CodeInvalidCombination   = "invalid_combination"
	CodeInvalidDiscardWindow = "invalid_discard_window"
	CodeInvalidJago          = "invalid_jago"
	CodeStaleRequest         = "stale_request"
	CodeSequenceGap          = "sequence_gap"
	CodeGameOver             = "game_over"
	CodeUnknownMessage       = "unknown_message"
	CodeInvalidCommand       = "invalid_command"
	CodeInvalidRoomID        = "missing_room_id"
	CodeMissingVersion       = "missing_version"
	CodeMissingType          = "missing_type"
	CodeMissingName          = "missing_name"
	CodeMissingRequestID     = "missing_request_id"
	CodeMissingRoomIDAlt     = "missing_room_id"
)
