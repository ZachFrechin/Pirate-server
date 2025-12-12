package usecase

import (
	"sync"
	"time"

	"game-server/internal/domain"
)

// Lobby contains the state for a single lobby/game.
//
// Transport code should not mutate Lobby fields directly; use LobbyService.
type Lobby struct {
	Code      string
	CreatedAt time.Time

	mu sync.Mutex
	g  *domain.Game

	playerOrder []string // stable join order for UI purposes
}

func NewLobby(code string) *Lobby {
	return &Lobby{Code: code, CreatedAt: time.Now().UTC(), g: domain.NewLobbyGame()}
}

func (l *Lobby) WithLock(fn func(g *domain.Game) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return fn(l.g)
}

func (l *Lobby) GameUnsafe() *domain.Game {
	return l.g
}
