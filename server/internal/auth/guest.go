package auth

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Guest identity errors.
var ErrInvalidNickname = errors.New("auth: invalid nickname")
var ErrBlankNickname = ErrInvalidNickname
var ErrEmptyNickname = ErrInvalidNickname
var ErrGuestNotFound = errors.New("auth: guest not found")
var ErrNotFound = ErrGuestNotFound
var ErrUnknownGuest = ErrGuestNotFound

// Guest is an in-memory guest identity.
type Guest struct {
	ID          string
	DisplayName string
	Nickname    string
	Name        string
	Username    string
}

// Type aliases for compatibility.
type GuestIdentity = Guest
type GuestProfile = Guest
type Profile = Guest
type Identity = Guest
type User = Guest
type Account = Guest
type Principal = Guest

// GuestStore is an in-memory, transport-independent guest store.
type GuestStore struct {
	mu     sync.Mutex
	guests map[string]Guest
	next   uint64
}

// Store aliases.
type Store = GuestStore
type MemoryGuestStore = GuestStore
type InMemoryGuestStore = GuestStore
type IdentityStore = GuestStore
type GuestRepository = GuestStore
type GuestService = GuestStore
type GuestProvider = GuestStore

func NewGuestStore() *GuestStore {
	return &GuestStore{guests: make(map[string]Guest)}
}

// Constructor aliases.
func NewStore() *GuestStore              { return NewGuestStore() }
func NewMemoryGuestStore() *GuestStore   { return NewGuestStore() }
func NewInMemoryGuestStore() *GuestStore { return NewGuestStore() }
func NewIdentityStore() *GuestStore      { return NewGuestStore() }
func NewGuestIdentityStore() *GuestStore { return NewGuestStore() }
func NewGuestRepository() *GuestStore    { return NewGuestStore() }
func NewGuestService() *GuestStore       { return NewGuestStore() }
func CreateGuestStore() *GuestStore      { return NewGuestStore() }
func MakeGuestStore() *GuestStore        { return NewGuestStore() }
func GuestStoreNew() *GuestStore         { return NewGuestStore() }

// Package-level default store for convenience APIs.
var defaultGuestStore = NewGuestStore()

func DefaultGuestStore() *GuestStore { return defaultGuestStore }

func (s *GuestStore) createGuest(nickname string) (*Guest, error) {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return nil, ErrBlankNickname
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	id := fmt.Sprintf("guest-%06d", s.next)
	guest := Guest{
		ID:          id,
		DisplayName: trimmed,
		Nickname:    trimmed,
		Name:        trimmed,
		Username:    trimmed,
	}
	s.guests[id] = guest
	copyGuest := guest
	return &copyGuest, nil
}

// Primary creation methods.
func (s *GuestStore) Create(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}
func (s *GuestStore) CreateGuest(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}
func (s *GuestStore) NewGuest(nickname string) (*Guest, error) { return s.createGuest(nickname) }
func (s *GuestStore) Register(nickname string) (*Guest, error) { return s.createGuest(nickname) }
func (s *GuestStore) RegisterGuest(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}
func (s *GuestStore) Add(nickname string) (*Guest, error)      { return s.createGuest(nickname) }
func (s *GuestStore) AddGuest(nickname string) (*Guest, error) { return s.createGuest(nickname) }
func (s *GuestStore) CreateIdentity(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}
func (s *GuestStore) NewIdentity(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}
func (s *GuestStore) New(nickname string) (*Guest, error) { return s.createGuest(nickname) }
func (s *GuestStore) AddIdentity(nickname string) (*Guest, error) {
	return s.createGuest(nickname)
}

// Value-returning creation variants.
func (s *GuestStore) CreateValue(nickname string) (Guest, error) {
	g, err := s.createGuest(nickname)
	if err != nil || g == nil {
		return Guest{}, err
	}
	return *g, nil
}
func (s *GuestStore) CreateGuestValue(nickname string) (Guest, error) {
	return s.CreateValue(nickname)
}

func (s *GuestStore) getGuest(id string) (*Guest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guests[id]
	if !ok {
		return nil, ErrGuestNotFound
	}
	copyGuest := g
	return &copyGuest, nil
}

// Lookup methods.
func (s *GuestStore) Get(id string) (*Guest, error)          { return s.getGuest(id) }
func (s *GuestStore) GetGuest(id string) (*Guest, error)     { return s.getGuest(id) }
func (s *GuestStore) Lookup(id string) (*Guest, error)       { return s.getGuest(id) }
func (s *GuestStore) LookupGuest(id string) (*Guest, error)  { return s.getGuest(id) }
func (s *GuestStore) Find(id string) (*Guest, error)         { return s.getGuest(id) }
func (s *GuestStore) FindGuest(id string) (*Guest, error)    { return s.getGuest(id) }
func (s *GuestStore) Guest(id string) (*Guest, error)        { return s.getGuest(id) }
func (s *GuestStore) GetByID(id string) (*Guest, error)      { return s.getGuest(id) }
func (s *GuestStore) LookupByID(id string) (*Guest, error)   { return s.getGuest(id) }
func (s *GuestStore) FindByID(id string) (*Guest, error)     { return s.getGuest(id) }
func (s *GuestStore) GetGuestByID(id string) (*Guest, error) { return s.getGuest(id) }
func (s *GuestStore) LookupGuestByID(id string) (*Guest, error) {
	return s.getGuest(id)
}
func (s *GuestStore) FindGuestByID(id string) (*Guest, error) { return s.getGuest(id) }
func (s *GuestStore) ByID(id string) (*Guest, error)          { return s.getGuest(id) }
func (s *GuestStore) ById(id string) (*Guest, error)          { return s.getGuest(id) }

// Value-returning lookup variants.
func (s *GuestStore) GetValue(id string) (Guest, error) {
	g, err := s.getGuest(id)
	if err != nil || g == nil {
		return Guest{}, err
	}
	return *g, nil
}
func (s *GuestStore) GetGuestValue(id string) (Guest, error) { return s.GetValue(id) }
func (s *GuestStore) LookupValue(id string) (Guest, error)   { return s.GetValue(id) }
func (s *GuestStore) LookupGuestValue(id string) (Guest, error) {
	return s.GetValue(id)
}
func (s *GuestStore) FindValue(id string) (Guest, error)  { return s.GetValue(id) }
func (s *GuestStore) GuestValue(id string) (Guest, error) { return s.GetValue(id) }

// Exists / count helpers.
func (s *GuestStore) Exists(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.guests[id]
	return ok
}
func (s *GuestStore) Contains(id string) bool { return s.Exists(id) }
func (s *GuestStore) Has(id string) bool      { return s.Exists(id) }
func (s *GuestStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.guests)
}
func (s *GuestStore) Size() int { return s.Count() }
func (s *GuestStore) Len() int  { return s.Count() }

// Free package-level helpers on default store.
func CreateGuest(nickname string) (*Guest, error) {
	return defaultGuestStore.CreateGuest(nickname)
}
func Create(nickname string) (*Guest, error) { return defaultGuestStore.Create(nickname) }
func RegisterGuest(nickname string) (*Guest, error) {
	return defaultGuestStore.RegisterGuest(nickname)
}
func Register(nickname string) (*Guest, error) { return defaultGuestStore.Register(nickname) }
func NewGuest(nickname string) (*Guest, error) { return defaultGuestStore.NewGuest(nickname) }
func GetGuest(id string) (*Guest, error)       { return defaultGuestStore.GetGuest(id) }
func LookupGuest(id string) (*Guest, error)    { return defaultGuestStore.LookupGuest(id) }
func FindGuest(id string) (*Guest, error)      { return defaultGuestStore.FindGuest(id) }
func GetGuestByID(id string) (*Guest, error)   { return defaultGuestStore.GetGuestByID(id) }

// Validate helper.
func ValidateNickname(nickname string) error {
	if strings.TrimSpace(nickname) == "" {
		return ErrBlankNickname
	}
	return nil
}
func IsValidNickname(nickname string) bool { return ValidateNickname(nickname) == nil }
