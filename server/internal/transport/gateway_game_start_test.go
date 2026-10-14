package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/room"

	"github.com/coder/websocket"
)

func TestGatewayGameStart(t *testing.T) {
	store := auth.NewGuestStore()
	g1, _ := store.CreateGuest("Host")
	g2, _ := store.CreateGuest("Player2")
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

	// create room capacity 2
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.create","request_id":"c1","payload":{"visibility":"public","capacity":2}}`))
	_, data, _ := conn1.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())
	var created map[string]interface{}
	json.Unmarshal(data, &created)
	roomID := created["payload"].(map[string]interface{})["room"].(map[string]interface{})["room_id"].(string)

	// second player joins via second connection
	token2, _ := auth.IssueJWT("secret", "iss", g2.ID, 3600000000000)
	conn2, _, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token2}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")
	conn2.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"room.join","request_id":"j1","room_id":"`+roomID+`"}`))
	_, _, _ = conn2.Read(context.Background())
	_, _, _ = conn2.Read(context.Background())
	_, _, _ = conn1.Read(context.Background())

	// host starts
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.start","request_id":"s1","room_id":"`+roomID+`"}`))
	// conn1 receives game.started (broadcast)
	_, data1, err := conn1.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var started1 map[string]interface{}
	json.Unmarshal(data1, &started1)
	if started1["name"] != "game.started" {
		t.Fatalf("conn1 msg=%v", started1["name"])
	}
	state1 := started1["payload"].(map[string]interface{})["state"].(map[string]interface{})
	if len(state1["own_hand"].([]interface{})) != 11 {
		t.Fatalf("host hand=%v", state1["own_hand"])
	}

	// conn2 receives game.started
	_, data2, err := conn2.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var started2 map[string]interface{}
	json.Unmarshal(data2, &started2)
	if started2["name"] != "game.started" {
		t.Fatalf("conn2 msg=%v", started2["name"])
	}
	state2 := started2["payload"].(map[string]interface{})["state"].(map[string]interface{})
	if state2["own_hand"] == nil {
		t.Fatal("missing own_hand")
	}
	// host cannot start again
	conn1.Write(context.Background(), websocket.MessageText, []byte(`{"version":"v1","type":"command","name":"game.start","request_id":"s2","room_id":"`+roomID+`"}`))
	_, data3, _ := conn1.Read(context.Background())
	var started3 map[string]interface{}
	json.Unmarshal(data3, &started3)
	if started3["code"] != "game_not_started" {
		t.Fatalf("restart err=%v body=%s", started3["code"], string(data3))
	}
}
