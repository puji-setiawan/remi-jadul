package transport

import (
	"net/http"
	"os"
	"strings"
	"time"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/persistence"
	"remi-jadul/server/internal/room"
)

// ServerConfig configures the unified HTTP API and WebSocket transport server.
type ServerConfig struct {
	Store       *auth.GuestStore
	GuestRepo   GuestIdentityStore
	Rooms       *room.Registry
	Matches     *MatchHistoryStore
	MatchRepo   *persistence.MatchRepository
	Connections *ConnectionRegistry
	Secret      string
	Issuer      string
	TokenTTL    time.Duration
}

// Alias for ServerConfig.
type APIServerConfig = ServerConfig

// APIServer wires REST v1 endpoints (/api/v1/guests, /api/v1/rooms, /api/v1/matches/{match_id})
// and the WebSocket gateway (/ws) over shared state.
type APIServer struct {
	mux        *http.ServeMux
	guests     *GuestHandler
	rooms      *RoomDiscoveryHandler
	matches    *MatchHistoryHandler
	gateway    *WebSocketGateway
	store      *auth.GuestStore
	roomReg    *room.Registry
	matchStore *MatchHistoryStore
}

// Aliases for APIServer.
type Server = APIServer
type TransportServer = APIServer

func NewAPIServer(args ...interface{}) *APIServer {
	cfg := ServerConfig{
		Issuer:   "remi-jadul",
		TokenTTL: 24 * time.Hour,
	}
	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch v := arg.(type) {
		case ServerConfig:
			cfg = mergeServerConfig(cfg, v)
		case *ServerConfig:
			if v != nil {
				cfg = mergeServerConfig(cfg, *v)
			}
		case *auth.GuestStore:
			cfg.Store = v
		case *persistence.GuestRepository:
			if v != nil {
				cfg.GuestRepo = v
			}
		case GuestIdentityStore:
			cfg.GuestRepo = v
		case *room.Registry:
			cfg.Rooms = v
		case *MatchHistoryStore:
			cfg.Matches = v
		case *persistence.MatchRepository:
			cfg.MatchRepo = v
		case *ConnectionRegistry:
			cfg.Connections = v
		case string:
			if cfg.Secret == "" {
				cfg.Secret = v
			} else if cfg.Issuer == "remi-jadul" {
				cfg.Issuer = v
			}
		case time.Duration:
			if v > 0 {
				cfg.TokenTTL = v
			}
		}
	}

	if cfg.Store == nil {
		cfg.Store = auth.NewGuestStore()
	}
	if cfg.Rooms == nil {
		cfg.Rooms = room.NewRegistry()
	}
	if cfg.Connections == nil {
		cfg.Connections = NewConnectionRegistry()
	}
	if cfg.Matches == nil {
		if cfg.MatchRepo != nil {
			cfg.Matches = NewMatchHistoryStore(cfg.MatchRepo)
		} else {
			cfg.Matches = NewMatchHistoryStore()
		}
	}
	if cfg.Secret == "" {
		cfg.Secret = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	if cfg.Secret == "" {
		cfg.Secret = "dev-secret-change-me"
	}
	if envIssuer := strings.TrimSpace(os.Getenv("JWT_ISSUER")); envIssuer != "" && cfg.Issuer == "remi-jadul" {
		cfg.Issuer = envIssuer
	}
	if envTTL := strings.TrimSpace(os.Getenv("JWT_TTL")); envTTL != "" && cfg.TokenTTL == 24*time.Hour {
		if parsed, err := time.ParseDuration(envTTL); err == nil && parsed > 0 {
			cfg.TokenTTL = parsed
		}
	}

	guestsHandler := NewGuestHandler(GuestHandlerConfig{
		Store:    cfg.Store,
		Repo:     cfg.GuestRepo,
		Secret:   cfg.Secret,
		Issuer:   cfg.Issuer,
		TokenTTL: cfg.TokenTTL,
	})
	roomsHandler := NewRoomDiscoveryHandler(cfg.Rooms)
	matchesHandler := NewMatchHistoryHandler(cfg.Matches)
	gateway := NewWebSocketGateway(WebSocketGatewayConfig{
		Store:       cfg.Store,
		GuestRepo:   cfg.GuestRepo,
		Rooms:       cfg.Rooms,
		Matches:     cfg.Matches,
		Connections: cfg.Connections,
		Secret:      cfg.Secret,
		Issuer:      cfg.Issuer,
	})

	mux := http.NewServeMux()
	mux.Handle("/api/v1/guests", guestsHandler)
	mux.Handle("/api/v1/rooms", roomsHandler)
	mux.Handle("/api/v1/matches/", matchesHandler)
	mux.Handle("/api/v1/matches", matchesHandler)
	mux.Handle("/ws", gateway)

	return &APIServer{
		mux:        mux,
		guests:     guestsHandler,
		rooms:      roomsHandler,
		matches:    matchesHandler,
		gateway:    gateway,
		store:      cfg.Store,
		roomReg:    cfg.Rooms,
		matchStore: cfg.Matches,
	}
}

func NewServer(args ...interface{}) *APIServer          { return NewAPIServer(args...) }
func NewTransportServer(args ...interface{}) *APIServer { return NewAPIServer(args...) }
func NewHTTPHandler(args ...interface{}) http.Handler   { return NewAPIServer(args...) }

func mergeServerConfig(base, v ServerConfig) ServerConfig {
	if v.Store != nil {
		base.Store = v.Store
	}
	if v.GuestRepo != nil {
		base.GuestRepo = v.GuestRepo
	}
	if v.Rooms != nil {
		base.Rooms = v.Rooms
	}
	if v.Matches != nil {
		base.Matches = v.Matches
	}
	if v.MatchRepo != nil {
		base.MatchRepo = v.MatchRepo
	}
	if v.Connections != nil {
		base.Connections = v.Connections
	}
	if v.Secret != "" {
		base.Secret = v.Secret
	}
	if v.Issuer != "" {
		base.Issuer = v.Issuer
	}
	if v.TokenTTL > 0 {
		base.TokenTTL = v.TokenTTL
	}
	return base
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *APIServer) Handler() http.Handler                { return s.mux }
func (s *APIServer) GuestsHandler() *GuestHandler         { return s.guests }
func (s *APIServer) RoomsHandler() *RoomDiscoveryHandler  { return s.rooms }
func (s *APIServer) MatchesHandler() *MatchHistoryHandler { return s.matches }
func (s *APIServer) Gateway() *WebSocketGateway           { return s.gateway }
func (s *APIServer) GuestStore() *auth.GuestStore         { return s.store }
func (s *APIServer) RoomRegistry() *room.Registry         { return s.roomReg }
func (s *APIServer) MatchStore() *MatchHistoryStore       { return s.matchStore }
