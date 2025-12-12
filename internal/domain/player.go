package domain

// Player is a participant in a lobby/game.
type Player struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        Role   `json:"role"`
	Hand        []Card `json:"-"` // never serialize directly (private information)
	Accusations int    `json:"accusations"`
	Eliminated  bool   `json:"eliminated"`
}

func (p *Player) Active() bool {
	return p != nil && !p.Eliminated
}
