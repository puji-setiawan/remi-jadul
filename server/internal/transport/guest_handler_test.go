package transport

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remi-jadul/server/internal/auth"
)

func TestGuestHandlerCreateAndIdempotency(t *testing.T) {
	store := auth.NewGuestStore()
	h := NewGuestHandler(store, "secret", "remi-jadul", 3600)
	// valid
	req := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"  Raka  "}`))
	req.Header.Set("Idempotency-Key", "key1")
	req.Header.Set("X-Request-ID", "req-1")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 201 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp guestSuccessResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.DisplayName != "Raka" {
		t.Fatalf("name=%s", resp.DisplayName)
	}
	if resp.RequestID != "req-1" {
		t.Fatalf("request_id=%s", resp.RequestID)
	}
	if resp.AccessToken == "" {
		t.Fatal("token missing")
	}
	if store.Count() != 1 {
		t.Fatalf("count=%d", store.Count())
	}

	// blank nickname
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"   "}`))
	req2.Header.Set("Idempotency-Key", "key2")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != 400 {
		t.Fatalf("blank status=%d", rr2.Code)
	}
	if store.Count() != 1 {
		t.Fatal("guest created on blank")
	}

	// missing key
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"Nina"}`))
	rr3 := httptest.NewRecorder()
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 400 {
		t.Fatalf("missing key status=%d", rr3.Code)
	}

	// idempotency replay
	req4 := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"  Raka  "}`))
	req4.Header.Set("Idempotency-Key", "key1")
	req4.Header.Set("X-Request-ID", "req-999")
	rr4 := httptest.NewRecorder()
	h.ServeHTTP(rr4, req4)
	if rr4.Code != 201 {
		t.Fatalf("replay status=%d", rr4.Code)
	}
	var resp4 guestSuccessResponse
	_ = json.Unmarshal(rr4.Body.Bytes(), &resp4)
	if resp4.RequestID != "req-1" {
		t.Fatalf("replay request_id=%s", resp4.RequestID)
	}
	if store.Count() != 1 {
		t.Fatal("replay created guest")
	}

	// conflict
	req5 := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"Changed"}`))
	req5.Header.Set("Idempotency-Key", "key1")
	rr5 := httptest.NewRecorder()
	h.ServeHTTP(rr5, req5)
	if rr5.Code != 409 {
		t.Fatalf("conflict status=%d", rr5.Code)
	}
	var errResp restErrorResponse
	_ = json.Unmarshal(rr5.Body.Bytes(), &errResp)
	if errResp.Code != "invalid_request" {
		t.Fatalf("code=%s", errResp.Code)
	}

	// generated request id when absent
	req6 := httptest.NewRequest(http.MethodPost, "/api/v1/guests", strings.NewReader(`{"display_name":"Nina"}`))
	req6.Header.Set("Idempotency-Key", "key3")
	rr6 := httptest.NewRecorder()
	h.ServeHTTP(rr6, req6)
	var resp6 guestSuccessResponse
	_ = json.Unmarshal(rr6.Body.Bytes(), &resp6)
	if !strings.HasPrefix(resp6.RequestID, "req_") {
		t.Fatalf("request_id=%s", resp6.RequestID)
	}

	// JWT claims
	parts := strings.Split(resp.AccessToken, ".")
	if len(parts) != 3 {
		t.Fatal("jwt shape")
	}
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]interface{}
	_ = json.Unmarshal(payloadBytes, &claims)
	if claims["iss"] != "remi-jadul" {
		t.Fatalf("iss=%v", claims["iss"])
	}
	if _, ok := claims["exp"]; !ok {
		t.Fatal("missing exp")
	}

	_ = io.ReadAll
}
