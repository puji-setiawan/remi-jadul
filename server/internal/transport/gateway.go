package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/realtime"

	"github.com/coder/websocket"
)

// WebSocketGatewayConfig configures the WebSocket gateway.
type WebSocketGatewayConfig struct {
	Store  *auth.GuestStore
	Secret string
	Issuer string
}

// WebSocketGateway is a minimal authenticated WebSocket gateway.
type WebSocketGateway struct {
	cfg WebSocketGatewayConfig
}

func NewWebSocketGateway(args ...interface{}) *WebSocketGateway {
	cfg := WebSocketGatewayConfig{Issuer: "remi-jadul"}
	var store *auth.GuestStore
	for _, arg := range args {
		switch v := arg.(type) {
		case WebSocketGatewayConfig:
			cfg = v
		case *WebSocketGatewayConfig:
			if v != nil {
				cfg = *v
			}
		case *auth.GuestStore:
			store = v
		case string:
			if cfg.Secret == "" {
				cfg.Secret = v
			} else if cfg.Issuer == "remi-jadul" {
				cfg.Issuer = v
			}
		}
	}
	if store != nil {
		cfg.Store = store
	}
	if cfg.Store == nil {
		cfg.Store = auth.NewGuestStore()
	}
	return &WebSocketGateway{cfg: cfg}
}

// Alias constructors.
func NewGateway(args ...interface{}) *WebSocketGateway         { return NewWebSocketGateway(args...) }
func NewWSServer(args ...interface{}) *WebSocketGateway        { return NewWebSocketGateway(args...) }
func NewWebSocketServer(args ...interface{}) *WebSocketGateway { return NewWebSocketGateway(args...) }

func (g *WebSocketGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	guest, err := auth.ValidateJWT(token, g.cfg.Secret, g.cfg.Issuer, g.cfg.Store)
	if err != nil || guest == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	stream := realtime.NewRecipientStream()
	_ = stream
	_ = guest

	ctx := r.Context()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		var env CommandEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			g.writeError(conn, ctx, "", NewActionError("invalid_request"))
			continue
		}
		if ae := ValidateCommand(env); ae.Code != "" {
			g.writeError(conn, ctx, env.RequestID, ae)
			continue
		}
	}
}

func (g *WebSocketGateway) writeError(conn *websocket.Conn, ctx context.Context, requestID string, ae ActionError) {
	wrapped := map[string]interface{}{
		"version":    "v1",
		"type":       "error",
		"name":       "protocol.error",
		"request_id": requestID,
		"code":       ae.Code,
		"message":    ae.Message,
		"retryable":  ae.Retryable,
		"payload": map[string]interface{}{
			"code":      ae.Code,
			"message":   ae.Message,
			"retryable": ae.Retryable,
		},
	}
	raw, _ := json.Marshal(wrapped)
	_ = conn.Write(ctx, websocket.MessageText, raw)
}

func (g *WebSocketGateway) Handler() http.Handler { return g }
