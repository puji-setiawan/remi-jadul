package transport

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"remi-jadul/server/internal/room"
)

const roomsEndpointPath = "/api/v1/rooms"

// PublicRoomSummary represents an entry in GET /api/v1/rooms.
type PublicRoomSummary struct {
	RoomID        string `json:"room_id"`
	Visibility    string `json:"visibility"`
	Capacity      int    `json:"capacity"`
	OccupiedSeats int    `json:"occupied_seats"`
	Status        string `json:"status"`
}

type roomsDiscoveryResponse struct {
	Version   string              `json:"version"`
	RequestID string              `json:"request_id"`
	Rooms     []PublicRoomSummary `json:"rooms"`
}

// RoomDiscoveryHandler serves GET /api/v1/rooms.
type RoomDiscoveryHandler struct {
	rooms *room.Registry
}

// Aliases for RoomDiscoveryHandler.
type RoomsHandler = RoomDiscoveryHandler
type RoomListHandler = RoomDiscoveryHandler
type PublicRoomsHandler = RoomDiscoveryHandler

func NewRoomDiscoveryHandler(args ...interface{}) *RoomDiscoveryHandler {
	var reg *room.Registry
	for _, arg := range args {
		switch v := arg.(type) {
		case *room.Registry:
			reg = v
		case *WebSocketGateway:
			if v != nil {
				reg = v.Rooms()
			}
		}
	}
	if reg == nil {
		reg = room.NewRegistry()
	}
	return &RoomDiscoveryHandler{rooms: reg}
}

func NewRoomsHandler(args ...interface{}) *RoomDiscoveryHandler {
	return NewRoomDiscoveryHandler(args...)
}
func NewRoomListHandler(args ...interface{}) *RoomDiscoveryHandler {
	return NewRoomDiscoveryHandler(args...)
}
func NewPublicRoomsHandler(args ...interface{}) *RoomDiscoveryHandler {
	return NewRoomDiscoveryHandler(args...)
}

func (h *RoomDiscoveryHandler) Rooms() *room.Registry { return h.rooms }

func (h *RoomDiscoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != roomsEndpointPath {
		http.NotFound(w, r)
		return
	}
	requestID := resolveRequestID(r)
	if r.Method != http.MethodGet {
		writeRESTError(w, requestID, http.StatusBadRequest, CodeInvalidRequest, "Method is not supported.")
		return
	}

	visValues, hasVis := r.URL.Query()["visibility"]
	if hasVis {
		if len(visValues) != 1 || strings.ToLower(strings.TrimSpace(visValues[0])) != "public" {
			writeRESTError(w, requestID, http.StatusBadRequest, CodeInvalidRequest, "Only visibility=public is supported.")
			return
		}
	}

	publicRooms := h.rooms.ListPublicRooms()
	sort.Slice(publicRooms, func(i, j int) bool {
		return publicRooms[i].ID < publicRooms[j].ID
	})

	summaries := make([]PublicRoomSummary, 0, len(publicRooms))
	for _, rm := range publicRooms {
		summaries = append(summaries, PublicRoomSummary{
			RoomID:        rm.ID,
			Visibility:    "public",
			Capacity:      rm.Capacity,
			OccupiedSeats: len(rm.Members),
			Status:        rm.Status,
		})
	}

	resp := roomsDiscoveryResponse{
		Version:   ProtocolVersion,
		RequestID: requestID,
		Rooms:     summaries,
	}
	writeRESTJSON(w, http.StatusOK, resp)
}

func resolveRequestID(r *http.Request) string {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = generateRequestID()
	}
	return requestID
}

func writeRESTError(w http.ResponseWriter, requestID string, status int, code, message string) {
	writeRESTJSON(w, status, restErrorResponse{
		Version:   ProtocolVersion,
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}

func writeRESTJSON(w http.ResponseWriter, status int, payload interface{}) {
	raw, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}
