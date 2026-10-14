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

func TestGatewayCloseInvalid(t *testing.T) {
	store := auth.NewGuestStore()
	g1, _ := store.CreateGuest("Host")
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

	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.create","request_id":"c1","payload":{"visibility":"public","capacity":2}}`))
	_, data, _ := conn1.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())
	var created map[string]interface{}
	json.Unmarshal(data, &created)
	roomID := created["payload"].(map[string]interface{})["room"].(map[string]interface{})["room_id"].(string)

	// close before draw -> invalid_phase
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.close","request_id":"x1","room_id":"`+roomID+`","payload":{"closing_card_id":"1"}}`))
	_, data, _ = conn1.Read(context.Background())
	var er map[string]interface{}
	json.Unmarshal(data, &er)
	if er["code"] != "game_not_started" && er["code"] != "invalid_phase" {
		t.Fatalf("code=%v body=%s", er["code"], string(data))
	}

	// missing closing card
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.close","request_id":"x2","room_id":"`+roomID+`","payload":{}}`))
	_, data, _ = conn1.Read(context.Background())
	json.Unmarshal(data, &er)
	if er["code"] != "invalid_request" {
		t.Fatalf("code=%v", er["code"])
	}
}
