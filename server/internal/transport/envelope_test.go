package transport

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeValidation(t *testing.T) {
	env := CommandEnvelope{Version: "v1", Type: "command", Name: "game.draw", RequestID: "r1", RoomID: "room"}
	if ae := ValidateCommand(env); ae.Code != "" {
		t.Fatalf("valid err=%v", ae)
	}
	// unknown field ignored
	data := []byte(`{"version":"v1","type":"command","name":"game.draw","request_id":"r2","room_id":"room","foo":"bar"}`)
	dec, ae := DecodeCommand(data)
	if ae.Code != "" {
		t.Fatalf("decode err=%v", ae)
	}
	if dec.RequestID != "r2" {
		t.Fatal("request id")
	}
	_ = json.Unmarshal
	// missing request id
	env2 := env
	env2.RequestID = ""
	if ae := ValidateCommand(env2); ae.Code != "missing_request_id" {
		t.Fatalf("code=%v", ae.Code)
	}
	// missing room id for scoped
	env3 := CommandEnvelope{Version: "v1", Type: "command", Name: "game.draw", RequestID: "r3"}
	if ae := ValidateCommand(env3); ae.Code != "missing_room_id" {
		t.Fatalf("code=%v", ae.Code)
	}
	// unknown command
	env4 := CommandEnvelope{Version: "v1", Type: "command", Name: "foo.bar", RequestID: "r4", RoomID: "room"}
	if ae := ValidateCommand(env4); ae.Code != "invalid_command" {
		t.Fatalf("code=%v", ae.Code)
	}
	// safe error
	if !NewActionError("invalid_command").Safe() {
		t.Fatal("unexpected unsafe")
	}
}
