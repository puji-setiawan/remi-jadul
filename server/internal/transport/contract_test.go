package transport

import (
	"encoding/json"
	"testing"
)

func TestContractActionErrorShape(t *testing.T) {
	ae := NewActionError("invalid_command")
	raw, _ := json.Marshal(ae)
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	if len(m) != 3 {
		t.Fatalf("keys=%d", len(m))
	}
	for _, k := range []string{"code", "message", "retryable"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
	if !ae.Safe() {
		t.Fatal("unsafe error")
	}
}
