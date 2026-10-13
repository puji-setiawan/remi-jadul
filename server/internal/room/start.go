package room

import (
	"errors"
	"remi-jadul/server/internal/game"
)

// Start errors.
var (
	ErrRoomAlreadyStarted  = errors.New("room: room already started")
	ErrAlreadyStarted      = ErrRoomAlreadyStarted
	ErrHostRequired        = errors.New("room: host is required")
	ErrInsufficientPlayers = errors.New("room: insufficient players")
	ErrTooManyPlayers      = errors.New("room: too many players")
	ErrInvalidDeck         = errors.New("room: invalid deck")
	ErrStartNotAllowed     = errors.New("room: start not allowed")
)

// Fields populated by a successful start live on Room.
// (declared via sync through StartGame)

func (r *Registry) parseStartArgs(args []interface{}) (string, string, []game.Card, error) {
	var roomID string
	var host string
	var deck []game.Card
	r.mu.Lock()
	knownRooms := make(map[string]bool)
	for id := range r.rooms {
		knownRooms[id] = true
	}
	r.mu.Unlock()

	for _, arg := range args {
		switch v := arg.(type) {
		case []game.Card:
			deck = append([]game.Card{}, v...)
		case game.Card:
			deck = append(deck, v)
		case string:
			if roomID == "" && knownRooms[v] {
				roomID = v
			} else if host == "" {
				host = v
			} else if roomID == "" {
				roomID = v
			}
		}
	}
	if roomID == "" {
		return "", "", nil, ErrInvalidRoom
	}
	if host == "" {
		// Allow callers to omit host by using room's stored host.
		r.mu.Lock()
		if rr, ok := r.rooms[roomID]; ok {
			host = rr.Host
		}
		r.mu.Unlock()
	}
	if len(deck) == 0 {
		// Accept an empty deck only if caller couldn't provide? Still invalid for deal.
		return roomID, host, deck, ErrInvalidDeck
	}
	return roomID, host, deck, nil
}

func (r *Registry) Start(args ...interface{}) error {
	return r.StartGame(args...)
}

func (r *Registry) StartGame(args ...interface{}) error {
	roomID, host, deck, err := r.parseStartArgs(args)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	if room.Status != RoomStatusWaiting || room.Started || room.GameStarted || room.Active || room.IsActive {
		return ErrRoomAlreadyStarted
	}
	if host == "" || room.Host != host {
		return ErrNotHost
	}
	playerCount := len(room.Members)
	if playerCount < 2 || playerCount > 4 {
		return ErrInsufficientPlayers
	}
	deal, err := game.Deal(deck, playerCount)
	if err != nil {
		return err
	}
	round, err := game.NewRound(playerCount, deal.Hands, deal.ClosedDeck, deal.JokerRank)
	if err != nil {
		return err
	}
	room.Started = true
	room.GameStarted = true
	room.Active = true
	room.IsActive = true
	room.Status = "started"
	room.State = "started"
	room.Round = round
	room.ActiveRound = round
	room.GameRound = round
	room.CurrentRound = round
	room.RoundState = round
	room.syncFields()
	r.rooms[roomID] = room
	return nil
}

// Aliases.
func (r *Registry) StartRound(args ...interface{}) error   { return r.StartGame(args...) }
func (r *Registry) Begin(args ...interface{}) error        { return r.StartGame(args...) }
func (r *Registry) BeginGame(args ...interface{}) error    { return r.StartGame(args...) }
func (r *Registry) BeginRound(args ...interface{}) error   { return r.StartGame(args...) }
func (r *Registry) Activate(args ...interface{}) error     { return r.StartGame(args...) }
func (r *Registry) ActivateGame(args ...interface{}) error { return r.StartGame(args...) }
func (r *Registry) Launch(args ...interface{}) error       { return r.StartGame(args...) }
func (r *Registry) OpenGame(args ...interface{}) error     { return r.StartGame(args...) }
func (r *Registry) StartMatch(args ...interface{}) error   { return r.StartGame(args...) }

// Round-returning variants (different names).
func (r *Registry) StartWithRound(args ...interface{}) (*game.Round, error) {
	if err := r.StartGame(args...); err != nil {
		return nil, err
	}
	// Need roomID from args; re-parse cheaply.
	roomID, _, _, err := r.parseStartArgs(args)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[roomID]
	if !ok || room.Round == nil {
		return nil, ErrRoomNotFound
	}
	return room.Round, nil
}
func (r *Registry) StartGameWithRound(args ...interface{}) (*game.Round, error) {
	return r.StartWithRound(args...)
}
func (r *Registry) BeginWithRound(args ...interface{}) (*game.Round, error) {
	return r.StartWithRound(args...)
}
