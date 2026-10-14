package transport

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/game"
	"remi-jadul/server/internal/realtime"
	"remi-jadul/server/internal/room"

	"github.com/coder/websocket"
)

// WebSocketGatewayConfig configures the WebSocket gateway.
type WebSocketGatewayConfig struct {
	Store       *auth.GuestStore
	Rooms       *room.Registry
	Connections *ConnectionRegistry
	Secret      string
	Issuer      string
}

// WebSocketGateway is a minimal authenticated WebSocket gateway.
type WebSocketGateway struct {
	cfg WebSocketGatewayConfig
}

func NewWebSocketGateway(args ...interface{}) *WebSocketGateway {
	cfg := WebSocketGatewayConfig{Issuer: "remi-jadul"}
	var store *auth.GuestStore
	var rooms *room.Registry
	var conns *ConnectionRegistry
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
		case *ConnectionRegistry:
			conns = v
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
	if conns != nil {
		cfg.Connections = conns
	}
	if cfg.Rooms == nil {
		cfg.Rooms = room.NewRegistry()
	}
	if cfg.Connections == nil {
		cfg.Connections = NewConnectionRegistry()
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
	connToken := "conn_" + generateRequestID()
	g.cfg.Connections.Register(&Connection{
		ID:      connToken,
		Token:   connToken,
		GuestID: guest.ID,
		Stream:  stream,
		Conn:    conn,
	})
	defer g.cfg.Connections.UnregisterToken(connToken)
	guestID := guest.ID
	_ = stream

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
		case "game.start":
			g.handleGameStart(conn, ctx, guestID, env)
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
	g.broadcastRoomUpdated(roomVal, map[string]bool{guestID: canSeeCode})
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
	codeVisible := map[string]bool{guestID: joinedWithCode, rm.Host: true}
	g.broadcastRoomUpdated(rm, codeVisible)
}

func (g *WebSocketGateway) handleGameStart(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
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
	if rm.Host != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_host"))
		return
	}
	// server-side deck
	deck := game.StandardDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	if err := g.cfg.Rooms.StartGame(roomID, guestID, deck); err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, mapRoomError(err))
		return
	}
	rm, _ = g.cfg.Rooms.Get(roomID)

	members := append([]string{}, rm.Members...)
	for _, member := range members {
		c, ok := g.cfg.Connections.GetByGuest(member)
		if !ok || c == nil {
			continue
		}
		ws, ok := c.Conn.(*websocket.Conn)
		if !ok || ws == nil {
			continue
		}
		proj, err := realtime.Project(&rm, rm.Round, member)
		if err != nil {
			continue
		}
		resp := map[string]interface{}{
			"version":    "v1",
			"type":       "event",
			"name":       "game.started",
			"request_id": env.RequestID,
			"room_id":    rm.ID,
			"payload": map[string]interface{}{
				"state": proj,
			},
		}
		raw, _ := json.Marshal(resp)
		_ = ws.Write(context.Background(), websocket.MessageText, raw)
	}
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
	priorMembers := append([]string{}, rm.Members...)
	if err := g.cfg.Rooms.Leave(roomID, guestID); err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, mapRoomError(err))
		return
	}
	rm.Status = "closed"
	rm.Members = []string{}
	for _, m := range priorMembers {
		if m == guestID {
			continue
		}
		rm.Members = append(rm.Members, m)
	}
	g.writeRoomUpdated(conn, ctx, env.RequestID, rm, guestID, false, nil)
	g.broadcastRoomUpdated(rm, map[string]bool{})
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
	case room.ErrRoomAlreadyStarted:
		return NewActionError("game_not_started")
	case room.ErrInsufficientPlayers:
		return NewActionError("invalid_request")
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

func (g *WebSocketGateway) broadcastRoomUpdated(rm room.Room, codeVisible map[string]bool) {
	for _, member := range rm.Members {
		conn, ok := g.cfg.Connections.GetByGuest(member)
		if !ok || conn == nil {
			continue
		}
		ws, ok := conn.Conn.(*websocket.Conn)
		if !ok || ws == nil {
			continue
		}
		seeCode := codeVisible[member] || member == rm.Host
		msg := g.roomUpdatedMessage(rm, member, seeCode)
		_ = ws.Write(context.Background(), websocket.MessageText, msg)
	}
}

func (g *WebSocketGateway) roomUpdatedMessage(rm room.Room, requesterID string, canSeeCode bool) []byte {
	rmSync := rm
	_ = rmSync
	players := make([]map[string]interface{}, 0, len(rm.Members))
	for _, member := range rm.Members {
		isHost := member == rm.Host
		displayName := member
		if mem, err := g.cfg.Store.GetGuest(member); err == nil && mem != nil && mem.DisplayName != "" {
			displayName = mem.DisplayName
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
	if rm.Visibility == "private" {
		if canSeeCode {
			roomObj["private_code"] = rm.AccessCode
		} else {
			roomObj["private_code"] = nil
		}
	} else {
		roomObj["private_code"] = nil
	}
	resp := map[string]interface{}{
		"version": "v1",
		"type":    "event",
		"name":    "room.updated",
		"room_id": rm.ID,
		"payload": map[string]interface{}{
			"room": roomObj,
		},
	}
	raw, _ := json.Marshal(resp)
	return raw
}

func (g *WebSocketGateway) Handler() http.Handler { return g }
