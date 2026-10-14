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

func TestGatewayRoomCreateJoinLeave(t *testing.T) {
	store := auth.NewGuestStore()
	guest, _ := store.CreateGuest("Raka")
	token, _ := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	reg := room.NewRegistry() // but room import needed
	_ = reg
	gw := NewGateway(store, reg, "secret", "iss")
	ts := httptest.NewServer(gw)
	defer ts.Close()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// create private room
	createMsg := `{"version":"v1","type":"command","name":"room.create","request_id":"r1","payload":{"visibility":"private","capacity":2,"access_code":"secret"}}`
	conn.Write(context.Background(), websocket.MessageText, []byte(createMsg))
	_, data, err := conn.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var created map[string]interface{}
	json.Unmarshal(data, &created)
	payload := created["payload"].(map[string]interface{})
	roomObj := payload["room"].(map[string]interface{})
	roomID := roomObj["room_id"].(string)
	if roomObj["private_code"] != "secret" {
		t.Fatalf("private code not exposed: %v", roomObj["private_code"])
	}

	// invalid capacity should not mutate? test separate key
	conn.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.create","request_id":"r2","payload":{"visibility":"public","capacity":5}}`))
	_, data2, _ := conn.Read(context.Background())
	var errResp map[string]interface{}
	json.Unmarshal(data2, &errResp)
	if errResp["code"] != "invalid_request" {
		t.Fatalf("capacity err=%v", errResp["code"])
	}

	// leave (host leaves removes)
	conn.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.leave","request_id":"r3","room_id":"`+roomID+`"}`))
	_, _, _ = conn.Read(context.Background())
	if _, ok := gw.Rooms().Get(roomID); ok {
		t.Fatal("room not removed")
	}
}
