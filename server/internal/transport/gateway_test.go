package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remi-jadul/server/internal/auth"

	"github.com/coder/websocket"
)

func TestGatewayUpgradeAndInvalidCommand(t *testing.T) {
	store := auth.NewGuestStore()
	guest, _ := store.CreateGuest("Raka")
	token, _ := auth.IssueJWT("secret", "iss", guest.ID, 3600000000000)
	gw := NewGateway(store, "secret", "iss")

	ts := httptest.NewServer(gw)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// missing token -> 401
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Fatalf("missing token status=%d", resp.StatusCode)
	}

	// invalid token -> 401
	conn, _, err := websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer bad"}}})
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "done")
		t.Fatal("invalid token upgraded")
	}

	// valid upgrade
	conn, _, err = websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + token}}})
	if err != nil {
		t.Fatalf("valid dial err=%v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// invalid command envelope (unknown command)
	msg := []byte(`{"version":"v1","type":"command","name":"foo.bar","request_id":"r1","room_id":"room"}`)
	if err := conn.Write(context.Background(), websocket.MessageText, msg); err != nil {
		t.Fatal(err)
	}
	_, data, err := conn.Read(context.Background())
	if err != nil {
		t.Fatalf("read err=%v", err)
	}
	var got map[string]interface{}
	_ = json.Unmarshal(data, &got)
	code := ""
	if c, ok := got["code"].(string); ok {
		code = c
	}
	if payload, ok := got["payload"].(map[string]interface{}); ok {
		if c, ok := payload["code"].(string); ok {
			code = c
		}
	}
	if code != "invalid_command" {
		t.Fatalf("code=%v body=%s", code, string(data))
	}

	// invalid JSON
	if err := conn.Write(context.Background(), websocket.MessageText, []byte(`{bad`)); err != nil {
		t.Fatal(err)
	}
	_, data, err = conn.Read(context.Background())
	if err != nil {
		t.Fatalf("read err=%v", err)
	}
	var got2 map[string]interface{}
	_ = json.Unmarshal(data, &got2)
	code2 := ""
	if c, ok := got2["code"].(string); ok {
		code2 = c
	}
	if payload, ok := got2["payload"].(map[string]interface{}); ok {
		if c, ok := payload["code"].(string); ok {
			code2 = c
		}
	}
	if code2 != "invalid_request" {
		t.Fatalf("code=%v body=%s", code2, string(data))
	}

	fmt.Println("gateway ok")
	_ = token
}
