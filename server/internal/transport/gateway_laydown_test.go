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
	"remi-jadul/server/internal/game"

	"github.com/coder/websocket"
)

func TestGatewayLaydownBasic(t *testing.T) {
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
	_, _, _ = conn1.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())
	_, _, _ = conn2.Read(context.Background())

	// find three hearts in hand? hand has random cards; instead laydown impossible reliably. Test invalid combination using arbitrary 3 cards not owned -> card_not_owned.
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.laydown","request_id":"l1","room_id":"`+roomID+`","payload":{"card_ids":["999","998","997"]}}`))
	_, data, _ = conn1.Read(context.Background())
	var er map[string]interface{}
	json.Unmarshal(data, &er)
	if er["code"] != "card_not_owned" {
		t.Fatalf("code=%v", er["code"])
	}

	// duplicate ids
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.laydown","request_id":"l2","room_id":"`+roomID+`","payload":{"card_ids":["1","1","1"]}}`))
	_, data, _ = conn1.Read(context.Background())
	json.Unmarshal(data, &er)
	if er["code"] != "invalid_request" {
		t.Fatalf("dup code=%v", er["code"])
	}

	_ = game.StandardDeck
}
