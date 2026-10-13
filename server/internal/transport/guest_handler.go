package transport

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"remi-jadul/server/internal/auth"
)

// GuestHandlerConfig configures REST guest creation.
type GuestHandlerConfig struct {
	Store    *auth.GuestStore
	Secret   string
	Issuer   string
	TokenTTL time.Duration
}

// GuestHandler serves POST /api/v1/guests.
type GuestHandler struct {
	cfg   GuestHandlerConfig
	idemp *idempotencyStore
}

type idempotencyEntry struct {
	status int
	body   []byte
}

type idempotencyStore struct {
	mu       sync.Mutex
	data     map[string]idempotencyEntry
	bodyHash map[string][]byte
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{data: make(map[string]idempotencyEntry), bodyHash: make(map[string][]byte)}
}

func NewGuestHandler(args ...interface{}) *GuestHandler {
	cfg := GuestHandlerConfig{Issuer: "remi-jadul", TokenTTL: 24 * time.Hour}
	var store *auth.GuestStore
	for _, arg := range args {
		switch v := arg.(type) {
		case GuestHandlerConfig:
			cfg = mergeGuestConfig(cfg, v)
		case *GuestHandlerConfig:
			if v != nil {
				cfg = mergeGuestConfig(cfg, *v)
			}
		case *auth.GuestStore:
			store = v
		case string:
			if store == nil && !strings.HasPrefix(v, "remi") {
				// heuristic: secret if not obviously issuer
			}
			if cfg.Secret == "" {
				cfg.Secret = v
			} else if cfg.Issuer == "remi-jadul" {
				cfg.Issuer = v
			}
		case time.Duration:
			cfg.TokenTTL = v
		case int:
			cfg.TokenTTL = time.Duration(v) * time.Second
		}
	}
	if store != nil {
		cfg.Store = store
	}
	if cfg.Store == nil {
		cfg.Store = auth.NewGuestStore()
	}
	if cfg.Secret == "" {
		cfg.Secret = os.Getenv("JWT_SECRET")
	}
	if cfg.Secret == "" {
		cfg.Secret = "dev-secret-change-me"
	}
	return &GuestHandler{cfg: cfg, idemp: newIdempotencyStore()}
}

func mergeGuestConfig(base, v GuestHandlerConfig) GuestHandlerConfig {
	if v.Store != nil {
		base.Store = v.Store
	}
	if v.Secret != "" {
		base.Secret = v.Secret
	}
	if v.Issuer != "" {
		base.Issuer = v.Issuer
	}
	if v.TokenTTL > 0 {
		base.TokenTTL = v.TokenTTL
	}
	return base
}

// Alias constructors.
func NewGuestHTTPHandler(args ...interface{}) *GuestHandler { return NewGuestHandler(args...) }
func NewRESTHandler(args ...interface{}) *GuestHandler      { return NewGuestHandler(args...) }
func NewGuestRESTHandler(args ...interface{}) *GuestHandler { return NewGuestHandler(args...) }

const guestEndpointPath = "/api/v1/guests"

type guestRequestBody struct {
	DisplayName string `json:"display_name"`
}

type guestSuccessResponse struct {
	Version     string `json:"version"`
	RequestID   string `json:"request_id"`
	GuestID     string `json:"guest_id"`
	DisplayName string `json:"display_name"`
	AccessToken string `json:"access_token"`
}

type restErrorResponse struct {
	Version   string `json:"version"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (h *GuestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != guestEndpointPath || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	h.handleCreate(w, r)
}

func (h *GuestHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = generateRequestID()
	}
	idempKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempKey == "" {
		h.writeError(w, requestID, 400, "invalid_request", "Idempotency key is required.")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, requestID, 400, "invalid_request", "Request body is invalid.")
		return
	}
	var req guestRequestBody
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		h.writeError(w, requestID, 400, "invalid_request", "Request body is invalid.")
		return
	}
	trimmed := strings.TrimSpace(req.DisplayName)
	if trimmed == "" {
		h.writeError(w, requestID, 400, "invalid_request", "Display name is required.")
		return
	}

	scope := "POST /api/v1/guests:" + idempKey

	// body comparison + existing entry under lock
	h.idemp.mu.Lock()
	if existingBody, ok := h.idemp.bodyHash[scope]; ok {
		if string(existingBody) != string(bodyBytes) {
			h.idemp.mu.Unlock()
			h.writeError(w, requestID, 409, "invalid_request", "Request body does not match the original idempotency key.")
			return
		}
		if entry, ok2 := h.idemp.data[scope]; ok2 {
			h.idemp.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(entry.status)
			_, _ = w.Write(entry.body)
			return
		}
	}
	h.idemp.bodyHash[scope] = append([]byte{}, bodyBytes...)
	h.idemp.mu.Unlock()

	guest, err := h.cfg.Store.CreateGuest(trimmed)
	if err != nil {
		h.writeError(w, requestID, 400, "invalid_request", "Display name is required.")
		return
	}
	token, err := issueJWT(h.cfg.Secret, h.cfg.Issuer, guest.ID, h.cfg.TokenTTL)
	if err != nil {
		h.writeError(w, requestID, 500, "invalid_request", "Token issuance failed.")
		return
	}
	resp := guestSuccessResponse{
		Version:     "v1",
		RequestID:   requestID,
		GuestID:     guest.ID,
		DisplayName: guest.DisplayName,
		AccessToken: token,
	}
	raw, _ := json.Marshal(resp)

	h.idemp.mu.Lock()
	h.idemp.data[scope] = idempotencyEntry{status: 201, body: raw}
	h.idemp.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	_, _ = w.Write(raw)
}

// Need bodyHash field; redefine idempotencyStore with it.
// We'll patch via init? Let's adjust struct definition instead.

func issueJWT(secret, issuer, subject string, ttl time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"iss": issuer,
		"sub": subject,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(ttl).Unix(),
	}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	encHeader := base64.RawURLEncoding.EncodeToString(hb)
	encPayload := base64.RawURLEncoding.EncodeToString(pb)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encHeader + "." + encPayload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encHeader + "." + encPayload + "." + sig, nil
}

func generateRequestID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("req_%x", buf)
}

func (h *GuestHandler) writeError(w http.ResponseWriter, requestID string, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := restErrorResponse{Version: "v1", Code: code, Message: message, RequestID: requestID}
	raw, _ := json.Marshal(resp)
	_, _ = w.Write(raw)
}
