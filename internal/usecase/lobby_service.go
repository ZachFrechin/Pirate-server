package usecase

import (
	"game-server/internal/domain"
)

// LobbyStore abstracts lobby persistence.
// This initial version uses an in-memory store.
type LobbyStore interface {
	Create(code string, lobby *Lobby) error
	Get(code string) (*Lobby, bool)
	Delete(code string)
}

// LobbyService orchestrates lobby creation/joining and game actions.
// It is transport-agnostic.
type LobbyService struct {
	store LobbyStore
}

func NewLobbyService(store LobbyStore) *LobbyService {
	return &LobbyService{store: store}
}

type CreateLobbyResult struct {
	LobbyCode string
	PlayerID  string
}

func (s *LobbyService) CreateLobby(playerName string) (CreateLobbyResult, error) {
	for i := 0; i < 10; i++ {
		code := NewLobbyCode()
		lobby := NewLobby(code)
		playerID := NewPlayerID()
		err := lobby.WithLock(func(g *domain.Game) error {
			return g.AddPlayer(&domain.Player{ID: playerID, Name: playerName})
		})
		if err != nil {
			return CreateLobbyResult{}, err
		}
		if err := s.store.Create(code, lobby); err != nil {
			if err == ErrLobbyCodeCollision {
				continue
			}
			return CreateLobbyResult{}, err
		}
		return CreateLobbyResult{LobbyCode: code, PlayerID: playerID}, nil
	}
	return CreateLobbyResult{}, ErrLobbyCodeCollision
}

type JoinLobbyResult struct {
	LobbyCode string
	PlayerID  string
}

func (s *LobbyService) JoinLobby(code string, playerName string) (JoinLobbyResult, error) {
	lobby, ok := s.store.Get(code)
	if !ok {
		return JoinLobbyResult{}, ErrLobbyNotFound
	}
	playerID := NewPlayerID()
	if err := lobby.WithLock(func(g *domain.Game) error {
		return g.AddPlayer(&domain.Player{ID: playerID, Name: playerName})
	}); err != nil {
		return JoinLobbyResult{}, err
	}
	return JoinLobbyResult{LobbyCode: code, PlayerID: playerID}, nil
}

func (s *LobbyService) StartGame(code string) error {
	lobby, ok := s.store.Get(code)
	if !ok {
		return ErrLobbyNotFound
	}
	return lobby.WithLock(func(g *domain.Game) error {
		return g.Start()
	})
}

func (s *LobbyService) PlayScore(code, playerID string, handIndex int) error {
	lobby, ok := s.store.Get(code)
	if !ok {
		return ErrLobbyNotFound
	}
	return lobby.WithLock(func(g *domain.Game) error {
		return g.PlayScoreCard(playerID, handIndex)
	})
}

func (s *LobbyService) PlayAccusation(code, playerID string, handIndex int, targetID string) error {
	lobby, ok := s.store.Get(code)
	if !ok {
		return ErrLobbyNotFound
	}
	return lobby.WithLock(func(g *domain.Game) error {
		return g.PlayAccusationCard(playerID, handIndex, targetID)
	})
}

func (s *LobbyService) CallOver(code, playerID string) error {
	lobby, ok := s.store.Get(code)
	if !ok {
		return ErrLobbyNotFound
	}
	return lobby.WithLock(func(g *domain.Game) error {
		return g.CallOver(playerID)
	})
}

func (s *LobbyService) ViewForPlayer(code, playerID string) (domain.GameView, error) {
	lobby, ok := s.store.Get(code)
	if !ok {
		return domain.GameView{}, ErrLobbyNotFound
	}
	var view domain.GameView
	err := lobby.WithLock(func(g *domain.Game) error {
		v, err := g.ViewFor(playerID, code)
		if err != nil {
			return err
		}
		view = v
		return nil
	})
	return view, err
}

func (s *LobbyService) LobbyPlayerIDs(code string) ([]string, error) {
	lobby, ok := s.store.Get(code)
	if !ok {
		return nil, ErrLobbyNotFound
	}
	ids := make([]string, 0, 8)
	_ = lobby.WithLock(func(g *domain.Game) error {
		for _, p := range g.Players {
			if p != nil {
				ids = append(ids, p.ID)
			}
		}
		return nil
	})
	return ids, nil
}
