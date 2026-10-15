package persistence

import (
	"errors"
	"os"
	"testing"

	"remi-jadul/server/internal/auth"
)

func TestGuestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL guest repository integration test")
	}
	repo, err := NewPostgresGuestRepository(dsn)
	if err != nil {
		t.Fatalf("open guest repository: %v", err)
	}

	guest, err := repo.CreateGuest("  Raka  ")
	if err != nil {
		t.Fatalf("CreateGuest: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteGuest(guest.ID)
	})

	if guest.ID == "" {
		t.Fatal("expected generated guest ID")
	}
	if guest.DisplayName != "Raka" || guest.Nickname != "Raka" || guest.Name != "Raka" {
		t.Fatalf("unexpected created guest: %+v", guest)
	}

	got, err := repo.GetGuest(guest.ID)
	if err != nil {
		t.Fatalf("GetGuest(%s): %v", guest.ID, err)
	}
	if got.ID != guest.ID || got.DisplayName != "Raka" || got.Nickname != "Raka" || got.Name != "Raka" {
		t.Fatalf("unexpected loaded guest: %+v", got)
	}

	const unknownID = "00000000-0000-0000-0000-000000000000"
	if _, err := repo.GetGuest(unknownID); !errors.Is(err, auth.ErrGuestNotFound) {
		t.Fatalf("GetGuest(unknown) err = %v, want %v", err, auth.ErrGuestNotFound)
	}
	if err := repo.DeleteGuest(unknownID); !errors.Is(err, auth.ErrGuestNotFound) {
		t.Fatalf("DeleteGuest(unknown) err = %v, want %v", err, auth.ErrGuestNotFound)
	}

	for _, invalid := range []string{"", "   ", "\t\n"} {
		if _, err := repo.CreateGuest(invalid); !errors.Is(err, auth.ErrBlankNickname) {
			t.Fatalf("CreateGuest(%q) err = %v, want %v", invalid, err, auth.ErrBlankNickname)
		}
	}

	if err := repo.DeleteGuest(guest.ID); err != nil {
		t.Fatalf("DeleteGuest(%s): %v", guest.ID, err)
	}
	if _, err := repo.GetGuest(guest.ID); !errors.Is(err, auth.ErrGuestNotFound) {
		t.Fatalf("GetGuest after delete err = %v, want %v", err, auth.ErrGuestNotFound)
	}
}
