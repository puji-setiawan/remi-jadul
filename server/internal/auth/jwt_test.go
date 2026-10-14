package auth

import (
	"testing"
	"time"
)

func TestValidateJWT(t *testing.T) {
	store := NewGuestStore()
	g, _ := store.CreateGuest("Raka")
	secret := "secret"
	token, _ := IssueJWT(secret, "iss", g.ID, time.Hour)
	got, err := ValidateJWT(token, secret, "iss", store)
	if err != nil || got.ID != g.ID {
		t.Fatalf("valid err=%v", err)
	}

	// expired
	expToken, _ := IssueJWT(secret, "iss", g.ID, -time.Hour)
	if _, err := ValidateJWT(expToken, secret, "iss", store); err != ErrExpiredToken {
		t.Fatalf("expired err=%v", err)
	}

	// wrong secret
	if _, err := ValidateJWT(token, "other", "iss", store); err != ErrInvalidSignature {
		t.Fatalf("sig err=%v", err)
	}

	// wrong issuer
	if _, err := ValidateJWT(token, secret, "other", store); err != ErrWrongIssuer {
		t.Fatalf("iss err=%v", err)
	}

	// malformed
	if _, err := ValidateJWT("bad.token", secret, "iss", store); err == nil {
		t.Fatal("malformed accepted")
	}

	// empty
	if _, err := ValidateJWT("", secret, "iss", store); err == nil {
		t.Fatal("empty accepted")
	}

	// unknown guest: new store
	store2 := NewGuestStore()
	if _, err := ValidateJWT(token, secret, "iss", store2); err != ErrGuestNotFound {
		t.Fatalf("unknown err=%v", err)
	}

	// no mutation
	if store.Count() != 1 {
		t.Fatal("mutated")
	}

	// config struct
	cfg := JWTConfig{Secret: secret, Issuer: "iss", Store: store}
	if _, err := ValidateJWT(token, cfg); err != nil {
		t.Fatalf("config err=%v", err)
	}
}
