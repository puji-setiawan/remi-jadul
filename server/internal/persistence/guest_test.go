package persistence

import (
	"os"
	"testing"

	"remi-jadul/server/internal/auth"
)

func TestGuestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	repo, err := NewPostgresGuestRepository(dsn)
	if err != nil {
		t.Fatal(err)
	}
	guest, err := repo.CreateGuest("  Raka  ")
	if err != nil {
		t.Fatal(err)
	}
	if guest.DisplayName != "Raka" {
		t.Fatalf("name=%s", guest.DisplayName)
	}
	got, err := repo.GetGuest(guest.ID)
	if err != nil || got.DisplayName != "Raka" {
		t.Fatalf("get err=%v", err)
	}
	if _, err := repo.GetGuest("00000000-0000-0000-0000-000000000000"); err != auth.ErrGuestNotFound {
		t.Fatalf("unknown err=%v", err)
	}
	if _, err := repo.CreateGuest("   "); err == nil {
		t.Fatal("expected blank error")
	}
}
