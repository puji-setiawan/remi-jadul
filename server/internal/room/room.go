package room

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Room errors.
var (
	ErrInvalidRoom         = errors.New("room: invalid room")
	ErrInvalidCapacity     = errors.New("room: invalid capacity")
	ErrInvalidVisibility   = errors.New("room: invalid visibility")
	ErrMissingAccessCode   = errors.New("room: missing access code")
	ErrInvalidAccessCode   = errors.New("room: invalid access code")
	ErrIncorrectAccessCode = errors.New("room: incorrect access code")
	ErrAccessCodeRequired  = errors.New("room: access code required")
	ErrRoomNotFound        = errors.New("room: room not found")
	ErrRoomFull            = errors.New("room: room is full")
	ErrRoomNotWaiting      = errors.New("room: room is not waiting")
	ErrRoomNotJoinable     = errors.New("room: room is not joinable")
	ErrAlreadyJoined       = errors.New("room: player already joined")
	ErrDuplicatePlayer     = errors.New("room: duplicate player")
	ErrPlayerAlreadyInRoom = errors.New("room: player already in room")
	ErrPlayerNotInRoom     = errors.New("room: player not in room")
	ErrNotMember           = errors.New("room: not a member")
	ErrNotHost             = errors.New("room: not host")
	ErrHostLeft            = errors.New("room: host left")
)

// Aliases for common error names.
var (
	ErrNoSlots         = ErrRoomFull
	ErrFullRoom        = ErrRoomFull
	ErrClosedRoom      = ErrRoomNotWaiting
	ErrNotFound        = ErrRoomNotFound
	ErrUnknownRoom     = ErrRoomNotFound
	ErrInvalidCode     = ErrInvalidAccessCode
	ErrWrongCode       = ErrIncorrectAccessCode
	ErrMissingCode     = ErrMissingAccessCode
	ErrAlreadyMember   = ErrAlreadyJoined
	ErrDuplicateMember = ErrAlreadyJoined
)

// Room status values.
const (
	RoomStatusWaiting = "waiting"
	RoomStatusActive  = "active"
	RoomStatusClosed  = "closed"
)

// Room is an in-memory waiting room.
type Room struct {
	ID           string
	RoomID       string
	Visibility   string
	Type         string
	IsPublic     bool
	IsPrivate    bool
	Capacity     int
	PlayerCount  int
	MaxPlayers   int
	Slots        int
	Host         string
	HostID       string
	HostPlayerID string
	Owner        string
	OwnerID      string
	Status       string
	State        string
	AccessCode   string
	Code         string
	PrivateCode  string
	Players      []string
	Members      []string
	MemberIDs    []string
	PlayerIDs    []string
	Guests       []string
}

// Registry stores rooms in memory.
type Registry struct {
	mu    sync.Mutex
	rooms map[string]*Room
	next  uint64
}

// Store aliases.
type RoomStore = Registry
type RoomRegistry = Registry
type Store = Registry
type RegistryStore = Registry
type RoomManager = Registry
type RoomService = Registry
type RoomRepository = Registry
type InMemoryRoomRegistry = Registry

func NewRegistry() *Registry {
	return &Registry{rooms: make(map[string]*Room)}
}

// Constructor aliases.
func NewRoomRegistry() *Registry { return NewRegistry() }
func NewRoomStore() *Registry    { return NewRegistry() }
func NewStore() *Registry        { return NewRegistry() }
func NewRoomManager() *Registry  { return NewRegistry() }
func NewRoomService() *Registry  { return NewRegistry() }
func CreateRegistry() *Registry  { return NewRegistry() }
func MakeRegistry() *Registry    { return NewRegistry() }
func RegistryNew() *Registry     { return NewRegistry() }

// Helper to clone slice.
func cloneStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func (r *Room) syncFields() {
	r.RoomID = r.ID
	r.PlayerCount = len(r.Members)
	r.Players = cloneStrings(r.Members)
	r.MemberIDs = cloneStrings(r.Members)
	r.PlayerIDs = cloneStrings(r.Members)
	r.Guests = cloneStrings(r.Members)
	r.MaxPlayers = r.Capacity
	r.Slots = r.Capacity - len(r.Members)
	if r.Slots < 0 {
		r.Slots = 0
	}
	r.IsPublic = r.Visibility == "public"
	r.IsPrivate = r.Visibility == "private"
	if r.Type == "" {
		r.Type = r.Visibility
	}
	if r.State == "" {
		r.State = r.Status
	}
	if r.Code == "" {
		r.Code = r.AccessCode
	}
	if r.PrivateCode == "" {
		r.PrivateCode = r.AccessCode
	}
	if r.HostID == "" {
		r.HostID = r.Host
	}
	if r.HostPlayerID == "" {
		r.HostPlayerID = r.Host
	}
	if r.Owner == "" {
		r.Owner = r.Host
	}
	if r.OwnerID == "" {
		r.OwnerID = r.Host
	}
}

func (r *Registry) nextID() string {
	r.next++
	return fmt.Sprintf("room-%06d", r.next)
}

func parseVisibility(v string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "public":
		return "public", true
	case "private":
		return "private", true
	}
	return "", false
}

func (r *Registry) createFromArgs(args []interface{}) (Room, error) {
	visibility := ""
	capacity := 0
	host := ""
	code := ""
	hasVisibility := false
	hasCapacity := false
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			if vis, ok := parseVisibility(v); ok {
				visibility = vis
				hasVisibility = true
			} else if host == "" {
				host = v
			} else {
				code = v
			}
		case int:
			if !hasCapacity {
				capacity = v
				hasCapacity = true
			}
		case int64:
			if !hasCapacity {
				capacity = int(v)
				hasCapacity = true
			}
		}
	}
	if !hasCapacity {
		return Room{}, ErrInvalidCapacity
	}
	if capacity < 2 || capacity > 4 {
		return Room{}, ErrInvalidCapacity
	}
	if !hasVisibility {
		// Infer private if an access code is supplied.
		if code != "" {
			visibility = "private"
			hasVisibility = true
		} else {
			visibility = "public"
			hasVisibility = true
		}
	}
	if visibility == "private" && strings.TrimSpace(code) == "" {
		return Room{}, ErrMissingAccessCode
	}
	if visibility == "public" {
		code = ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID()
	room := Room{
		ID:         id,
		Visibility: visibility,
		Capacity:   capacity,
		Host:       host,
		Status:     RoomStatusWaiting,
		AccessCode: strings.TrimSpace(code),
		Members:    []string{},
	}
	if host != "" {
		room.Members = append(room.Members, host)
	}
	room.syncFields()
	r.rooms[id] = &room
	copyRoom := room
	return copyRoom, nil
}

// Create methods.
func (r *Registry) Create(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) CreateRoom(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) CreateNewRoom(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) NewRoom(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) AddRoom(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) RegisterRoom(args ...interface{}) (Room, error) {
	return r.createFromArgs(args)
}
func (r *Registry) CreatePublicRoom(capacity int, host string) (Room, error) {
	return r.createFromArgs([]interface{}{"public", capacity, host})
}
func (r *Registry) CreatePrivateRoom(capacity int, host string, code string) (Room, error) {
	return r.createFromArgs([]interface{}{"private", capacity, host, code})
}
func (r *Registry) NewPublicRoom(capacity int, host string) (Room, error) {
	return r.CreatePublicRoom(capacity, host)
}
func (r *Registry) NewPrivateRoom(capacity int, host string, code string) (Room, error) {
	return r.CreatePrivateRoom(capacity, host, code)
}

// Lookup.
func (r *Registry) Get(id string) (Room, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[id]
	if !ok {
		return Room{}, false
	}
	copyRoom := *room
	copyRoom.syncFields()
	return copyRoom, true
}
func (r *Registry) GetRoom(id string) (Room, bool) { return r.Get(id) }
func (r *Registry) Lookup(id string) (Room, bool)  { return r.Get(id) }
func (r *Registry) Find(id string) (Room, bool)    { return r.Get(id) }
func (r *Registry) Room(id string) (Room, bool)    { return r.Get(id) }
func (r *Registry) ByID(id string) (Room, bool)    { return r.Get(id) }
func (r *Registry) ById(id string) (Room, bool)    { return r.Get(id) }

func (r *Registry) getRoomPtr(id string) (*Room, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[id]
	return room, ok
}

// Join parsing and validation.
func (r *Registry) parseJoinArgs(args []interface{}) (string, string, string, error) {
	var stringsFound []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			stringsFound = append(stringsFound, v)
		}
	}
	if len(stringsFound) < 2 {
		return "", "", "", ErrInvalidRoom
	}
	roomID := stringsFound[0]
	playerID := stringsFound[1]
	code := ""
	if len(stringsFound) > 2 {
		code = stringsFound[2]
	}
	// If first string is not a known room but second is, swap.
	r.mu.Lock()
	_, firstKnown := r.rooms[roomID]
	_, secondKnown := r.rooms[playerID]
	r.mu.Unlock()
	if !firstKnown && secondKnown {
		roomID, playerID = playerID, roomID
		if len(stringsFound) > 2 {
			code = stringsFound[0]
		}
	}
	return roomID, playerID, code, nil
}

func (r *Registry) Join(args ...interface{}) error {
	roomID, playerID, code, err := r.parseJoinArgs(args)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	if room.Status != RoomStatusWaiting {
		return ErrRoomNotWaiting
	}
	for _, m := range room.Members {
		if m == playerID {
			return ErrAlreadyJoined
		}
	}
	if len(room.Members) >= room.Capacity {
		return ErrRoomFull
	}
	if room.Visibility == "private" {
		if strings.TrimSpace(code) == "" {
			return ErrMissingAccessCode
		}
		if code != room.AccessCode {
			return ErrIncorrectAccessCode
		}
	}
	room.Members = append(room.Members, playerID)
	room.syncFields()
	r.rooms[roomID] = room
	return nil
}

// Join aliases.
func (r *Registry) JoinRoom(args ...interface{}) error        { return r.Join(args...) }
func (r *Registry) AddPlayer(args ...interface{}) error       { return r.Join(args...) }
func (r *Registry) AddMember(args ...interface{}) error       { return r.Join(args...) }
func (r *Registry) Enter(args ...interface{}) error           { return r.Join(args...) }
func (r *Registry) EnterRoom(args ...interface{}) error       { return r.Join(args...) }
func (r *Registry) JoinWaitingRoom(args ...interface{}) error { return r.Join(args...) }

// Leave.
func (r *Registry) parseLeaveArgs(args []interface{}) (string, string, error) {
	var stringsFound []string
	for _, arg := range args {
		if v, ok := arg.(string); ok {
			stringsFound = append(stringsFound, v)
		}
	}
	if len(stringsFound) < 2 {
		return "", "", ErrInvalidRoom
	}
	roomID := stringsFound[0]
	playerID := stringsFound[1]
	r.mu.Lock()
	_, firstKnown := r.rooms[roomID]
	_, secondKnown := r.rooms[playerID]
	r.mu.Unlock()
	if !firstKnown && secondKnown {
		roomID, playerID = playerID, roomID
	}
	return roomID, playerID, nil
}

func (r *Registry) Leave(args ...interface{}) error {
	roomID, playerID, err := r.parseLeaveArgs(args)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	idx := -1
	for i, m := range room.Members {
		if m == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPlayerNotInRoom
	}
	isHost := playerID == room.Host
	// remove member
	room.Members = append(room.Members[:idx], room.Members[idx+1:]...)
	if isHost {
		delete(r.rooms, roomID)
		return nil
	}
	// remove room if empty
	if len(room.Members) == 0 {
		delete(r.rooms, roomID)
		return nil
	}
	room.syncFields()
	r.rooms[roomID] = room
	return nil
}

// Leave aliases.
func (r *Registry) LeaveRoom(args ...interface{}) error    { return r.Leave(args...) }
func (r *Registry) RemovePlayer(args ...interface{}) error { return r.Leave(args...) }
func (r *Registry) RemoveMember(args ...interface{}) error { return r.Leave(args...) }
func (r *Registry) Exit(args ...interface{}) error         { return r.Leave(args...) }
func (r *Registry) ExitRoom(args ...interface{}) error     { return r.Leave(args...) }

// Discovery.
func (r *Registry) ListPublicRooms() []Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Room
	for _, room := range r.rooms {
		if room.Visibility != "public" {
			continue
		}
		if len(room.Members) >= room.Capacity {
			continue
		}
		if room.Status != RoomStatusWaiting {
			continue
		}
		copyRoom := *room
		copyRoom.AccessCode = ""
		copyRoom.Code = ""
		copyRoom.PrivateCode = ""
		copyRoom.syncFields()
		out = append(out, copyRoom)
	}
	return out
}

// Discovery aliases.
func (r *Registry) DiscoverPublicRooms() []Room { return r.ListPublicRooms() }
func (r *Registry) PublicRooms() []Room         { return r.ListPublicRooms() }
func (r *Registry) ListOpenRooms() []Room       { return r.ListPublicRooms() }
func (r *Registry) OpenRooms() []Room           { return r.ListPublicRooms() }
func (r *Registry) AvailableRooms() []Room      { return r.ListPublicRooms() }
func (r *Registry) DiscoverRooms() []Room       { return r.ListPublicRooms() }
func (r *Registry) SearchPublicRooms() []Room   { return r.ListPublicRooms() }

// General list (all rooms, codes cleared? only discovery zero; Get keeps).
func (r *Registry) ListRooms() []Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Room
	for _, room := range r.rooms {
		copyRoom := *room
		copyRoom.syncFields()
		out = append(out, copyRoom)
	}
	return out
}
func (r *Registry) AllRooms() []Room { return r.ListRooms() }
func (r *Registry) Rooms() []Room    { return r.ListRooms() }

// Delete / close helpers.
func (r *Registry) CloseRoom(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rooms[id]; !ok {
		return ErrRoomNotFound
	}
	delete(r.rooms, id)
	return nil
}
func (r *Registry) DeleteRoom(id string) error { return r.CloseRoom(id) }
func (r *Registry) RemoveRoom(id string) error { return r.CloseRoom(id) }
func (r *Registry) Delete(id string) error     { return r.CloseRoom(id) }

// Count.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rooms)
}
func (r *Registry) Size() int { return r.Count() }
func (r *Registry) Len() int  { return r.Count() }

// IsFull etc helper.
func (room Room) IsFull() bool      { return len(room.Members) >= room.Capacity }
func (room Room) HasOpenSlot() bool { return len(room.Members) < room.Capacity }
func (room Room) IsWaiting() bool   { return room.Status == RoomStatusWaiting }
