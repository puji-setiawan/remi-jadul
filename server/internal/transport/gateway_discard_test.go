package transport

import "remi-jadul/server/internal/room"

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remi-jadul/server/internal/auth"

	"github.com/coder/websocket"
)

func TestGatewayDiscardBasic(t *testing.T) {
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

	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.start","request_id":"s1","room_id":"`+roomID+`"}`))
	_, _, _ = conn1.Read(context.Background())
	_, _, _ = conn2.Read(context.Background())

	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d1","room_id":"`+roomID+`","payload":{"source":"closed_deck"}}`))
	_, _, _ = conn1.Read(context.Background())          // game.drawn
	_, stateData, _ := conn1.Read(context.Background()) // game.state
	_, _, _ = conn2.Read(context.Background())

	var st map[string]interface{}
	json.Unmarshal(stateData, &st)
	ownHand := st["payload"].(map[string]interface{})["state"].(map[string]interface{})["own_hand"].([]interface{})
	var cardID string
	for _, c := range ownHand {
		cm := c.(map[string]interface{})
		if b, ok := cm["is_active_joker"].(bool); ok && !b {
			cardID = cm["card_id"].(string)
			break
		}
	}
	if cardID == "" {
		t.Fatal("no non-joker card")
	}

	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.discard","request_id":"x1","room_id":"`+roomID+`","payload":{"card_id":"`+cardID+`"}}`))
	_, data, err = conn1.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]interface{}
	json.Unmarshal(data, &state)
	if state["name"] != "game.state" {
		t.Fatalf("resp=%v body=%s", state["name"], string(data))
	}

	// phase rejection before draw
	conn2.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"d2","room_id":"`+roomID+`","payload":{"source":"closed_deck"}}`))
	// out-of-turn
	_, data, _ = conn2.Read(context.Background())
	var errResp map[string]interface{}
	json.Unmarshal(data, &errResp)
	if errResp["code"] == "" {
		t.Fatalf("expected error")
	}
}
