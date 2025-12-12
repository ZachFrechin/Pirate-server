package domain

// PublicPlayerView is safe to broadcast to all clients.
type PublicPlayerView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Accusations int    `json:"accusations"`
	Eliminated  bool   `json:"eliminated"`
	HandCount   int    `json:"handCount"`
}

// SelfView contains private information for the requesting player only.
type SelfView struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
	Hand []Card `json:"hand"`
}

// GameView is the per-player state payload.
// It never reveals other players' hands or roles.
type GameView struct {
	Status GameStatus `json:"status"`
	Winner Winner     `json:"winner"`

	LobbyCode string `json:"lobbyCode"`

	ChestScore int `json:"chestScore"`
	GoalScore  int `json:"goalScore"`

	DrawCount int `json:"drawCount"`

	CurrentTurnPlayerID string             `json:"currentTurnPlayerId"`
	Players             []PublicPlayerView `json:"players"`
	You                 SelfView           `json:"you"`
}

func (g *Game) ViewFor(playerID string, lobbyCode string) (GameView, error) {
	p, err := g.mustPlayer(playerID)
	if err != nil {
		return GameView{}, err
	}
	view := GameView{
		Status:              g.Status,
		Winner:              g.Winner,
		LobbyCode:           lobbyCode,
		ChestScore:          g.ChestScore,
		GoalScore:           g.GoalScore,
		DrawCount:           len(g.DrawPile),
		CurrentTurnPlayerID: g.CurrentPlayerID(),
		Players:             make([]PublicPlayerView, 0, len(g.Players)),
		You: SelfView{
			ID:   p.ID,
			Role: p.Role,
			Hand: append([]Card(nil), p.Hand...),
		},
	}
	for _, other := range g.Players {
		if other == nil {
			continue
		}
		view.Players = append(view.Players, PublicPlayerView{
			ID:          other.ID,
			Name:        other.Name,
			Accusations: other.Accusations,
			Eliminated:  other.Eliminated,
			HandCount:   len(other.Hand),
		})
	}
	return view, nil
}
