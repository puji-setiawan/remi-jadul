package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildServerRoutesMounted(t *testing.T) {
	srv := buildServer("")

	// 1. POST /api/v1/guests
	reqGuest := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"Raka"}`))
	reqGuest.Header.Set("Idempotency-Key", "cmd-api-key-1")
	rrGuest := httptest.NewRecorder()
	srv.ServeHTTP(rrGuest, reqGuest)
	if rrGuest.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/guests status=%d body=%s", rrGuest.Code, rrGuest.Body.String())
	}

	// 2. GET /api/v1/rooms?visibility=public
	reqRooms := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?visibility=public", nil)
	rrRooms := httptest.NewRecorder()
	srv.ServeHTTP(rrRooms, reqRooms)
	if rrRooms.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/rooms status=%d body=%s", rrRooms.Code, rrRooms.Body.String())
	}

	// 3. GET /api/v1/matches/unknown -> 404
	reqMatch := httptest.NewRequest(http.MethodGet, "/api/v1/matches/unknown", nil)
	rrMatch := httptest.NewRecorder()
	srv.ServeHTTP(rrMatch, reqMatch)
	if rrMatch.Code != http.StatusNotFound {
		t.Fatalf("GET /api/v1/matches/unknown status=%d body=%s", rrMatch.Code, rrMatch.Body.String())
	}

	// 4. GET /ws without Bearer token -> 401
	reqWS := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rrWS := httptest.NewRecorder()
	srv.ServeHTTP(rrWS, reqWS)
	if rrWS.Code != http.StatusUnauthorized {
		t.Fatalf("GET /ws unauthenticated status=%d", rrWS.Code)
	}
}
