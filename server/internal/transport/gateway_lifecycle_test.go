package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/realtime"
	"remi-jadul/server/internal/room"

	"github.com/coder/websocket"
)

func lifecycleRoom(t *testing.T, host, other string) (*WebSocketGateway, room.Room) {
	t.Helper()
	rooms := room.NewRegistry()
	rm, err := rooms.CreatePublicRoom(2, host)
	if err != nil {
		t.Fatal(err)
	}
	if err := rooms.Join(rm.ID, other, ""); err != nil {
		t.Fatal(err)
	}
	return NewGateway(rooms), rm
}

func TestGatewayDiscardActiveJokerEmitsLifecycle(t *testing.T) {
	g, rm := lifecycleRoom(t, "host", "other")
	joker := game.Card{ID: "joker", Rank: game.Rank5, Suit: game.SuitSpades}
	round, err := game.NewRound(2, [][]game.Card{{joker}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	round.HasDrawn = true
	if err := g.cfg.Rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	env := CommandEnvelope{RequestID: "discard-joker", RoomID: rm.ID, Payload: json.RawMessage(`{"card_id":"joker"}`)}
	g.handleGameDiscard(nil, context.Background(), "host", env)

	got, _ := g.cfg.Rooms.Get(rm.ID)
	if got.Match == nil || got.Match.RoundNumber != 2 {
		t.Fatalf("active joker did not complete round: match=%#v", got.Match)
	}
}

func TestGatewayDrawClosedDeckEmptyEmitsLifecycle(t *testing.T) {
	g, rm := lifecycleRoom(t, "host", "other")
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.cfg.Rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	env := CommandEnvelope{RequestID: "draw-empty", RoomID: rm.ID, Payload: json.RawMessage(`{"source":"closed_deck"}`)}
	g.handleGameDraw(nil, context.Background(), "host", env)

	got, _ := g.cfg.Rooms.Get(rm.ID)
	if got.Match == nil || got.Match.RoundNumber != 2 {
		t.Fatalf("empty deck did not complete round: match=%#v", got.Match)
	}
}

func TestGatewayLifecycleCompletesThreeRounds(t *testing.T) {
	g, rm := lifecycleRoom(t, "host", "other")
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.cfg.Rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		got, _ := g.cfg.Rooms.Get(rm.ID)
		got.Round.Ended = true
		got.Round.Status = "ended"
		got.Round.EndReason = "deck_empty"
		got.Round.Scores = []int{0, 0}
		g.emitRoundLifecycle(rm.ID, "round", "deck_empty")
	}

	got, _ := g.cfg.Rooms.Get(rm.ID)
	if got.Match == nil || !got.Match.Ended {
		t.Fatalf("three rounds did not end match: match=%#v", got.Match)
	}
	if len(got.Match.CumulativeScores) != 2 || got.Match.CumulativeScores[0] != 0 || got.Match.CumulativeScores[1] != 0 {
		t.Fatalf("cumulative scores=%v", got.Match.CumulativeScores)
	}
}

func TestGatewayLifecycleReturnsCorrelatedErrorOnScoreCardinality(t *testing.T) {
	g, rm := lifecycleRoom(t, "host", "other")
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	round.Ended = true
	round.Scores = []int{1}
	if err := g.cfg.Rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	err = g.emitRoundLifecycle(rm.ID, "bad-lifecycle", "close")
	if err == nil || lifecycleActionError(err).Code != "invalid_request" {
		t.Fatalf("lifecycle error=%v", err)
	}
	got, _ := g.cfg.Rooms.Get(rm.ID)
	if got.Match != nil {
		t.Fatalf("invalid lifecycle finalized match: %#v", got.Match)
	}
}

func TestGatewayLifecycleBroadcastReturnsJSONError(t *testing.T) {
	g := NewGateway(room.NewRegistry())
	err := g.broadcastLifecycleMessage(room.Room{ID: "room-json"}, map[string]interface{}{"unsupported": make(chan int)})
	if err == nil {
		t.Fatal("expected lifecycle JSON error")
	}
}

func TestGatewayBroadcastsSkipFailedRecipientAndContinue(t *testing.T) {
	g, rm := lifecycleRoom(t, "failed", "healthy")
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.cfg.Rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	rm, _ = g.cfg.Rooms.Get(rm.ID)

	type accepted struct {
		guest string
		conn  *websocket.Conn
	}
	acceptedCh := make(chan accepted, 2)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		acceptedCh <- accepted{guest: r.URL.Query().Get("guest"), conn: conn}
		<-r.Context().Done()
	}))
	defer ts.Close()
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	failedClient, _, err := websocket.Dial(context.Background(), url+"?guest=failed", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer failedClient.Close(websocket.StatusNormalClosure, "done")
	healthyClient, _, err := websocket.Dial(context.Background(), url+"?guest=healthy", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer healthyClient.Close(websocket.StatusNormalClosure, "done")

	connections := make(map[string]*websocket.Conn)
	for range 2 {
		got := <-acceptedCh
		connections[got.guest] = got.conn
	}
	g.cfg.Connections.Register(&Connection{GuestID: "failed", Token: "failed", Conn: connections["failed"], Stream: realtime.NewRecipientStream()})
	g.cfg.Connections.Register(&Connection{GuestID: "healthy", Token: "healthy", Conn: connections["healthy"], Stream: realtime.NewRecipientStream()})

	// Close the server-side failed connection so its write fails, while the
	// healthy recipient must still receive this lifecycle event and the next
	// state event in order.
	_ = connections["failed"].Close(websocket.StatusGoingAway, "disconnected")
	msg := map[string]interface{}{"version": "v1", "type": "event", "name": "round.ended", "room_id": rm.ID}
	if err := g.broadcastLifecycleMessage(rm, msg); err != nil {
		t.Fatalf("broadcast lifecycle: %v", err)
	}
	if err := g.broadcastGameState(rm.ID, "request-1"); err != nil {
		t.Fatalf("broadcast state: %v", err)
	}
	if _, ok := g.cfg.Connections.GetByGuest("failed"); ok {
		t.Fatal("failed connection remained registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for i, wantName := range []string{"round.ended", "game.state"} {
		_, data, err := healthyClient.Read(ctx)
		if err != nil {
			t.Fatalf("healthy read %d: %v", i, err)
		}
		var got map[string]interface{}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got["name"] != wantName || got["sequence"] != float64(i+1) {
			t.Fatalf("healthy event %d=%v", i, got)
		}
	}
}

func TestGatewayScopedValidationErrorIncludesRoomAndRecipientSequence(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	conn, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(ts.URL, "http"), &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	if err := conn.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"unknown.command","request_id":"bad","room_id":"room-validation"}`)); err != nil {
		t.Fatal(err)
	}
	event := readLifecycleEvent(t, conn)
	if event["code"] != "invalid_command" || event["room_id"] != "room-validation" || event["sequence"] != float64(1) {
		t.Fatalf("scoped validation error=%v", event)
	}
}

func TestBuildNextRoundReturnsDealErrorBeforeTransition(t *testing.T) {
	if _, err := buildNextRound(1); err == nil {
		t.Fatal("expected invalid player count from next-round setup")
	}
}

func TestGatewayNextRoundSetupFailureLeavesMatchUnchanged(t *testing.T) {
	rooms := room.NewRegistry()
	rm, err := rooms.CreatePublicRoom(2, "host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(rooms)
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	round.Scores = []int{0}
	if err := rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	if err := g.emitRoundLifecycle(rm.ID, "setup-failure", "close"); err == nil {
		t.Fatal("expected next-round setup error")
	}
	got, ok := rooms.Get(rm.ID)
	if !ok || got.Match != nil || got.Round != round {
		t.Fatalf("setup failure mutated room: match=%#v round=%p want=%p", got.Match, got.Round, round)
	}
}

func readLifecycleEvent(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper()
	_, data, err := conn.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatal(err)
	}
	return event
}

func TestGatewayLifecycleEventOrderAndRequestReplay(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	invalid := []byte(`{"version":"v1","type":"command","name":"unknown.command","request_id":"replay-me"}`)
	if err := conn.Write(context.Background(), websocket.MessageText, invalid); err != nil {
		t.Fatal(err)
	}
	first := readLifecycleEvent(t, conn)
	if first["type"] != "error" || first["code"] != "invalid_command" {
		t.Fatalf("first invalid command=%v", first)
	}
	if err := conn.Write(context.Background(), websocket.MessageText, invalid); err != nil {
		t.Fatal(err)
	}
	second := readLifecycleEvent(t, conn)
	if second["type"] != "error" || second["code"] != "invalid_command" || second["request_id"] != "replay-me" {
		t.Fatalf("replayed invalid command=%v", second)
	}

	// A malformed frame has no safely recoverable request ID and remains an
	// uncached protocol error; the following valid request still executes.
	if err := conn.Write(context.Background(), websocket.MessageText, []byte(`{"bad"`)); err != nil {
		t.Fatal(err)
	}
	malformed := readLifecycleEvent(t, conn)
	if malformed["code"] != "invalid_request" || malformed["request_id"] != "" {
		t.Fatalf("malformed frame=%v", malformed)
	}
	if _, ok := malformed["room_id"]; ok {
		t.Fatalf("malformed frame reused stale room scope: %v", malformed)
	}
}

func TestGatewayReplayPreservesSequencesForMultipleEvents(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	registered, ok := g.cfg.Connections.GetByGuest(guest.ID)
	if !ok {
		t.Fatal("connection was not registered")
	}
	serverConn, ok := registered.Conn.(*websocket.Conn)
	if !ok {
		t.Fatal("registered connection has unexpected type")
	}
	roomID := "room-replay"
	responses := [][]byte{
		[]byte(`{"type":"event","name":"game.state","room_id":"room-replay","sequence":1}`),
		[]byte(`{"type":"event","name":"round.ended","room_id":"room-replay","sequence":2}`),
	}
	for _, response := range responses {
		g.writeReplayResponse(serverConn, context.Background(), guest.ID, response)
	}
	for i, want := range []float64{1, 2} {
		event := readLifecycleEvent(t, conn)
		if event["room_id"] != roomID || event["sequence"] != want {
			t.Fatalf("replay event %d=%v, want sequence %v", i, event, want)
		}
	}
}

func TestGatewayDuplicateRequestOnReplacementReplaysOriginalSequence(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	header := http.Header{"Authorization": {"Bearer " + token}}
	first, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close(websocket.StatusNormalClosure, "done")
	request := []byte(`{"version":"v1","type":"command","name":"foo.bar","request_id":"duplicate","room_id":"room-replay"}`)
	if err := first.Write(context.Background(), websocket.MessageText, request); err != nil {
		t.Fatal(err)
	}
	original := readLifecycleEvent(t, first)
	if original["sequence"] != float64(1) {
		t.Fatalf("original response=%v", original)
	}

	replacement, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer replacement.Close(websocket.StatusNormalClosure, "done")
	if err := replacement.Write(context.Background(), websocket.MessageText, request); err != nil {
		t.Fatal(err)
	}
	replayed := readLifecycleEvent(t, replacement)
	if replayed["sequence"] != float64(1) {
		t.Fatalf("replayed response=%v", replayed)
	}
}

func TestGatewayRequestFingerprintMismatchDoesNotPolluteOriginalReplay(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	original := []byte(`{"version":"v1","type":"command","name":"foo.bar","request_id":"same-request","room_id":"room-replay"}`)
	if err := conn.Write(context.Background(), websocket.MessageText, original); err != nil {
		t.Fatal(err)
	}
	first := readLifecycleEvent(t, conn)
	if first["type"] != "error" || first["code"] != "invalid_command" {
		t.Fatalf("original response=%v", first)
	}

	mismatch := []byte(`{"version":"v1","type":"command","name":"foo.bar","request_id":"same-request","room_id":"room-replay","payload":{"different":true}}`)
	if err := conn.Write(context.Background(), websocket.MessageText, mismatch); err != nil {
		t.Fatal(err)
	}
	if got := readLifecycleEvent(t, conn); got["type"] != "error" || got["code"] != "invalid_request" {
		t.Fatalf("fingerprint mismatch response=%v", got)
	}

	if err := conn.Write(context.Background(), websocket.MessageText, original); err != nil {
		t.Fatal(err)
	}
	replayed := readLifecycleEvent(t, conn)
	if replayed["type"] != "error" || replayed["code"] != "invalid_command" {
		t.Fatalf("replayed original response=%v", replayed)
	}
	readCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, _, err := conn.Read(readCtx); err == nil {
		t.Fatal("original retry replayed an unexpected additional response")
	}
}

func TestGatewayReconnectBeforeResponseStillCachesForRequestOwner(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	header := http.Header{"Authorization": {"Bearer " + token}}
	first, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close(websocket.StatusNormalClosure, "done")
	registered, ok := g.cfg.Connections.GetByGuest(guest.ID)
	if !ok {
		t.Fatal("first connection was not registered")
	}
	serverConn, ok := registered.Conn.(*websocket.Conn)
	if !ok {
		t.Fatal("registered connection has unexpected type")
	}

	env := CommandEnvelope{Name: "foo.bar", RequestID: "late-response", RoomID: "room-reconnect"}
	replay, finish := g.beginRequest(serverConn, guest.ID, env)
	if replay {
		t.Fatal("request unexpectedly replayed")
	}
	replacement, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer replacement.Close(websocket.StatusNormalClosure, "done")

	raw := []byte(`{"version":"v1","type":"event","name":"late.response","request_id":"late-response","room_id":"room-reconnect"}`)
	if err := g.writeRaw(serverConn, context.Background(), raw, "room-reconnect"); err != nil {
		t.Fatal(err)
	}
	finish()
	if got := readLifecycleEvent(t, first); got["sequence"] != float64(1) {
		t.Fatalf("original response=%v", got)
	}
	if err := replacement.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"foo.bar","request_id":"late-response","room_id":"room-reconnect"}`)); err != nil {
		t.Fatal(err)
	}
	if got := readLifecycleEvent(t, replacement); got["sequence"] != float64(1) {
		t.Fatalf("replayed response=%v", got)
	}
}

func TestGatewayReconnectKeepsRoomEventSequence(t *testing.T) {
	store := auth.NewGuestStore()
	guest, err := store.CreateGuest("Host")
	if err != nil {
		t.Fatal(err)
	}
	rooms := room.NewRegistry()
	rm, err := rooms.CreatePublicRoom(2, guest.ID)
	if err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, rooms, "secret", "iss")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, err := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	if err != nil {
		t.Fatal(err)
	}
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	header := http.Header{"Authorization": {"Bearer " + token}}
	first, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close(websocket.StatusNormalClosure, "done")
	g.broadcastRoomUpdated(rm, map[string]bool{})
	if got := readLifecycleEvent(t, first); got["sequence"] != float64(1) {
		t.Fatalf("first room event=%v", got)
	}

	replacement, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer replacement.Close(websocket.StatusNormalClosure, "done")
	g.broadcastRoomUpdated(rm, map[string]bool{})
	if got := readLifecycleEvent(t, replacement); got["sequence"] != float64(2) {
		t.Fatalf("reconnected room event=%v", got)
	}
}

func TestGatewayStateResyncWaitsForRoomActionLock(t *testing.T) {
	rooms := room.NewRegistry()
	rm, err := rooms.CreatePublicRoom(2, "host")
	if err != nil {
		t.Fatal(err)
	}
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	if err := rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	g := NewGateway(rooms)

	unlock := g.lockRoom(rm.ID)
	done := make(chan struct{})
	go func() {
		g.handleStateResync(nil, context.Background(), "host", CommandEnvelope{RoomID: rm.ID})
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("resync ran while room action lock was held")
	case <-time.After(20 * time.Millisecond):
	}
	unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("resync did not run after room action lock was released")
	}
}

func TestGatewayLifecycleEmitsThirdRoundFinalRankingOverWebSocket(t *testing.T) {
	store := auth.NewGuestStore()
	guest, _ := store.CreateGuest("Host")
	rooms := room.NewRegistry()
	rm, err := rooms.CreatePublicRoom(2, guest.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := rooms.Join(rm.ID, "other", ""); err != nil {
		t.Fatal(err)
	}
	g := NewGateway(store, rooms, "secret", "remi-jadul")
	ts := httptest.NewServer(g)
	defer ts.Close()
	token, _ := auth.IssueJWT("secret", "remi-jadul", guest.ID, 3600000000000)
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	// The gateway's connection registry is populated by ServeHTTP before the
	// command loop; emit against that registered recipient.
	round, err := game.NewRound(2, [][]game.Card{{}, {}}, nil, game.Rank5)
	if err != nil {
		t.Fatal(err)
	}
	if err := rooms.SetRound(rm.ID, round); err != nil {
		t.Fatal(err)
	}
	for roundNumber := 1; roundNumber <= 3; roundNumber++ {
		got, ok := rooms.Get(rm.ID)
		if !ok || got.Round == nil {
			t.Fatalf("round %d missing: %#v", roundNumber, got)
		}
		got.Round.Ended = true
		got.Round.Status = "ended"
		got.Round.EndReason = "deck_empty"
		got.Round.Scores = []int{10, 10}
		g.emitRoundLifecycle(rm.ID, "lifecycle", "deck_empty")

		want := []string{"game.state", "round.ended", "match.updated"}
		for _, name := range want {
			event := readLifecycleEvent(t, conn)
			if event["name"] != name {
				t.Fatalf("round %d event=%v want name=%q", roundNumber, event, name)
			}
		}
		if roundNumber == 3 {
			event := readLifecycleEvent(t, conn)
			if event["name"] != "match.ended" {
				t.Fatalf("final event=%v", event)
			}
			payload, ok := event["payload"].(map[string]interface{})
			if !ok {
				t.Fatalf("final payload=%v", event["payload"])
			}
			ranking, ok := payload["final_ranking"].([]interface{})
			if !ok || len(ranking) != 2 {
				t.Fatalf("final ranking=%v", payload["final_ranking"])
			}
			for i, wantPlayer := range []string{guest.ID, "other"} {
				entry, ok := ranking[i].(map[string]interface{})
				if !ok || entry["player_id"] != wantPlayer || entry["rank"] != float64(1) {
					t.Fatalf("final ranking[%d]=%v", i, ranking[i])
				}
			}
		}
	}
}
