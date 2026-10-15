package transport

import (
	"errors"
	"sync"

	"remi-jadul/server/internal/realtime"
)

// Connection registration errors.
var (
	ErrConnectionNotFound  = errors.New("transport: connection not found")
	ErrDuplicateConnection = errors.New("transport: duplicate connection")
)

// Connection represents one authenticated WebSocket connection.
type Connection struct {
	mu      sync.RWMutex
	ID      string
	Token   string
	GuestID string
	Stream  *realtime.RecipientStream
	Meta    map[string]interface{}
	// Conn may hold a transport-specific connection object.
	Conn interface{}
}

func (c *Connection) clearRoomContext() {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Meta, "room_id")
	c.Stream = nil
}

func (c *Connection) setRoomContext(roomID string, stream *realtime.RecipientStream) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Meta == nil {
		c.Meta = make(map[string]interface{})
	}
	c.Meta["room_id"] = roomID
	c.Stream = stream
}

func (c *Connection) roomContext() (string, *realtime.RecipientStream) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	roomID, _ := c.Meta["room_id"].(string)
	return roomID, c.Stream
}

// ConnectionRegistry is a thread-safe authenticated connection registry.
type ConnectionRegistry struct {
	mu      sync.RWMutex
	byToken map[string]*Connection
	byGuest map[string]string // guest ID -> token
}

func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{byToken: make(map[string]*Connection), byGuest: make(map[string]string)}
}

// Aliases.
func NewConnectionManager() *ConnectionRegistry { return NewConnectionRegistry() }
func NewConnectionStore() *ConnectionRegistry   { return NewConnectionRegistry() }
func NewRegistry() *ConnectionRegistry          { return NewConnectionRegistry() }
func NewAuthenticatedConnections() *ConnectionRegistry {
	return NewConnectionRegistry()
}

// Register stores a connection and replaces any prior connection for the same guest.
func (r *ConnectionRegistry) Register(conn *Connection) string {
	if conn == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if conn.ID == "" {
		conn.ID = conn.Token
	}
	if conn.Token == "" {
		conn.Token = conn.ID
	}
	if conn.ID == "" {
		conn.ID = conn.GuestID
	}
	if conn.Token == "" {
		conn.Token = conn.ID
	}
	// Remove previous connection for guest.
	if prevToken, ok := r.byGuest[conn.GuestID]; ok && prevToken != conn.Token {
		delete(r.byToken, prevToken)
	}
	// If token exists, replace.
	r.byToken[conn.Token] = conn
	r.byGuest[conn.GuestID] = conn.Token
	return conn.Token
}

func (r *ConnectionRegistry) RegisterConnection(conn *Connection) string {
	return r.Register(conn)
}
func (r *ConnectionRegistry) Add(conn *Connection) string           { return r.Register(conn) }
func (r *ConnectionRegistry) AddConnection(conn *Connection) string { return r.Register(conn) }

// Unregister by token.
func (r *ConnectionRegistry) UnregisterToken(token string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	conn, ok := r.byToken[token]
	if !ok {
		return false
	}
	delete(r.byToken, token)
	if current, ok2 := r.byGuest[conn.GuestID]; ok2 && current == token {
		delete(r.byGuest, conn.GuestID)
	}
	return true
}

// UnregisterIfCurrent removes conn only when it is still the guest's active
// connection. A failed write from an old connection must not remove a newer
// connection registered during reconnect.
func (r *ConnectionRegistry) UnregisterIfCurrent(conn *Connection) bool {
	if conn == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	token, ok := r.byGuest[conn.GuestID]
	if !ok || r.byToken[token] != conn {
		return false
	}
	delete(r.byGuest, conn.GuestID)
	delete(r.byToken, token)
	return true
}

// Unregister by guest ID.
func (r *ConnectionRegistry) UnregisterGuest(guestID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	token, ok := r.byGuest[guestID]
	if !ok {
		return false
	}
	delete(r.byGuest, guestID)
	delete(r.byToken, token)
	return true
}

func (r *ConnectionRegistry) Unregister(args ...interface{}) bool {
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			// Try guest first.
			if r.hasGuest(v) {
				return r.UnregisterGuest(v)
			}
			return r.UnregisterToken(v)
		case *Connection:
			if v != nil {
				if v.Token != "" {
					return r.UnregisterToken(v.Token)
				}
				if v.ID != "" {
					return r.UnregisterToken(v.ID)
				}
				return r.UnregisterGuest(v.GuestID)
			}
		}
	}
	return false
}

func (r *ConnectionRegistry) hasGuest(guestID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.byGuest[guestID]
	return ok
}

func (r *ConnectionRegistry) UnregisterConnection(args ...interface{}) bool {
	return r.Unregister(args...)
}
func (r *ConnectionRegistry) Remove(args ...interface{}) bool { return r.Unregister(args...) }
func (r *ConnectionRegistry) RemoveConnection(args ...interface{}) bool {
	return r.Unregister(args...)
}
func (r *ConnectionRegistry) Delete(args ...interface{}) bool { return r.Unregister(args...) }

// Get by guest or token.
func (r *ConnectionRegistry) GetByGuest(guestID string) (*Connection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.byGuest[guestID]
	if !ok {
		return nil, false
	}
	conn, ok := r.byToken[token]
	return conn, ok
}

func (r *ConnectionRegistry) GetByToken(token string) (*Connection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, ok := r.byToken[token]
	return conn, ok
}

func (r *ConnectionRegistry) GetByConn(target interface{}) (*Connection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.byToken {
		if conn != nil && conn.Conn == target {
			return conn, true
		}
	}
	return nil, false
}

func (r *ConnectionRegistry) Get(args ...interface{}) (*Connection, bool) {
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			if conn, ok := r.GetByGuest(v); ok {
				return conn, true
			}
			if conn, ok := r.GetByToken(v); ok {
				return conn, true
			}
		case *Connection:
			if v != nil {
				if conn, ok := r.GetByToken(v.Token); ok {
					return conn, true
				}
				if conn, ok := r.GetByGuest(v.GuestID); ok {
					return conn, true
				}
			}
		}
	}
	return nil, false
}

func (r *ConnectionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byToken)
}
func (r *ConnectionRegistry) Len() int  { return r.Count() }
func (r *ConnectionRegistry) Size() int { return r.Count() }

// Broadcast delivers a personalized payload via sender callback.
// sender receives (recipientID, connection); return error is ignored except for skip tracking.
func (r *ConnectionRegistry) Broadcast(recipients []string, sender func(string, *Connection) error) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var failed []string
	for _, recipient := range recipients {
		conn, ok := r._getByGuestNoLock(recipient)
		if !ok || conn == nil {
			failed = append(failed, recipient)
			continue
		}
		if err := sender(recipient, conn); err != nil {
			failed = append(failed, recipient)
		}
	}
	return failed
}

func (r *ConnectionRegistry) _getByGuestNoLock(guestID string) (*Connection, bool) {
	token, ok := r.byGuest[guestID]
	if !ok {
		return nil, false
	}
	conn, ok := r.byToken[token]
	return conn, ok
}

// BroadcastPersonalized alias accepts a builder.
func (r *ConnectionRegistry) BroadcastPersonalized(recipients []string, build func(string) interface{}, send func(string, *Connection, interface{}) error) []string {
	return r.Broadcast(recipients, func(recipient string, conn *Connection) error {
		payload := build(recipient)
		return send(recipient, conn, payload)
	})
}

func (r *ConnectionRegistry) BroadcastTo(recipients []string, send func(string, *Connection) error) []string {
	return r.Broadcast(recipients, send)
}
func (r *ConnectionRegistry) Deliver(recipients []string, send func(string, *Connection) error) []string {
	return r.Broadcast(recipients, send)
}
func (r *ConnectionRegistry) SendTo(recipients []string, send func(string, *Connection) error) []string {
	return r.Broadcast(recipients, send)
}
