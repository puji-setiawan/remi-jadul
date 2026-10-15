package transport

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"

	"remi-jadul/server/internal/persistence"
)

const matchesEndpointPrefix = "/api/v1/matches/"

// MatchRankingEntry represents one player's final ranking in GET /api/v1/matches/{match_id}.
type MatchRankingEntry struct {
	PlayerID        string `json:"player_id"`
	DisplayName     string `json:"display_name"`
	CumulativeScore int    `json:"cumulative_score"`
	Rank            int    `json:"rank"`
}

type matchHistoryResponse struct {
	Version      string              `json:"version"`
	RequestID    string              `json:"request_id"`
	MatchID      string              `json:"match_id"`
	RoundCount   int                 `json:"round_count"`
	FinalRanking []MatchRankingEntry `json:"final_ranking"`
}

// MatchHistoryLoader loads a completed match by ID.
type MatchHistoryLoader interface {
	Load(matchID string) (persistence.CompletedMatch, error)
}

// MatchHistorySaver persists a completed match.
type MatchHistorySaver interface {
	Save(match persistence.CompletedMatch) error
}

// MatchHistoryStore is an in-memory completed match store with optional fallback/backing repository.
type MatchHistoryStore struct {
	mu      sync.RWMutex
	matches map[string]persistence.CompletedMatch
	repo    MatchHistoryLoader
	saver   MatchHistorySaver
}

// Aliases.
type MatchStore = MatchHistoryStore
type InMemoryMatchStore = MatchHistoryStore

func NewMatchHistoryStore(args ...interface{}) *MatchHistoryStore {
	s := &MatchHistoryStore{
		matches: make(map[string]persistence.CompletedMatch),
	}
	for _, arg := range args {
		if arg == nil {
			continue
		}
		if repo, ok := arg.(*persistence.MatchRepository); ok {
			if repo != nil {
				s.repo = repo
				s.saver = repo
			}
			continue
		}
		if loader, ok := arg.(MatchHistoryLoader); ok {
			s.repo = loader
		}
		if saver, ok := arg.(MatchHistorySaver); ok {
			s.saver = saver
		}
	}
	return s
}

func NewMatchStore(args ...interface{}) *MatchHistoryStore {
	return NewMatchHistoryStore(args...)
}

func (s *MatchHistoryStore) Save(match persistence.CompletedMatch) error {
	if strings.TrimSpace(match.MatchID) == "" {
		return errors.New("transport: match_id is required")
	}
	if match.RoundCount <= 0 {
		match.RoundCount = 3
	}
	if s.saver != nil {
		if err := s.saver.Save(match); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	players := make([]persistence.CompletedMatchPlayer, len(match.Players))
	copy(players, match.Players)
	scores := make([]persistence.CompletedRoundScore, len(match.RoundScores))
	copy(scores, match.RoundScores)
	cp := match
	cp.Players = players
	cp.RoundScores = scores
	s.matches[match.MatchID] = cp
	return nil
}

func (s *MatchHistoryStore) SaveCompletedMatch(match persistence.CompletedMatch) error {
	return s.Save(match)
}

func (s *MatchHistoryStore) Load(matchID string) (persistence.CompletedMatch, error) {
	id := strings.TrimSpace(matchID)
	if id == "" {
		return persistence.CompletedMatch{}, persistence.ErrCompletedMatchNotFound
	}
	s.mu.RLock()
	m, ok := s.matches[id]
	s.mu.RUnlock()
	if ok {
		return m, nil
	}
	if s.repo != nil {
		loaded, err := s.repo.Load(id)
		if err == nil {
			return loaded, nil
		}
		return persistence.CompletedMatch{}, err
	}
	return persistence.CompletedMatch{}, persistence.ErrCompletedMatchNotFound
}

func (s *MatchHistoryStore) Get(matchID string) (persistence.CompletedMatch, error) {
	return s.Load(matchID)
}

// MatchHistoryHandler serves GET /api/v1/matches/{match_id}.
type MatchHistoryHandler struct {
	loader MatchHistoryLoader
	store  *MatchHistoryStore
}

// Aliases.
type MatchesHandler = MatchHistoryHandler
type MatchHandler = MatchHistoryHandler

func NewMatchHistoryHandler(args ...interface{}) *MatchHistoryHandler {
	var store *MatchHistoryStore
	var loader MatchHistoryLoader
	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch v := arg.(type) {
		case *MatchHistoryStore:
			if v != nil {
				store = v
				loader = v
			}
		case *persistence.MatchRepository:
			if v != nil {
				loader = v
			}
		case MatchHistoryLoader:
			loader = v
		}
	}
	if store == nil {
		if loader != nil {
			store = NewMatchHistoryStore(loader)
		} else {
			store = NewMatchHistoryStore()
		}
	}
	if loader == nil {
		loader = store
	}
	return &MatchHistoryHandler{loader: loader, store: store}
}

func NewMatchesHandler(args ...interface{}) *MatchHistoryHandler {
	return NewMatchHistoryHandler(args...)
}
func NewMatchHandler(args ...interface{}) *MatchHistoryHandler {
	return NewMatchHistoryHandler(args...)
}

func (h *MatchHistoryHandler) Store() *MatchHistoryStore { return h.store }

func (h *MatchHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := resolveRequestID(r)
	if !strings.HasPrefix(r.URL.Path, matchesEndpointPrefix) {
		writeRESTError(w, requestID, http.StatusNotFound, CodeNotFound, "Match not found.")
		return
	}
	if r.Method != http.MethodGet {
		writeRESTError(w, requestID, http.StatusBadRequest, CodeInvalidRequest, "Method is not supported.")
		return
	}

	matchID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, matchesEndpointPrefix))
	if matchID == "" || strings.Contains(matchID, "/") {
		writeRESTError(w, requestID, http.StatusNotFound, CodeNotFound, "Match not found.")
		return
	}

	match, err := h.loader.Load(matchID)
	if err != nil {
		if errors.Is(err, persistence.ErrCompletedMatchNotFound) {
			writeRESTError(w, requestID, http.StatusNotFound, CodeNotFound, "Match not found.")
			return
		}
		writeRESTError(w, requestID, http.StatusInternalServerError, CodeInvalidRequest, "Could not load match history.")
		return
	}

	players := make([]persistence.CompletedMatchPlayer, len(match.Players))
	copy(players, match.Players)
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Rank != players[j].Rank {
			return players[i].Rank < players[j].Rank
		}
		return players[i].PlayerID < players[j].PlayerID
	})

	ranking := make([]MatchRankingEntry, 0, len(players))
	for _, p := range players {
		ranking = append(ranking, MatchRankingEntry{
			PlayerID:        p.PlayerID,
			DisplayName:     p.DisplayName,
			CumulativeScore: p.CumulativeScore,
			Rank:            p.Rank,
		})
	}

	roundCount := match.RoundCount
	if roundCount <= 0 {
		roundCount = 3
	}

	resp := matchHistoryResponse{
		Version:      ProtocolVersion,
		RequestID:    requestID,
		MatchID:      match.MatchID,
		RoundCount:   roundCount,
		FinalRanking: ranking,
	}
	writeRESTJSON(w, http.StatusOK, resp)
}
