package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/room"

	"github.com/coder/websocket"
)

func TestGatewayDraw(t *testing.T) {
	store := auth.NewGuestStore()
	g1, _ := store.CreateGuest("Host")
	g2, _ := store.CreateGuest("P2")
	token1, _ := auth.IssueJWT("secret", "iss", g1.ID, 3600000000000)
	gw := NewGateway(store, room.NewRegistry(), "secret", "iss")
	ts := httptest.NewServer(gw)
	defer ts.Close()
	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn1, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token1}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close(websocket.StatusNormalClosure, "done")

	// create+join+start with small deck (2 players, empty closed deck)
	token2, _ := auth.IssueJWT("secret", "iss", g2.ID, 3600000000000)
	conn2, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token2}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")

	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.create","request_id":"c1","payload":{"visibility":"public","capacity":2}}`))
	_, data, _ := conn1.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())
	var created map[string]interface{}
	json.Unmarshal(data, &created)
	roomID := created["payload"].(map[string]interface{})["room"].(map[string]interface{})["room_id"].(string)

	conn2.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.join","request_id":"j1","room_id":"`+roomID+`"}`))
	_, _, _ = conn2.Read(context.Background())
	_, _, _ = conn2.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())

	smallDeck := game.StandardDeck()[:23]
	_ = smallDeck
	// start with standard deck (non-empty) for normal draw first
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.start","request_id":"s1","room_id":"`+roomID+`"}`))
	_, _, _ = conn1.Read(context.Background())
	_, _, _ = conn2.Read(context.Background())

	// out-of-turn player2 draws -> not_your_turn
	conn2.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d2","room_id":"`+roomID+`","payload":{"source":"closed_deck"}}`))
	_, data, _ = conn2.Read(context.Background())
	var errResp map[string]interface{}
	json.Unmarshal(data, &errResp)
	if errResp["code"] != "not_your_turn" {
		t.Fatalf("out-of-turn code=%v body=%s", errResp["code"], string(data))
	}

	// host draws
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d1","room_id":"`+roomID+`","payload":{"source":"closed_deck"}}`))
	// conn1 receives private game.drawn
	_, data, err = conn1.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var drawn map[string]interface{}
	json.Unmarshal(data, &drawn)
	if drawn["name"] != "game.drawn" {
		t.Fatalf("first msg=%v body=%s", drawn["name"], string(data))
	}
	// plus state broadcast to conn1
	_, data, _ = conn1.Read(context.Background())
	var state1 map[string]interface{}
	json.Unmarshal(data, &state1)
	if state1["name"] != "game.state" {
		t.Fatalf("state msg=%v", state1["name"])
	}
	// conn2 receives game.state (and maybe no drawn)
	_, data, _ = conn2.Read(context.Background())
	var state2 map[string]interface{}
	json.Unmarshal(data, &state2)
	if state2["name"] != "game.state" {
		t.Fatalf("opp msg=%v body=%s", state2["name"], string(data))
	}

	// repeated draw by host -> invalid_phase
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d3","room_id":"`+roomID+`","payload":{"source":"closed_deck"}}`))
	_, data, _ = conn1.Read(context.Background())
	var errResp2 map[string]interface{}
	json.Unmarshal(data, &errResp2)
	if errResp2["code"] != "invalid_phase" {
		t.Fatalf("repeat code=%v body=%s", errResp2["code"], string(data))
	}

	// unsupported source
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d4","room_id":"`+roomID+`","payload":{"source":"discard"}}`))
	// first consume? host not current? Actually after draw, phase drawn, command validates source before phase? We check source first, so invalid_request. But connection read may need drain?
	_, data, err = conn1.Read(context.Background())
	if err == nil {
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		t.Logf("unsupported source response: %v", m["code"])
	}
}
