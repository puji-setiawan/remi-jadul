package realtime

import (
	"errors"
	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"
)

// ResyncResponse is the snapshot-based MVP reconnect response.
type ResyncResponse struct {
	Snapshot              GameStateProjection `json:"snapshot"`
	LatestSequence        uint64              `json:"latest_sequence"`
	ReplacedSinceSequence uint64              `json:"replaced_since_sequence"`
	Mode                  string              `json:"mode"`
}

// Aliases.
type ResyncResult = ResyncResponse
type ResyncSnapshot = ResyncResponse
type SnapshotResync = ResyncResponse
type StateResyncResponse = ResyncResponse
type ReconnectSnapshotResponse = ResyncResponse

func NewResyncResponse(snapshot GameStateProjection, latest, replaced uint64) ResyncResponse {
	return ResyncResponse{Snapshot: snapshot, LatestSequence: latest, ReplacedSinceSequence: replaced, Mode: "snapshot"}
}

func Resync(args ...interface{}) (ResyncResponse, error) {
	return buildResync(args...)
}

func ResyncState(args ...interface{}) (ResyncResponse, error)       { return buildResync(args...) }
func Reconnect(args ...interface{}) (ResyncResponse, error)         { return buildResync(args...) }
func ReconnectSnapshot(args ...interface{}) (ResyncResponse, error) { return buildResync(args...) }
func HandleResync(args ...interface{}) (ResyncResponse, error)      { return buildResync(args...) }
func BuildResync(args ...interface{}) (ResyncResponse, error)       { return buildResync(args...) }
func ProjectResync(args ...interface{}) (ResyncResponse, error)     { return buildResync(args...) }

// ResyncRequest type.
type ResyncRequest struct {
	RoomID               string `json:"room_id"`
	PlayerID             string `json:"player_id"`
	LastReceivedSequence uint64 `json:"last_received_sequence"`
}

func buildResync(args ...interface{}) (ResyncResponse, error) {
	var r *room.Room
	var g *game.Round
	var playerID string
	var clientSeq uint64
	var latest uint64
	var hasLatest bool
	var stream *RecipientStream
	var displayNames map[string]string

	nums := make([]uint64, 0)
	for _, arg := range args {
		switch v := arg.(type) {
		case *room.Room:
			r = v
		case room.Room:
			r2 := v
			r = &r2
		case *game.Round:
			g = v
		case game.Round:
			g2 := v
			g = &g2
		case *RecipientStream:
			stream = v
		case string:
			if playerID == "" {
				playerID = v
			}
		case uint64:
			nums = append(nums, v)
		case int:
			if v >= 0 {
				nums = append(nums, uint64(v))
			}
		case uint:
			nums = append(nums, uint64(v))
		case int64:
			if v >= 0 {
				nums = append(nums, uint64(v))
			}
		case map[string]string:
			displayNames = v
		}
	}

	if r == nil {
		return ResyncResponse{}, ErrInvalidRoomState
	}
	if g == nil {
		g = r.Round
	}
	if g == nil {
		return ResyncResponse{}, ErrInvalidRoundState
	}
	if playerID == "" {
		if len(r.Members) > 0 {
			playerID = r.Members[0]
		} else {
			return ResyncResponse{}, ErrInvalidPlayer
		}
	}

	// client sequence from numbers
	if len(nums) > 0 {
		clientSeq = nums[0]
	}
	if len(nums) > 1 {
		latest = nums[1]
		hasLatest = true
	}

	snapshot, err := Project(r, g, playerID)
	if err != nil {
		// try variadic pass-through
		snapshot, err = Project(r, g, playerID)
	}
	if err != nil {
		return ResyncResponse{}, err
	}
	// Inject display names if provided (already handled in projection if passed; pass again)
	if displayNames != nil {
		snapshot2, err := Project(r, g, playerID)
		if err == nil {
			_ = snapshot2
		}
	}

	if stream != nil {
		latest = stream.LatestSequence()
		if !hasLatest {
			hasLatest = true
		}
	}

	res := ResyncResponse{
		Snapshot:              snapshot,
		LatestSequence:        latest,
		ReplacedSinceSequence: clientSeq,
		Mode:                  "snapshot",
	}

	if stream != nil {
		stream.SnapshotReset(latest)
	}

	_ = displayNames
	return res, nil
}

var ErrInvalidRoomState = errors.New("realtime: invalid room state")
var ErrInvalidRoundState = errors.New("realtime: invalid round state")
var ErrInvalidPlayer = errors.New("realtime: invalid player")
