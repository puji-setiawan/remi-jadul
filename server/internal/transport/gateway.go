package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
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
	GuestRepo   GuestIdentityStore
	Rooms       *room.Registry
	Matches     *MatchHistoryStore
	Connections *ConnectionRegistry
	Secret      string
	Issuer      string
}

// WebSocketGateway is a minimal authenticated WebSocket gateway.
type WebSocketGateway struct {
	cfg       WebSocketGatewayConfig
	mu        sync.Mutex
	requests  map[string]cachedRequest
	roomMu    sync.Mutex
	roomLocks map[string]*sync.Mutex
	streams   *realtime.StreamManager
}

type cachedRequest struct {
	owner       string
	connection  *websocket.Conn
	fingerprint string
	responses   [][]byte
	ready       chan struct{}
}

func NewWebSocketGateway(args ...interface{}) *WebSocketGateway {
	cfg := WebSocketGatewayConfig{Issuer: "remi-jadul"}
	var store *auth.GuestStore
	var rooms *room.Registry
	var matches *MatchHistoryStore
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
		case GuestIdentityStore:
			if v != nil {
				cfg.GuestRepo = v
			}
		case *room.Registry:
			rooms = v
		case *MatchHistoryStore:
			matches = v
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
	if matches != nil {
		cfg.Matches = matches
	}
	if conns != nil {
		cfg.Connections = conns
	}
	if cfg.Rooms == nil {
		cfg.Rooms = room.NewRegistry()
	}
	if cfg.Matches == nil {
		cfg.Matches = NewMatchHistoryStore()
	}
	if cfg.Connections == nil {
		cfg.Connections = NewConnectionRegistry()
	}
	return &WebSocketGateway{cfg: cfg, requests: make(map[string]cachedRequest), roomLocks: make(map[string]*sync.Mutex), streams: realtime.NewStreamManager()}
}

func (g *WebSocketGateway) lockRoom(id string) func() {
	g.roomMu.Lock()
	lock := g.roomLocks[id]
	if lock == nil {
		lock = &sync.Mutex{}
		g.roomLocks[id] = lock
	}
	g.roomMu.Unlock()
	lock.Lock()
	return lock.Unlock
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
	if (err != nil || guest == nil) && g.cfg.GuestRepo != nil {
		guest, err = validateTokenWithRepo(token, g.cfg.Secret, g.cfg.Issuer, g.cfg.GuestRepo)
	}
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
		if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
			c.clearRoomContext()
		}
		var env CommandEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			g.writeEnvelopeError(conn, ctx, "", NewActionError("invalid_request"))
			continue
		}
		// Attach before idempotency handling as well: request-fingerprint
		// failures are room-scoped protocol errors when room_id is present.
		if env.RoomID != "" {
			if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
				c.setRoomContext(env.RoomID, g.streams.Get(env.RoomID, guestID))
			}
		}
		claimed := false
		var finishRequest func()
		if env.RequestID != "" {
			var replay bool
			replay, finishRequest = g.beginRequest(conn, guestID, env)
			if replay {
				continue
			}
			claimed = true
		}
		if ae := ValidateCommand(env); ae.Code != "" {
			g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
			if claimed {
				finishRequest()
			}
			continue
		}
		if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
			var stream *realtime.RecipientStream
			if env.RoomID != "" {
				stream = g.streams.Get(env.RoomID, guestID)
			}
			c.setRoomContext(env.RoomID, stream)
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
		case "game.draw":
			g.handleGameDraw(conn, ctx, guestID, env)
		case "game.discard":
			g.handleGameDiscard(conn, ctx, guestID, env)
		case "game.laydown":
			g.handleGameLaydown(conn, ctx, guestID, env)
		case "game.close":
			g.handleGameClose(conn, ctx, guestID, env)
		case "game.take_discard":
			g.handleGameTakeDiscard(conn, ctx, guestID, env)
		case "state.resync":
			g.handleStateResync(conn, ctx, guestID, env)
		default:
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_command"))
		}
		if claimed {
			finishRequest()
		}
	}
}

func requestKey(guestID, requestID string) string { return guestID + "\x00" + requestID }

func requestFingerprint(env CommandEnvelope) string {
	b, _ := json.Marshal(struct {
		Name, Room string
		Payload    json.RawMessage
		Sequence   uint64
	}{env.Name, env.RoomID, env.Payload, env.Sequence})
	return string(b)
}

func (g *WebSocketGateway) beginRequest(conn *websocket.Conn, guestID string, env CommandEnvelope) (bool, func()) {
	fingerprint := requestFingerprint(env)
	g.mu.Lock()
	k := requestKey(guestID, env.RequestID)
	if req, ok := g.requests[k]; ok {
		ready := req.ready
		g.mu.Unlock()
		if req.fingerprint != fingerprint {
			g.writeEnvelopeErrorUncached(conn, context.Background(), env.RequestID, NewActionError("invalid_request"))
			return true, nil
		}
		<-ready
		g.mu.Lock()
		req = g.requests[k]
		g.mu.Unlock()
		for _, response := range req.responses {
			g.writeReplayResponse(conn, context.Background(), guestID, response)
		}
		return true, nil
	}
	g.requests[k] = cachedRequest{owner: guestID, connection: conn, fingerprint: fingerprint, ready: make(chan struct{})}
	g.mu.Unlock()
	return false, func() {
		g.mu.Lock()
		if req, ok := g.requests[k]; ok {
			close(req.ready)
		}
		g.mu.Unlock()
	}
}

// writeReplayResponse reuses the accepted wire response without re-running the
// command. In particular, it must not allocate a new room sequence when a
// request is retried on a replacement connection.
func (g *WebSocketGateway) writeReplayResponse(conn *websocket.Conn, ctx context.Context, guestID string, raw []byte) {
	_ = conn.Write(ctx, websocket.MessageText, raw)
}

func (g *WebSocketGateway) cacheResponse(conn *websocket.Conn, raw []byte) {
	var env struct {
		RequestID string `json:"request_id"`
	}
	if json.Unmarshal(raw, &env) != nil || env.RequestID == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	owner := ""
	if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
		owner = c.GuestID
	}
	if owner != "" {
		key := requestKey(owner, env.RequestID)
		if req, exists := g.requests[key]; exists {
			req.responses = append(req.responses, append([]byte(nil), raw...))
			g.requests[key] = req
		}
		return
	}
	// The original connection may have been replaced before its response was
	// written. The claimed request remains the source of truth for ownership.
	for key, req := range g.requests {
		if req.owner == "" || req.connection != conn || !strings.HasSuffix(key, "\x00"+env.RequestID) {
			continue
		}
		req.responses = append(req.responses, append([]byte(nil), raw...))
		g.requests[key] = req
		return
	}
}

func (g *WebSocketGateway) handleStateResync(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	unlockRoom := g.lockRoom(env.RoomID)
	defer unlockRoom()
	rm, ok := g.cfg.Rooms.Get(env.RoomID)
	if !ok || rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	var p struct {
		LastReceivedSequence uint64 `json:"last_received_sequence"`
	}
	if len(env.Payload) > 0 && json.Unmarshal(env.Payload, &p) != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	c, _ := g.cfg.Connections.GetByConn(conn)
	stream := realtime.NewRecipientStream()
	if c != nil {
		if _, connectionStream := c.roomContext(); connectionStream != nil {
			stream = connectionStream
		}
	}
	res, err := realtime.Resync(&rm, rm.Round, guestID, p.LastReceivedSequence, stream)
	if err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	msg := map[string]interface{}{"version": "v1", "type": "event", "name": "state.resynced", "request_id": env.RequestID, "room_id": rm.ID, "payload": res}
	raw, _ := json.Marshal(msg)
	g.writeRaw(conn, ctx, raw, rm.ID)
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
	if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
		c.setRoomContext(roomVal.ID, g.streams.Get(roomVal.ID, guestID))
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
	// Joining mutates the same room membership state as leave/start/actions;
	// serialize the complete lookup, validation, and mutation with them.
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
	rm, found = g.cfg.Rooms.Get(roomID)
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
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
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
		g.writeRaw(ws, context.Background(), raw, "")
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
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
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
		if mem, err := g.lookupGuest(member); err == nil && mem != nil {
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
	g.writeRaw(conn, ctx, raw, rm.ID)
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

func (g *WebSocketGateway) Rooms() *room.Registry       { return g.cfg.Rooms }
func (g *WebSocketGateway) Matches() *MatchHistoryStore { return g.cfg.Matches }

func (g *WebSocketGateway) lookupGuest(id string) (*auth.Guest, error) {
	if g.cfg.Store != nil {
		if guest, err := g.cfg.Store.GetGuest(id); err == nil && guest != nil {
			return guest, nil
		}
	}
	if g.cfg.GuestRepo != nil {
		return g.cfg.GuestRepo.GetGuest(id)
	}
	return nil, auth.ErrGuestNotFound
}

func (g *WebSocketGateway) writeEnvelopeError(conn *websocket.Conn, ctx context.Context, requestID string, ae ActionError) {
	g.writeEnvelopeErrorWithCache(conn, ctx, requestID, ae, true)
}

func (g *WebSocketGateway) writeEnvelopeErrorUncached(conn *websocket.Conn, ctx context.Context, requestID string, ae ActionError) {
	g.writeEnvelopeErrorWithCache(conn, ctx, requestID, ae, false)
}

func (g *WebSocketGateway) writeEnvelopeErrorWithCache(conn *websocket.Conn, ctx context.Context, requestID string, ae ActionError, cache bool) {
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
	if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
		if roomID, _ := c.roomContext(); roomID != "" {
			wrapped["room_id"] = roomID
		}
	}
	raw, _ := json.Marshal(wrapped)
	if cache {
		g.writeRaw(conn, ctx, raw, "")
		return
	}
	g.writeRawUncached(conn, ctx, raw, "")
}

// writeRaw assigns a sequence only to room-scoped server messages. The room
// ID is read from the envelope so callers cannot accidentally use another
// recipient's sequence stream.
func (g *WebSocketGateway) writeRaw(conn *websocket.Conn, ctx context.Context, raw []byte, roomID string) error {
	return g.writeRawWithCache(conn, ctx, raw, roomID, true)
}

func (g *WebSocketGateway) writeRawUncached(conn *websocket.Conn, ctx context.Context, raw []byte, roomID string) error {
	return g.writeRawWithCache(conn, ctx, raw, roomID, false)
}

func (g *WebSocketGateway) writeRawWithCache(conn *websocket.Conn, ctx context.Context, raw []byte, roomID string, cache bool) error {
	if conn == nil {
		return fmt.Errorf("transport: websocket connection is nil")
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal(raw, &envelope); err == nil {
		if roomID == "" {
			if v, ok := envelope["room_id"].(string); ok {
				roomID = v
			}
		}
		guestID := ""
		if c, ok := g.cfg.Connections.GetByConn(conn); ok && c != nil {
			guestID = c.GuestID
		}
		if guestID == "" {
			guestID = g.requestOwner(conn, envelope["request_id"])
		}
		if roomID != "" {
			if guestID != "" {
				envelope["sequence"] = g.streams.Get(roomID, guestID).NextSequence()
				if encoded, err := json.Marshal(envelope); err == nil {
					raw = encoded
				}
			}
		}
	}
	if cache {
		g.cacheResponse(conn, raw)
	}
	return conn.Write(ctx, websocket.MessageText, raw)
}

func (g *WebSocketGateway) requestOwner(conn *websocket.Conn, requestID interface{}) string {
	id, ok := requestID.(string)
	if !ok || id == "" {
		return ""
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for key, req := range g.requests {
		if req.owner != "" && req.connection == conn && strings.HasSuffix(key, "\x00"+id) {
			// Request IDs are scoped by guest, but the cached request map stores
			// the owner alongside the fingerprint so reconnects do not depend on
			// the active connection registry entry.
			return req.owner
		}
	}
	return ""
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
		g.writeRaw(ws, context.Background(), msg, "")
	}
}

func (g *WebSocketGateway) roomUpdatedMessage(rm room.Room, requesterID string, canSeeCode bool) []byte {
	rmSync := rm
	_ = rmSync
	players := make([]map[string]interface{}, 0, len(rm.Members))
	for _, member := range rm.Members {
		isHost := member == rm.Host
		displayName := member
		if mem, err := g.lookupGuest(member); err == nil && mem != nil && mem.DisplayName != "" {
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

type gameDrawPayload struct {
	Source string `json:"source"`
	RoomID string `json:"room_id"`
}

func (g *WebSocketGateway) handleGameDraw(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	var payload gameDrawPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	if roomID == "" {
		roomID = payload.RoomID
	}
	if roomID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
	if payload.Source != "" && payload.Source != "closed_deck" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	if payload.Source == "" {
		// default allow source omitted? protocol requires closed_deck; enforce.
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}

	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("game_not_started"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	// current player id
	currentID := ""
	if rm.Round.CurrentPlayer >= 0 && rm.Round.CurrentPlayer < len(rm.Members) {
		currentID = rm.Members[rm.Round.CurrentPlayer]
	}
	if currentID != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_your_turn"))
		return
	}
	if rm.Round.HasDrawn {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	// capture top card if any
	var drawnCard *game.Card
	if len(rm.Round.ClosedDeck) > 0 {
		c := rm.Round.ClosedDeck[0]
		drawnCard = &c
	}
	playerIndex := rm.Round.CurrentPlayer

	err := rm.Round.Draw(playerIndex)
	// Note: Draw mutates rm.Round (pointer shared); rm struct already has pointer.
	if err != nil {
		if err == game.ErrClosedDeckEmpty {
			if err := g.emitRoundLifecycle(roomID, env.RequestID, "deck_empty"); err != nil {
				g.writeEnvelopeError(conn, ctx, env.RequestID, lifecycleActionError(err))
			}
			return
		}
		// map
		ae := NewActionError("invalid_request")
		if err == game.ErrNotCurrentPlayer {
			ae = NewActionError("not_your_turn")
		} else if err == game.ErrAlreadyDrawn {
			ae = NewActionError("invalid_phase")
		} else {
			ae = NewActionError("invalid_phase")
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
		return
	}

	// send private game.drawn
	if drawnCard != nil {
		cardProj := map[string]interface{}{
			"card_id":         drawCardID(drawnCard.ID),
			"rank":            drawnCard.Rank.String(),
			"suit":            drawSuitName(drawnCard.Suit),
			"is_active_joker": drawnCard.Rank == rm.Round.JokerRank,
		}
		msg := map[string]interface{}{
			"version":    "v1",
			"type":       "event",
			"name":       "game.drawn",
			"request_id": env.RequestID,
			"room_id":    rm.ID,
			"payload": map[string]interface{}{
				"card": cardProj,
			},
		}
		raw, _ := json.Marshal(msg)
		g.writeRaw(conn, ctx, raw, rm.ID)
	}

	// refresh room
	rm2, _ := g.cfg.Rooms.Get(roomID)
	_ = rm2
	// update room round already pointer; no room registry mutation needed for round.
	for _, member := range rm.Members {
		c, ok := g.cfg.Connections.GetByGuest(member)
		if !ok || c == nil {
			continue
		}
		ws, ok := c.Conn.(*websocket.Conn)
		if !ok || ws == nil {
			continue
		}
		// ensure room fetched fresh
		rmf, _ := g.cfg.Rooms.Get(roomID)
		proj, err := realtime.Project(&rmf, rmf.Round, member)
		if err != nil {
			continue
		}
		resp := map[string]interface{}{
			"version":    "v1",
			"type":       "event",
			"name":       "game.state",
			"request_id": env.RequestID,
			"room_id":    rmf.ID,
			"payload": map[string]interface{}{
				"state": proj,
			},
		}
		raw, _ := json.Marshal(resp)
		g.writeRaw(ws, context.Background(), raw, "")
	}
}

func drawCardID(id interface{}) string {
	return fmt.Sprintf("%v", id)
}

func drawSuitName(s game.Suit) string {
	switch s {
	case game.SuitSpades:
		return "spades"
	case game.SuitHearts:
		return "hearts"
	case game.SuitDiamonds:
		return "diamonds"
	case game.SuitClubs:
		return "clubs"
	default:
		return "unknown"
	}
}

type gameDiscardPayload struct {
	CardID string `json:"card_id"`
	RoomID string `json:"room_id"`
}

func (g *WebSocketGateway) handleGameDiscard(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	var payload gameDiscardPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	if roomID == "" {
		roomID = payload.RoomID
	}
	cardID := strings.TrimSpace(payload.CardID)
	if roomID == "" || cardID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()

	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("game_not_started"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	currentID := ""
	if rm.Round.CurrentPlayer >= 0 && rm.Round.CurrentPlayer < len(rm.Members) {
		currentID = rm.Members[rm.Round.CurrentPlayer]
	}
	if currentID != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_your_turn"))
		return
	}
	if !rm.Round.HasDrawn {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	playerIndex := rm.Round.CurrentPlayer
	if playerIndex < 0 || playerIndex >= len(rm.Round.Hands) {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	var selected *game.Card
	for i := range rm.Round.Hands[playerIndex] {
		c := rm.Round.Hands[playerIndex][i]
		if fmt.Sprintf("%v", c.ID) == cardID {
			cc := c
			selected = &cc
			break
		}
	}
	if selected == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
		return
	}

	err := rm.Round.Discard(playerIndex, *selected)
	if err != nil {
		ae := NewActionError("invalid_request")
		if err == game.ErrCardNotInHand {
			ae = NewActionError("card_not_owned")
		} else if err == game.ErrNotCurrentPlayer {
			ae = NewActionError("not_your_turn")
		} else if err == game.ErrDrawRequired {
			ae = NewActionError("invalid_phase")
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
		return
	}
	if rm.Round.Ended {
		if err := g.emitRoundLifecycle(roomID, env.RequestID, "game_over"); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, lifecycleActionError(err))
		}
		return
	}

	rm2, _ := g.cfg.Rooms.Get(roomID)
	for _, member := range rm2.Members {
		c, ok := g.cfg.Connections.GetByGuest(member)
		if !ok || c == nil {
			continue
		}
		ws, ok := c.Conn.(*websocket.Conn)
		if !ok || ws == nil {
			continue
		}
		proj, err := realtime.Project(&rm2, rm2.Round, member)
		if err != nil {
			continue
		}
		resp := map[string]interface{}{
			"version":    "v1",
			"type":       "event",
			"name":       "game.state",
			"request_id": env.RequestID,
			"room_id":    rm2.ID,
			"payload": map[string]interface{}{
				"state": proj,
			},
		}
		raw, _ := json.Marshal(resp)
		g.writeRaw(ws, context.Background(), raw, "")
	}
}

type gameLaydownPayload struct {
	CardIDs []interface{} `json:"card_ids"`
	Type    string        `json:"type"`
	RoomID  string        `json:"room_id"`
}

func (g *WebSocketGateway) handleGameLaydown(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	var payload gameLaydownPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	if roomID == "" {
		roomID = payload.RoomID
	}
	if roomID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("game_not_started"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	currentID := ""
	if rm.Round.CurrentPlayer >= 0 && rm.Round.CurrentPlayer < len(rm.Members) {
		currentID = rm.Members[rm.Round.CurrentPlayer]
	}
	if currentID != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_your_turn"))
		return
	}
	if !rm.Round.HasDrawn {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	if len(payload.CardIDs) < 3 {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_combination"))
		return
	}
	// convert ids to strings
	idStrings := make([]string, 0, len(payload.CardIDs))
	for _, v := range payload.CardIDs {
		idStrings = append(idStrings, fmt.Sprintf("%v", v))
	}
	// duplicates
	seen := make(map[string]bool)
	for _, id := range idStrings {
		if seen[id] {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
		seen[id] = true
	}

	playerIndex := rm.Round.CurrentPlayer
	cards := make([]game.Card, 0, len(idStrings))
	for _, id := range idStrings {
		foundCard := false
		for _, c := range rm.Round.Hands[playerIndex] {
			if fmt.Sprintf("%v", c.ID) == id {
				cards = append(cards, c)
				foundCard = true
				break
			}
		}
		if !foundCard {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
			return
		}
	}
	// type validation
	jokerRank := rm.Round.JokerRank
	if payload.Type != "" {
		t := strings.ToLower(strings.TrimSpace(payload.Type))
		if t == "parallel" {
			t = "paralel"
		}
		var err error
		switch t {
		case "seri":
			err = game.ValidateSeri(cards, jokerRank)
		case "paralel":
			err = game.ValidateParalel(cards, jokerRank)
		default:
			err = game.ErrInvalidCombination
		}
		if err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_combination"))
			return
		}
	} else {
		if !game.IsValidSeri(cards, jokerRank) && !game.IsValidParalel(cards, jokerRank) {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_combination"))
			return
		}
	}

	if err := rm.Round.LayCombination(playerIndex, cards); err != nil {
		ae := NewActionError("invalid_combination")
		if err == game.ErrCardNotInHand || err == game.ErrCombinationDuplicateCard {
			ae = NewActionError("card_not_owned")
		} else if err == game.ErrTooFewCombinationCards {
			ae = NewActionError("invalid_combination")
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
		return
	}

	rm2, _ := g.cfg.Rooms.Get(roomID)
	for _, member := range rm2.Members {
		c, ok := g.cfg.Connections.GetByGuest(member)
		if !ok || c == nil {
			continue
		}
		ws, ok := c.Conn.(*websocket.Conn)
		if !ok || ws == nil {
			continue
		}
		proj, err := realtime.Project(&rm2, rm2.Round, member)
		if err != nil {
			continue
		}
		resp := map[string]interface{}{
			"version":    "v1",
			"type":       "event",
			"name":       "game.state",
			"request_id": env.RequestID,
			"room_id":    rm2.ID,
			"payload": map[string]interface{}{
				"state": proj,
			},
		}
		raw, _ := json.Marshal(resp)
		g.writeRaw(ws, context.Background(), raw, "")
	}
}

type gameCloseCombinationPayload struct {
	CardIDs []interface{} `json:"card_ids"`
	Type    string        `json:"type"`
}
type gameClosePayload struct {
	ClosingCardID string                        `json:"closing_card_id"`
	Combinations  []gameCloseCombinationPayload `json:"combinations"`
	RoomID        string                        `json:"room_id"`
}

func parseCloseCombinationType(t string, cards []game.Card, jokerRank game.Rank) (game.CombinationKind, error) {
	t = strings.ToLower(strings.TrimSpace(t))
	if t == "parallel" {
		t = "paralel"
	}
	if t == "" {
		if game.IsValidSeri(cards, jokerRank) {
			return game.CombinationSeri, nil
		}
		if game.IsValidParalel(cards, jokerRank) {
			return game.CombinationParalel, nil
		}
		return 0, game.ErrInvalidCombination
	}
	switch t {
	case "seri":
		if err := game.ValidateSeri(cards, jokerRank); err != nil {
			return 0, err
		}
		return game.CombinationSeri, nil
	case "paralel":
		if err := game.ValidateParalel(cards, jokerRank); err != nil {
			return 0, err
		}
		return game.CombinationParalel, nil
	default:
		return 0, game.ErrInvalidCombination
	}
}

func (g *WebSocketGateway) handleGameClose(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	var payload gameClosePayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	if roomID == "" {
		roomID = payload.RoomID
	}
	closeCardID := strings.TrimSpace(payload.ClosingCardID)
	if roomID == "" || closeCardID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()

	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("game_not_started"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	currentID := ""
	if rm.Round.CurrentPlayer >= 0 && rm.Round.CurrentPlayer < len(rm.Members) {
		currentID = rm.Members[rm.Round.CurrentPlayer]
	}
	if currentID != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_your_turn"))
		return
	}
	if !rm.Round.HasDrawn {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	playerIndex := rm.Round.CurrentPlayer
	if playerIndex < 0 || playerIndex >= len(rm.Round.Hands) {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	// resolve closing card
	var closeCard *game.Card
	for i := range rm.Round.Hands[playerIndex] {
		c := rm.Round.Hands[playerIndex][i]
		if fmt.Sprintf("%v", c.ID) == closeCardID {
			cc := c
			closeCard = &cc
			break
		}
	}
	if closeCard == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
		return
	}

	// resolve combinations if supplied
	var comboArgs []interface{}
	comboArgs = append(comboArgs, *closeCard)
	if len(payload.Combinations) > 0 {
		var combos []game.Combination
		for _, pc := range payload.Combinations {
			idStrings := make([]string, 0, len(pc.CardIDs))
			for _, v := range pc.CardIDs {
				idStrings = append(idStrings, fmt.Sprintf("%v", v))
			}
			cards := make([]game.Card, 0, len(idStrings))
			for _, id := range idStrings {
				foundCard := false
				for _, c := range rm.Round.Hands[playerIndex] {
					if fmt.Sprintf("%v", c.ID) == id {
						cards = append(cards, c)
						foundCard = true
						break
					}
				}
				if !foundCard {
					g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
					return
				}
			}
			kind, err := parseCloseCombinationType(pc.Type, cards, rm.Round.JokerRank)
			if err != nil {
				g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_combination"))
				return
			}
			combos = append(combos, game.Combination{Kind: kind, Cards: cards})
		}
		comboArgs = []interface{}{*closeCard, combos}
	}

	err := rm.Round.Close(playerIndex, comboArgs...)
	if err != nil {
		ae := NewActionError("invalid_combination")
		if err == game.ErrCardNotInHand {
			ae = NewActionError("card_not_owned")
		} else if err == game.ErrRoundAlreadyEnded || err == game.ErrRoundEnded {
			ae = NewActionError("invalid_phase")
		} else if err == game.ErrDrawRequired {
			ae = NewActionError("invalid_phase")
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
		return
	}

	if err := g.emitRoundLifecycle(roomID, env.RequestID, "close"); err != nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, lifecycleActionError(err))
	}
}

func matchScoreboard(rm room.Room, scores []int) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rm.Members))
	for i, member := range rm.Members {
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		out = append(out, map[string]interface{}{"player_id": member, "display_name": displayNameForRoomMember(member), "cumulative_score": score})
	}
	return out
}

func finalRanking(rm room.Room, m *game.Match) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rm.Members))
	for _, rank := range m.FinalRankings() {
		member := rm.Members[rank.Player]
		out = append(out, map[string]interface{}{"player_id": member, "display_name": displayNameForRoomMember(member), "cumulative_score": rank.Score, "rank": rank.Rank})
	}
	return out
}

func (g *WebSocketGateway) displayName(id string) string {
	if guest, err := g.lookupGuest(id); err == nil && guest != nil && guest.DisplayName != "" {
		return guest.DisplayName
	}
	return id
}

func displayNameForRoomMember(id string) string { return id }

func lifecycleActionError(err error) ActionError {
	if ae, ok := err.(ActionError); ok {
		return ae
	}
	if err == game.ErrMatchAlreadyEnded || err == game.ErrMatchEnded || err == game.ErrInvalidMatchRound || err == game.ErrFourthRound {
		return NewActionError("invalid_phase")
	}
	return NewActionError("invalid_request")
}

func (g *WebSocketGateway) emitRoundLifecycle(roomID, requestID, defaultReason string) error {
	rm, ok := g.cfg.Rooms.Get(roomID)
	if !ok || rm.Round == nil {
		return NewActionError("not_found")
	}
	round := rm.Round
	roundNumber := 1
	if rm.Match != nil {
		roundNumber = rm.Match.RoundNumber
	}
	scores := round.ScoresSlice()
	if len(scores) != len(rm.Members) {
		return NewActionError("invalid_request")
	}
	match := rm.Match
	if match == nil {
		match = game.NewMatch(len(rm.Members), round)
	}
	if len(match.CumulativeScores) != len(rm.Members) {
		return NewActionError("invalid_request")
	}

	// Build the next round before mutating Match. Deal and NewRound errors are
	// returned to the caller; they must not leave a completed match without an
	// active round.
	var next *game.Round
	if match.RoundNumber < 3 {
		var err error
		next, err = buildNextRound(len(rm.Members))
		if err != nil {
			return err
		}
	}
	if err := g.cfg.Rooms.ApplyRoundTransition(rm.ID, next, func(stored *room.Room) error {
		if err := match.CompleteRound(scores, round); err != nil {
			return err
		}
		if stored.Match == nil {
			stored.Match = match
		}
		if next != nil {
			return match.StartNextRound(next)
		}
		return nil
	}); err != nil {
		return err
	}
	var found bool
	rm, found = g.cfg.Rooms.Get(roomID)
	if !found || rm.Match == nil || rm.Round == nil {
		return NewActionError("not_found")
	}
	matchEnded := rm.Match.Ended
	// game.state is deliberately first: it acknowledges the accepted action.
	if err := g.broadcastGameState(roomID, requestID); err != nil {
		return err
	}
	reason := round.EndReason
	if reason == "" {
		reason = round.Reason
	}
	if reason == "" {
		reason = defaultReason
	}
	scoreList := make([]map[string]interface{}, 0, len(rm.Members))
	for i, member := range rm.Members {
		entry := map[string]interface{}{"player_id": member, "score_status": "calculated", "cumulative_score": 0}
		if i < len(scores) {
			entry["cumulative_score"] = rm.Match.CumulativeScores[i]
			if round.GameOver && scores[i] != -50 {
				entry["score_status"] = "not_calculated"
			} else {
				entry["round_score"] = scores[i]
			}
		}
		scoreList = append(scoreList, entry)
	}
	ended := map[string]interface{}{"version": "v1", "type": "event", "name": "round.ended", "request_id": requestID, "room_id": rm.ID,
		"payload": map[string]interface{}{"round_number": roundNumber, "reason": reason, "scores": scoreList}}
	if err := g.broadcastLifecycleMessage(rm, ended); err != nil {
		return err
	}
	updated := map[string]interface{}{"version": "v1", "type": "event", "name": "match.updated", "request_id": requestID, "room_id": rm.ID,
		"payload": map[string]interface{}{"round_number": rm.Match.RoundNumber, "scoreboard": g.matchScoreboard(rm, rm.Match.CumulativeScores)}}
	if err := g.broadcastLifecycleMessage(rm, updated); err != nil {
		return err
	}
	if matchEnded {
		endedMatch := map[string]interface{}{"version": "v1", "type": "event", "name": "match.ended", "request_id": requestID, "room_id": rm.ID,
			"payload": map[string]interface{}{"round_number": roundNumber, "final_ranking": g.finalRanking(rm, rm.Match)}}
		if err := g.broadcastLifecycleMessage(rm, endedMatch); err != nil {
			return err
		}
	}
	return nil
}

func buildNextRound(playerCount int) (*game.Round, error) {
	deck := game.StandardDeck()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	deal, err := game.Deal(deck, playerCount)
	if err != nil {
		return nil, err
	}
	return game.NewRound(playerCount, deal.Hands, deal.ClosedDeck, deal.JokerRank)
}

func (g *WebSocketGateway) broadcastLifecycleMessage(rm room.Room, msg map[string]interface{}) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	for _, member := range rm.Members {
		if c, ok := g.cfg.Connections.GetByGuest(member); ok && c != nil {
			if ws, ok := c.Conn.(*websocket.Conn); ok && ws != nil {
				if err := g.writeRaw(ws, context.Background(), raw, rm.ID); err != nil {
					g.cfg.Connections.UnregisterIfCurrent(c)
				}
			}
		}
	}
	return nil
}

func (g *WebSocketGateway) matchScoreboard(rm room.Room, scores []int) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rm.Members))
	for i, member := range rm.Members {
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		out = append(out, map[string]interface{}{"player_id": member, "display_name": g.displayName(member), "cumulative_score": score})
	}
	return out
}

func (g *WebSocketGateway) finalRanking(rm room.Room, m *game.Match) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rm.Members))
	for _, rank := range m.FinalRankings() {
		member := rm.Members[rank.Player]
		out = append(out, map[string]interface{}{"player_id": member, "display_name": g.displayName(member), "cumulative_score": rank.Score, "rank": rank.Rank})
	}
	return out
}

func (g *WebSocketGateway) broadcastGameState(roomID, requestID string) error {
	rm, ok := g.cfg.Rooms.Get(roomID)
	if !ok || rm.Round == nil {
		return NewActionError("not_found")
	}
	for _, member := range rm.Members {
		if c, ok := g.cfg.Connections.GetByGuest(member); ok && c != nil {
			if ws, ok := c.Conn.(*websocket.Conn); ok && ws != nil {
				proj, err := realtime.Project(&rm, rm.Round, member)
				if err != nil {
					continue
				}
				msg := map[string]interface{}{"version": "v1", "type": "event", "name": "game.state", "request_id": requestID, "room_id": rm.ID, "payload": map[string]interface{}{"state": proj}}
				raw, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				if err := g.writeRaw(ws, context.Background(), raw, rm.ID); err != nil {
					g.cfg.Connections.UnregisterIfCurrent(c)
				}
			}
		}
	}
	return nil
}

type gameTakeDiscardPayload struct {
	DiscardCardID    interface{}   `json:"discard_card_id"`
	JagoCardIDs      []interface{} `json:"jago_card_ids"`
	ExtensionCardIDs []interface{} `json:"extension_card_ids"`
	RoomID           string        `json:"room_id"`
}

func (g *WebSocketGateway) handleGameTakeDiscard(conn *websocket.Conn, ctx context.Context, guestID string, env CommandEnvelope) {
	roomID := env.RoomID
	var payload gameTakeDiscardPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
	}
	if roomID == "" {
		roomID = payload.RoomID
	}
	if roomID == "" {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	unlockRoom := g.lockRoom(roomID)
	defer unlockRoom()
	if payload.DiscardCardID == nil || len(payload.JagoCardIDs) != 2 {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	discardID := fmt.Sprintf("%v", payload.DiscardCardID)
	jagoIDs := make([]string, 0, 2)
	for _, v := range payload.JagoCardIDs {
		jagoIDs = append(jagoIDs, fmt.Sprintf("%v", v))
	}
	extIDs := make([]string, 0, len(payload.ExtensionCardIDs))
	for _, v := range payload.ExtensionCardIDs {
		extIDs = append(extIDs, fmt.Sprintf("%v", v))
	}

	rm, found := g.cfg.Rooms.Get(roomID)
	if !found {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_found"))
		return
	}
	if rm.Round == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("game_not_started"))
		return
	}
	if rm.Match != nil && rm.Match.Ended {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	currentID := ""
	if rm.Round.CurrentPlayer >= 0 && rm.Round.CurrentPlayer < len(rm.Members) {
		currentID = rm.Members[rm.Round.CurrentPlayer]
	}
	if currentID != guestID {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("not_your_turn"))
		return
	}
	if rm.Round.HasDrawn {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}
	playerIndex := rm.Round.CurrentPlayer
	if playerIndex < 0 || playerIndex >= len(rm.Round.Hands) {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_phase"))
		return
	}

	// selected discard from pile
	var selected *game.Card
	for i := range rm.Round.DiscardPile {
		c := rm.Round.DiscardPile[i]
		if fmt.Sprintf("%v", c.ID) == discardID {
			cc := c
			selected = &cc
			break
		}
	}
	if selected == nil {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("unknown_card"))
		return
	}
	handCards := rm.Round.Hands[playerIndex]
	findHand := func(id string) (game.Card, bool) {
		for _, c := range handCards {
			if fmt.Sprintf("%v", c.ID) == id {
				return c, true
			}
		}
		return game.Card{}, false
	}
	findPile := func(id string) (game.Card, bool) {
		for _, c := range rm.Round.DiscardPile {
			if fmt.Sprintf("%v", c.ID) == id {
				return c, true
			}
		}
		return game.Card{}, false
	}

	// distinct jago ids
	if jagoIDs[0] == jagoIDs[1] {
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
		return
	}
	seen := map[string]bool{discardID: true, jagoIDs[0]: true, jagoIDs[1]: true}
	for _, id := range extIDs {
		if seen[id] {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("invalid_request"))
			return
		}
		seen[id] = true
	}
	for _, id := range jagoIDs {
		if seenDiscardOrJago(id, discardID) {
		}
	}

	// resolve jago cards (must be in hand)
	jagoCards := make([]game.Card, 0, 2)
	for _, id := range jagoIDs {
		c, ok := findHand(id)
		if !ok {
			g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
			return
		}
		jagoCards = append(jagoCards, c)
	}
	// resolve extensions
	extCards := make([]game.Card, 0, len(extIDs))
	for _, id := range extIDs {
		if c, ok := findHand(id); ok {
			extCards = append(extCards, c)
			continue
		}
		if c, ok := findPile(id); ok {
			extCards = append(extCards, c)
			continue
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, NewActionError("card_not_owned"))
		return
	}

	// prepare call args
	args := []interface{}{*selected}
	for _, c := range jagoCards {
		args = append(args, c)
	}
	for _, c := range extCards {
		args = append(args, c)
	}

	err := rm.Round.TakeDiscard(playerIndex, args...)
	if err != nil {
		ae := NewActionError("invalid_jago")
		if err == game.ErrInvalidDiscardWindow {
			ae = NewActionError("invalid_discard_window")
		} else if err == game.ErrCardNotInHand {
			ae = NewActionError("card_not_owned")
		} else if err == game.ErrInvalidJago {
			ae = NewActionError("invalid_jago")
		} else if err == game.ErrCombinationDuplicateCard {
			ae = NewActionError("invalid_request")
		} else if err == game.ErrAlreadyDrawn || err == game.ErrDrawRequired {
			ae = NewActionError("invalid_phase")
		} else if err == game.ErrNotCurrentPlayer {
			ae = NewActionError("not_your_turn")
		}
		g.writeEnvelopeError(conn, ctx, env.RequestID, ae)
		return
	}

	if rm.Round.GameOver {
		if err := g.emitRoundLifecycle(roomID, env.RequestID, "game_over"); err != nil {
			g.writeEnvelopeError(conn, ctx, env.RequestID, lifecycleActionError(err))
		}
		return
	}
	g.broadcastGameState(roomID, env.RequestID)
}

func seenDiscardOrJago(id, discardID string) bool { return false }
func (g *WebSocketGateway) Handler() http.Handler { return g }
