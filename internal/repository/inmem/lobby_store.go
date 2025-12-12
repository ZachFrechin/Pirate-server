package inmem

import (
	"sync"

	"game-server/internal/usecase"
)

// LobbyStore is an in-memory implementation of usecase.LobbyStore.
// It is safe for concurrent use.
type LobbyStore struct {
	mu      sync.RWMutex
	lobbies map[string]*usecase.Lobby
}

func NewLobbyStore() *LobbyStore {
	return &LobbyStore{lobbies: make(map[string]*usecase.Lobby)}
}

func (s *LobbyStore) Create(code string, lobby *usecase.Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.lobbies[code]; exists {
		return usecase.ErrLobbyCodeCollision
	}
	s.lobbies[code] = lobby
	return nil
}

func (s *LobbyStore) Get(code string) (*usecase.Lobby, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.lobbies[code]
	return l, ok
}

func (s *LobbyStore) Delete(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lobbies, code)
}
