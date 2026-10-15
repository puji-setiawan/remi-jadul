package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/persistence"
	"remi-jadul/server/internal/room"
)

func TestRoomDiscoveryHandlerHappyAndFiltering(t *testing.T) {
	reg := room.NewRegistry()

	// Waiting public room with open seats (1/4) -> should be included.
	pubOpen, err := reg.CreatePublicRoom(4, "guest_1")
	if err != nil {
		t.Fatalf("create public open room: %v", err)
	}
	if err := reg.Join(pubOpen.ID, "guest_2"); err != nil {
		t.Fatalf("join public open room: %v", err)
	}

	// Private room with open seats -> must be excluded.
	privRoom, err := reg.CreatePrivateRoom(4, "guest_3", "SECRET99")
	if err != nil {
		t.Fatalf("create private room: %v", err)
	}

	// Full public room (2/2) -> must be excluded.
	pubFull, err := reg.CreatePublicRoom(2, "guest_4")
	if err != nil {
		t.Fatalf("create public full room: %v", err)
	}
	if err := reg.Join(pubFull.ID, "guest_5"); err != nil {
		t.Fatalf("join public full room: %v", err)
	}

	// Started/active public room with capacity 3 and 2 players -> must be excluded.
	pubActive, err := reg.CreatePublicRoom(3, "guest_6")
	if err != nil {
		t.Fatalf("create public active room: %v", err)
	}
	if err := reg.Join(pubActive.ID, "guest_7"); err != nil {
		t.Fatalf("join public active room: %v", err)
	}
	if err := reg.StartGame(pubActive.ID, "guest_6", game.NewDeck()); err != nil {
		t.Fatalf("start public room: %v", err)
	}

	h := NewRoomDiscoveryHandler(reg)

	// 1. Explicit visibility=public with X-Request-ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?visibility=public", nil)
	req.Header.Set("X-Request-ID", "req_rooms_01")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	rawBody := rr.Body.String()
	if strings.Contains(rawBody, "SECRET99") || strings.Contains(rawBody, privRoom.ID) {
		t.Fatalf("private room or code leaked in discovery response: %s", rawBody)
	}
	if strings.Contains(rawBody, "private_code") || strings.Contains(rawBody, "own_hand") {
		t.Fatalf("forbidden fields present in discovery response: %s", rawBody)
	}

	var resp roomsDiscoveryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Version != "v1" {
		t.Fatalf("version=%s", resp.Version)
	}
	if resp.RequestID != "req_rooms_01" {
		t.Fatalf("request_id=%s", resp.RequestID)
	}
	if len(resp.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d: %+v", len(resp.Rooms), resp.Rooms)
	}
	got := resp.Rooms[0]
	if got.RoomID != pubOpen.ID || got.Visibility != "public" || got.Capacity != 4 || got.OccupiedSeats != 2 || got.Status != "waiting" {
		t.Fatalf("unexpected room summary: %+v", got)
	}

	// Verify exact JSON keys on room item
	var rawMap map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &rawMap)
	roomsArr, ok := rawMap["rooms"].([]interface{})
	if !ok || len(roomsArr) != 1 {
		t.Fatalf("unexpected rooms array: %v", rawMap["rooms"])
	}
	roomMap, ok := roomsArr[0].(map[string]interface{})
	if !ok || len(roomMap) != 5 {
		t.Fatalf("expected 5 keys in room summary, got %v", roomMap)
	}

	// 2. Omitted X-Request-ID generates one, omitted visibility query defaults to public rooms
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	var resp2 roomsDiscoveryResponse
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	if !strings.HasPrefix(resp2.RequestID, "req_") {
		t.Fatalf("expected generated request_id, got %s", resp2.RequestID)
	}
	if len(resp2.Rooms) != 1 || resp2.Rooms[0].RoomID != pubOpen.ID {
		t.Fatalf("unexpected rooms without query param: %+v", resp2.Rooms)
	}
}

func TestRoomDiscoveryHandlerInvalidRequests(t *testing.T) {
	h := NewRoomDiscoveryHandler(room.NewRegistry())

	cases := []struct {
		name      string
		method    string
		target    string
		requestID string
	}{
		{name: "visibility=private", method: http.MethodGet, target: "/api/v1/rooms?visibility=private", requestID: "req_bad_vis_1"},
		{name: "visibility=all", method: http.MethodGet, target: "/api/v1/rooms?visibility=all", requestID: "req_bad_vis_2"},
		{name: "empty visibility", method: http.MethodGet, target: "/api/v1/rooms?visibility=", requestID: "req_bad_vis_3"},
		{name: "multiple visibility values", method: http.MethodGet, target: "/api/v1/rooms?visibility=public&visibility=private", requestID: ""},
		{name: "unsupported method POST", method: http.MethodPost, target: "/api/v1/rooms?visibility=public", requestID: "req_bad_method"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			if tc.requestID != "" {
				req.Header.Set("X-Request-ID", tc.requestID)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
			}
			var errResp restErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("unmarshal error response: %v", err)
			}
			if errResp.Version != "v1" {
				t.Fatalf("version=%s", errResp.Version)
			}
			if errResp.Code != "invalid_request" {
				t.Fatalf("code=%s", errResp.Code)
			}
			if errResp.Message == "" {
				t.Fatal("expected non-empty error message")
			}
			if tc.requestID != "" && errResp.RequestID != tc.requestID {
				t.Fatalf("request_id=%s want=%s", errResp.RequestID, tc.requestID)
			}
			if tc.requestID == "" && !strings.HasPrefix(errResp.RequestID, "req_") {
				t.Fatalf("expected generated request_id, got %s", errResp.RequestID)
			}
		})
	}
}

func TestMatchHistoryHandlerHappyAndNotFound(t *testing.T) {
	store := NewMatchHistoryStore()
	err := store.Save(persistence.CompletedMatch{
		MatchID:     "match_123",
		RoundCount:  3,
		CompletedAt: time.Now().UTC(),
		Players: []persistence.CompletedMatchPlayer{
			{MatchID: "match_123", PlayerID: "player_456", DisplayName: "Nina", CumulativeScore: -15, Rank: 2},
			{MatchID: "match_123", PlayerID: "player_123", DisplayName: "Raka", CumulativeScore: 60, Rank: 1},
		},
	})
	if err != nil {
		t.Fatalf("save completed match: %v", err)
	}

	h := NewMatchHistoryHandler(store)

	// 1. Happy path with X-Request-ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match_123", nil)
	req.Header.Set("X-Request-ID", "req_history_01")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp matchHistoryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal match history: %v", err)
	}
	if resp.Version != "v1" || resp.RequestID != "req_history_01" || resp.MatchID != "match_123" || resp.RoundCount != 3 {
		t.Fatalf("unexpected top-level response: %+v", resp)
	}
	if len(resp.FinalRanking) != 2 {
		t.Fatalf("expected 2 ranking entries, got %d", len(resp.FinalRanking))
	}
	// Sorted by rank ascending
	if resp.FinalRanking[0].PlayerID != "player_123" || resp.FinalRanking[0].Rank != 1 || resp.FinalRanking[0].CumulativeScore != 60 || resp.FinalRanking[0].DisplayName != "Raka" {
		t.Fatalf("unexpected rank 1 entry: %+v", resp.FinalRanking[0])
	}
	if resp.FinalRanking[1].PlayerID != "player_456" || resp.FinalRanking[1].Rank != 2 || resp.FinalRanking[1].CumulativeScore != -15 || resp.FinalRanking[1].DisplayName != "Nina" {
		t.Fatalf("unexpected rank 2 entry: %+v", resp.FinalRanking[1])
	}

	// 2. Unknown match_id returns 404 not_found with generated request_id
	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match_unknown", nil)
	rrNotFound := httptest.NewRecorder()
	h.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rrNotFound.Code, rrNotFound.Body.String())
	}
	var errResp restErrorResponse
	if err := json.Unmarshal(rrNotFound.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal 404 response: %v", err)
	}
	if errResp.Version != "v1" || errResp.Code != "not_found" || !strings.HasPrefix(errResp.RequestID, "req_") {
		t.Fatalf("unexpected 404 error payload: %+v", errResp)
	}

	// 3. Empty match_id or nested path returns 404 not_found
	for _, badPath := range []string{"/api/v1/matches/", "/api/v1/matches", "/api/v1/matches/match_123/extra"} {
		reqBad := httptest.NewRequest(http.MethodGet, badPath, nil)
		reqBad.Header.Set("X-Request-ID", "req_path_check")
		rrBad := httptest.NewRecorder()
		h.ServeHTTP(rrBad, reqBad)
		if rrBad.Code != http.StatusNotFound {
			t.Fatalf("path %s: expected 404, got %d", badPath, rrBad.Code)
		}
		var pathErr restErrorResponse
		_ = json.Unmarshal(rrBad.Body.Bytes(), &pathErr)
		if pathErr.Code != "not_found" || pathErr.RequestID != "req_path_check" {
			t.Fatalf("path %s: unexpected error payload %+v", badPath, pathErr)
		}
	}

	// 4. Unsupported method returns 400 invalid_request
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match_123", nil)
	reqPost.Header.Set("X-Request-ID", "req_post_match")
	rrPost := httptest.NewRecorder()
	h.ServeHTTP(rrPost, reqPost)
	if rrPost.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on POST, got %d", rrPost.Code)
	}
	var postErr restErrorResponse
	_ = json.Unmarshal(rrPost.Body.Bytes(), &postErr)
	if postErr.Code != "invalid_request" || postErr.RequestID != "req_post_match" {
		t.Fatalf("unexpected POST error payload: %+v", postErr)
	}
}

func TestAPIServerEndToEndIntegration(t *testing.T) {
	api := NewAPIServer(ServerConfig{
		Secret: "integration-secret",
		Issuer: "remi-jadul",
	})
	ts := httptest.NewServer(api)
	defer ts.Close()

	client := ts.Client()

	// 1. Create two guests over REST POST /api/v1/guests
	createGuest := func(name, idempKey, reqID string) guestSuccessResponse {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/guests", strings.NewReader(`{"display_name":"`+name+`"}`))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Idempotency-Key", idempKey)
		req.Header.Set("X-Request-ID", reqID)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("do guest request: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", res.StatusCode)
		}
		var out guestSuccessResponse
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Fatalf("decode guest response: %v", err)
		}
		return out
	}

	g1 := createGuest("Raka", "idemp-g1", "req-g1")
	g2 := createGuest("Nina", "idemp-g2", "req-g2")

	// 2. Connect both guests to /ws and create rooms
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer " + g1.AccessToken}},
	})
	if err != nil {
		t.Fatalf("dial ws g1: %v", err)
	}
	defer conn1.Close(websocket.StatusNormalClosure, "done")

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer " + g2.AccessToken}},
	})
	if err != nil {
		t.Fatalf("dial ws g2: %v", err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")

	// g1 creates a public room with capacity 2
	createPubCmd := `{"version":"v1","type":"command","name":"room.create","request_id":"req_c1","payload":{"visibility":"public","capacity":2}}`
	if err := conn1.Write(ctx, websocket.MessageText, []byte(createPubCmd)); err != nil {
		t.Fatalf("write room.create public: %v", err)
	}
	_, msg1, err := conn1.Read(ctx)
	if err != nil {
		t.Fatalf("read room.updated public: %v", err)
	}
	var ev1 struct {
		RoomID string `json:"room_id"`
	}
	_ = json.Unmarshal(msg1, &ev1)
	if ev1.RoomID == "" {
		t.Fatalf("missing room_id in event: %s", string(msg1))
	}

	// g2 creates a private room with capacity 2
	createPrivCmd := `{"version":"v1","type":"command","name":"room.create","request_id":"req_c2","payload":{"visibility":"private","capacity":2,"private_code":"PRIV42"}}`
	if err := conn2.Write(ctx, websocket.MessageText, []byte(createPrivCmd)); err != nil {
		t.Fatalf("write room.create private: %v", err)
	}
	_, _, err = conn2.Read(ctx)
	if err != nil {
		t.Fatalf("read room.updated private: %v", err)
	}

	// 3. Query GET /api/v1/rooms?visibility=public -> only g1's public room appears
	resRooms, err := client.Get(ts.URL + "/api/v1/rooms?visibility=public")
	if err != nil {
		t.Fatalf("get rooms: %v", err)
	}
	defer resRooms.Body.Close()
	if resRooms.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from rooms, got %d", resRooms.StatusCode)
	}
	var disc roomsDiscoveryResponse
	if err := json.NewDecoder(resRooms.Body).Decode(&disc); err != nil {
		t.Fatalf("decode rooms discovery: %v", err)
	}
	if len(disc.Rooms) != 1 || disc.Rooms[0].RoomID != ev1.RoomID || disc.Rooms[0].OccupiedSeats != 1 {
		t.Fatalf("unexpected rooms in discovery: %+v", disc.Rooms)
	}

	// 4. Persist a completed match and fetch it via GET /api/v1/matches/{match_id}
	if err := api.MatchStore().Save(persistence.CompletedMatch{
		MatchID:    "match_e2e_01",
		RoundCount: 3,
		Players: []persistence.CompletedMatchPlayer{
			{MatchID: "match_e2e_01", PlayerID: g1.GuestID, DisplayName: "Raka", CumulativeScore: 45, Rank: 1},
			{MatchID: "match_e2e_01", PlayerID: g2.GuestID, DisplayName: "Nina", CumulativeScore: 10, Rank: 2},
		},
	}); err != nil {
		t.Fatalf("save match: %v", err)
	}

	reqMatch, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/matches/match_e2e_01", nil)
	reqMatch.Header.Set("X-Request-ID", "req_match_e2e")
	resMatch, err := client.Do(reqMatch)
	if err != nil {
		t.Fatalf("get match: %v", err)
	}
	defer resMatch.Body.Close()
	if resMatch.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from match history, got %d", resMatch.StatusCode)
	}
	var mh matchHistoryResponse
	if err := json.NewDecoder(resMatch.Body).Decode(&mh); err != nil {
		t.Fatalf("decode match history: %v", err)
	}
	if mh.RequestID != "req_match_e2e" || mh.MatchID != "match_e2e_01" || mh.RoundCount != 3 || len(mh.FinalRanking) != 2 {
		t.Fatalf("unexpected match history response: %+v", mh)
	}
}
