package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/persistence"
	"remi-jadul/server/internal/room"
	"remi-jadul/server/internal/transport"
)

func buildServer(databaseURL string) *transport.APIServer {
	guestStore := auth.NewGuestStore()
	roomRegistry := room.NewRegistry()

	var guestRepo transport.GuestIdentityStore
	var matchRepo *persistence.MatchRepository

	if dsn := strings.TrimSpace(databaseURL); dsn != "" {
		pgGuestRepo, err := persistence.NewPostgresGuestRepository(dsn)
		if err != nil {
			log.Printf("warning: failed to connect to PostgreSQL (%v); falling back to in-memory stores", err)
		} else {
			guestRepo = pgGuestRepo
			matchRepo = persistence.NewMatchRepository(pgGuestRepo.DB())
		}
	}

	matchStore := transport.NewMatchHistoryStore(matchRepo)

	tokenTTL := 24 * time.Hour
	if rawTTL := strings.TrimSpace(os.Getenv("JWT_TTL")); rawTTL != "" {
		if parsed, err := time.ParseDuration(rawTTL); err == nil && parsed > 0 {
			tokenTTL = parsed
		}
	}

	return transport.NewAPIServer(transport.ServerConfig{
		Store:     guestStore,
		GuestRepo: guestRepo,
		Rooms:     roomRegistry,
		Matches:   matchStore,
		MatchRepo: matchRepo,
		Secret:    strings.TrimSpace(os.Getenv("JWT_SECRET")),
		Issuer:    strings.TrimSpace(os.Getenv("JWT_ISSUER")),
		TokenTTL:  tokenTTL,
	})
}

func resolveListenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("APP_HTTP_ADDR")); addr != "" {
		return addr
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func main() {
	srv := buildServer(os.Getenv("DATABASE_URL"))
	addr := resolveListenAddr()
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
