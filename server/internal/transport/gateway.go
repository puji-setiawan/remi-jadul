package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/realtime"
	"remi-jadul/server/internal/room"

	"github.com/coder/websocket"
)

// WebSocketGatewayConfig configures the WebSocket gateway.
type WebSocketGatewayConfig struct {
	Store  *auth.GuestStore
	Rooms  *room.Registry
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
	var rooms *room.Registry
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
		case *room.Registry:
			rooms = v
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
	if rooms != nil {
		cfg.Rooms = rooms
	}
	if cfg.Rooms == nil {
		cfg.Rooms = room.NewRegistry()
	}
	return &WebSocketGateway{cfg: cfg}
}

// Alias constructors.
func NewGateway(args ...interface{}) *WebSocketGateway  { return NewWebSocketGateway(args...) }
func NewWSServer(args ...interface{}) *WebSocketGateway { return NewWebSocketGateway(args...) }
func NewWebSocketServer(args ...interface{}) *WebSocketGateway {
	return NewWebSocketGateway(args...)
}

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
	guestID := guest.ID

	ctx := r.Context()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		var env CommandEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			g.writeEnvelopeError(conn, ctx, "", NewActionError("invalid_request"))
			continue
		}
		if ae := ValidateCommand(env); ae.Code != "" {
			g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
			continue
		}
		switch env.Name {
		case "room.create":
			g.handleRoomCreate(conn, ctx, guestID, env)
		case "room.join":
			g.handleRoomJoin(conn, ctx, guestID, env)
		case "room.leave":
			g.handleRoomLeave(conn, ctx, guestID, env)
		default:
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_command"))
		}
	}
}

type roomCreatePayload struct {
	Visibility  string `json:"visibility"`
	Type        string `json:"type"`
	Capacity    int    `json:"capacity"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
	AccessCode  string `json:"access_code"`
	PrivateCode string `json:"private_code"`
	Code        string `json:"code"`
}

func (g *WebSocketGateway) handleRoomCreate(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	var payload roomCreatePayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	visibility := payload.Visibility
	if visibility == "" {
		visibility = payload.Type
	}
	if visibility == "" {
		if payload.AccessCode != "" || payload.PrivateCode != "" || payload.Code != "" {
			visibility = "private"
		} else {
			visibility = "public"
		}
	}
	vis, ok := parseVisibilityString(visibility)
	if !ok {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	capacity := payload.Capacity
	if capacity == 0 {
		capacity = payload.PlayerCount
	}
	if capacity == 0 {
		capacity = payload.MaxPlayers
	}
	code := payload.AccessCode
	if code == "" {
		code = payload.PrivateCode
	}
	if code == "" {
		code = payload.Code
	}
	var (
		roomVal room.Room
		err     error
	)
	if vis == "public" {
		roomVal, err = g.cfg.Rooms.CreatePublicRoom(capacity, guestID)
	} else {
		roomVal, err = g.cfg.Rooms.CreatePrivateRoom(capacity, guestID, code)
	}
	if err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, mapRoomError(err))
		return
	}
	canSeeCode := vis == "private"
	g.writeRoomUpdated(conn, ctx, env.RequestID, roomVal, guestID, canSeeCode, []string{code})
}

type roomJoinPayload struct {
	RoomID      string `json:"room_id"`
	PrivateCode string `json:"private_code"`
	AccessCode  string `json:"access_code"`
	Code        string `json:"code"`
}

func (g *WebSocketGateway) handleRoomJoin(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	var payload roomJoinPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	roomID := env.RoomID
	if roomID == "" {
		roomID = payload.RoomID
	}
	code := payload.PrivateCode
	if code == "" {
		code = payload.AccessCode
	}
	if code == "" {
		code = payload.Code
	}

	var rm room.Room
	var found bool
	if roomID != "" {
		rm, found = g.cfg.Rooms.Get(roomID)
	} else if code != "" {
		rm, found = g.cfg.Rooms.FindByAccessCode(code)
		roomID = rm.ID
	}
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	joinedWithCode := false
	if rm.Visibility == "private" {
		if code == "" {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
		if code != rm.AccessCode {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("room_private"))
			return
		}
		joinedWithCode = true
	}
	if err := g.cfg.Rooms.Join(roomID, guestID, code); err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, mapRoomError(err))
		return
	}
	rm, _ = g.cfg.Rooms.Get(roomID)
	g.writeRoomUpdated(conn, ctx, env.RequestID, rm, guestID, joinedWithCode, []string{code})
}

func (g *WebSocketGateway) handleRoomLeave(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	if roomID == "" {
		var payload struct {
			RoomID string `json:"room_id"`
		}
		if len(env.Payload) > 0 {
			_ = json.Unmarshal(env.Payload, &payload)
		}
		roomID = payload.RoomID
	}
	if roomID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if err := g.cfg.Rooms.Leave(roomID, guestID); err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, mapRoomError(err))
		return
	}
	rm.Status = "closed"
	g.writeRoomUpdated(conn, ctx, env.RequestID, rm, guestID, false, nil)
}

func (g *WebSocketGateway) writeRoomUpdated(conn *websocket.Conn, ctx context.Context, requestID string, rm room.Room, requesterID string, canSeeCode bool, suppliedCodes []string) {
	players := make([]map[string]interface{}, 0, len(rm.Members))
	for _, member := range rm.Members {
		isHost := member == rm.Host
		displayName := member
		if mem, err := g.cfg.Store.GetGuest(member); err == nil && mem != nil {
			if mem.DisplayName != "" {
				displayName = mem.DisplayName
			}
		}
		players = append(players, map[string]interface{}{
			"player_id":    member,
			"display_name": displayName,
			"connection":   "online",
			"is_host":      isHost,
		})
	}
	roomObj := map[string]interface{}{
		"room_id":        rm.ID,
		"visibility":     rm.Visibility,
		"host_player_id": rm.Host,
		"capacity":       rm.Capacity,
		"occupied_seats": len(rm.Members),
		"status":         rm.Status,
		"players":        players,
	}
	if rm.Status == "closed" {
		roomObj["status"] = "closed"
	}
	codeToExpose := ""
	if canSeeCode && rm.Visibility == "private" {
		codeToExpose = rm.AccessCode
	} else {
		for _, c := range suppliedCodes {
			if c != "" && c == rm.AccessCode {
				// joiner with correct code may see code
				codeToExpose = rm.AccessCode
			}
		}
	}
	if rm.Visibility == "private" && codeToExpose != "" {
		roomObj["private_code"] = codeToExpose
	} else {
		roomObj["private_code"] = nil
	}
	if rm.Visibility == "public" {
		roomObj["private_code"] = nil
	}
	resp := map[string]interface{}{
		"version":    "v1",
		"type":       "event",
		"name":       "room.updated",
		"request_id": requestID,
		"room_id":    rm.ID,
		"payload": map[string]interface{}{
			"room": roomObj,
		},
	}
	raw, _ := json.Marshal(resp)
	_ = conn.Write(ctx, websocket.MessageText, raw)
}

func parseVisibilityString(v string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "public":
		return "public", true
	case "private":
		return "private", true
	}
	return "", false
}

func mapRoomError(err error) ActionError {
	switch err {
	case room.ErrRoomNotFound:
		return NewActionError("not_found")
	case room.ErrRoomFull:
		return NewActionError("room_full")
	case room.ErrRoomNotWaiting:
		return NewActionError("invalid_phase")
	case room.ErrAlreadyJoined:
		return NewActionError("invalid_request")
	case room.ErrInvalidCapacity:
		return NewActionError("invalid_request")
	case room.ErrMissingAccessCode:
		return NewActionError("invalid_request")
	case room.ErrInvalidAccessCode, room.ErrIncorrectAccessCode:
		return NewActionError("room_private")
	case room.ErrNotHost:
		return NewActionError("not_host")
	default:
		return NewActionError("invalid_request")
	}
}

func RoomsFromGateway(g *WebSocketGateway) *room.Registry { return g.cfg.Rooms }

func (g *WebSocketGateway) Rooms() *room.Registry { return g.cfg.Rooms }

func (g *WebSocketGateway) writeEnvelopeError(conn *websocket.Conn, ctx context.Context, requestID string, ae ActionError) {
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
