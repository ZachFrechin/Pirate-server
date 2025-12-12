package domain

import (
	"crypto/rand"
	"math/big"
)

const (
	MinPlayers = 3
	MaxPlayers = 8

	StartingHandSize       = 3
	AccusationsToEliminate = 3
)

// Game contains all authoritative rules/state.
// It is transport-agnostic and safe to test in isolation.
type Game struct {
	Status  GameStatus
	Winner  Winner
	Players []*Player

	ChestScore int
	GoalScore  int

	DrawPile    []Card
	DiscardPile []Card

	TurnIndex int // index within Players slice; skips eliminated players
}

func NewLobbyGame() *Game {
	return &Game{Status: GameStatusLobby, Winner: WinnerNone}
}

// AddPlayer adds a player to the lobby before the game starts.
func (g *Game) AddPlayer(p *Player) error {
	if g.Status != GameStatusLobby {
		return ErrAlreadyInGame
	}
	if p == nil || p.ID == "" {
		return ErrPlayerNotFound
	}
	if len(g.Players) >= MaxPlayers {
		return ErrTooManyPlayers
	}
	for _, existing := range g.Players {
		if existing.ID == p.ID {
			return ErrDuplicatePlayerID
		}
	}
	g.Players = append(g.Players, p)
	return nil
}

// Start transitions the lobby into an active game, assigns roles, builds/shuffles the deck,
// and deals StartingHandSize cards to each player.
func (g *Game) Start() error {
	if g.Status != GameStatusLobby {
		return ErrCannotStart
	}
	if len(g.Players) < MinPlayers {
		return ErrNotEnoughPlayers
	}
	if len(g.Players) > MaxPlayers {
		return ErrTooManyPlayers
	}

	g.Status = GameStatusInGame
	g.Winner = WinnerNone
	g.ChestScore = 0
	g.GoalScore = len(g.Players) + 6
	g.TurnIndex = 0

	impostors := impostorCount(len(g.Players))
	assignRoles(g.Players, impostors)

	g.DrawPile = buildDeck(len(g.Players))
	shuffleCards(g.DrawPile)

	for _, p := range g.Players {
		p.Accusations = 0
		p.Eliminated = false
		p.Hand = p.Hand[:0]
		for i := 0; i < StartingHandSize; i++ {
			c, ok := g.drawOne()
			if !ok {
				// If deck is insufficient, end immediately.
				g.finishByOver()
				return nil
			}
			p.Hand = append(p.Hand, c)
		}
	}

	g.normalizeTurnIndex()
	return g.checkEndConditions()
}

func impostorCount(players int) int {
	if players <= 5 {
		return 1
	}
	return 2
}

func assignRoles(players []*Player, impostors int) {
	// Randomly select impostors.
	idx := make([]int, 0, len(players))
	for i := range players {
		idx = append(idx, i)
		players[i].Role = RoleGood
	}
	// Fisher-Yates partial shuffle for first impostors elements.
	for i := 0; i < len(idx); i++ {
		j := randInt(i, len(idx))
		idx[i], idx[j] = idx[j], idx[i]
	}
	for k := 0; k < impostors && k < len(idx); k++ {
		players[idx[k]].Role = RoleImpostor
	}
}

// buildDeck defines the server-side authoritative deck.
//
// The game spec does not define exact deck composition, only the scoring card values
// and the existence of accusation cards. This builder creates a balanced deck that scales
// with player count and is large enough for longer games.
func buildDeck(players int) []Card {
	// Per player:
	// - 6x +1
	// - 4x  0
	// - 3x -2
	// - 3x accusation
	// Total = 16 * players cards.
	deck := make([]Card, 0, 16*players)
	for i := 0; i < players; i++ {
		for j := 0; j < 6; j++ {
			deck = append(deck, Card{Type: CardTypeScore, Score: 1})
		}
		for j := 0; j < 4; j++ {
			deck = append(deck, Card{Type: CardTypeScore, Score: 0})
		}
		for j := 0; j < 3; j++ {
			deck = append(deck, Card{Type: CardTypeScore, Score: -2})
		}
		for j := 0; j < 3; j++ {
			deck = append(deck, Card{Type: CardTypeAccusation})
		}
	}
	return deck
}

func shuffleCards(cards []Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := randInt(0, i+1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

func randInt(minInclusive, maxExclusive int) int {
	if maxExclusive <= minInclusive {
		return minInclusive
	}
	n := big.NewInt(int64(maxExclusive - minInclusive))
	r, err := rand.Int(rand.Reader, n)
	if err != nil {
		// Very unlikely; deterministic fallback.
		return minInclusive
	}
	return minInclusive + int(r.Int64())
}

// CurrentPlayerID returns the active player's ID.
func (g *Game) CurrentPlayerID() string {
	g.normalizeTurnIndex()
	if len(g.Players) == 0 {
		return ""
	}
	p := g.Players[g.TurnIndex]
	if p == nil || p.Eliminated {
		return ""
	}
	return p.ID
}

func (g *Game) PlayScoreCard(playerID string, handIndex int) error {
	if g.Status != GameStatusInGame {
		if g.Status == GameStatusFinished {
			return ErrGameFinished
		}
		return ErrInvalidState
	}
	p, err := g.mustPlayer(playerID)
	if err != nil {
		return err
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	if g.CurrentPlayerID() != playerID {
		return ErrNotPlayersTurn
	}
	card, err := removeCard(&p.Hand, handIndex)
	if err != nil {
		return err
	}
	if !card.IsScore() {
		// Put it back to preserve hand order as much as possible.
		p.Hand = insertCard(p.Hand, handIndex, card)
		return ErrInvalidCardType
	}

	g.ChestScore += card.Score
	g.DiscardPile = append(g.DiscardPile, card)

	g.afterPlayDrawAdvance(p)
	return g.checkEndConditions()
}

func (g *Game) PlayAccusationCard(playerID string, handIndex int, targetID string) error {
	if g.Status != GameStatusInGame {
		if g.Status == GameStatusFinished {
			return ErrGameFinished
		}
		return ErrInvalidState
	}
	p, err := g.mustPlayer(playerID)
	if err != nil {
		return err
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	if g.CurrentPlayerID() != playerID {
		return ErrNotPlayersTurn
	}
	if targetID == "" || targetID == playerID {
		return ErrTargetInvalid
	}
	card, err := removeCard(&p.Hand, handIndex)
	if err != nil {
		return err
	}
	if !card.IsAccusation() {
		p.Hand = insertCard(p.Hand, handIndex, card)
		return ErrInvalidCardType
	}
	target, err := g.mustPlayer(targetID)
	if err != nil {
		// restore card
		p.Hand = insertCard(p.Hand, handIndex, card)
		return ErrTargetNotFound
	}
	if target.Eliminated {
		p.Hand = insertCard(p.Hand, handIndex, card)
		return ErrTargetInvalid
	}

	target.Accusations++
	if target.Accusations >= AccusationsToEliminate {
		target.Eliminated = true
	}

	g.DiscardPile = append(g.DiscardPile, card)

	g.afterPlayDrawAdvance(p)
	return g.checkEndConditions()
}

func (g *Game) CallOver(playerID string) error {
	if g.Status != GameStatusInGame {
		if g.Status == GameStatusFinished {
			return ErrGameFinished
		}
		return ErrInvalidState
	}
	p, err := g.mustPlayer(playerID)
	if err != nil {
		return err
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	if p.Role != RoleGood {
		return ErrOnlyGoodCanCall
	}
	g.finishByOver()
	return nil
}

func (g *Game) afterPlayDrawAdvance(current *Player) {
	// Draw 1 card. If the draw pile is empty, the game ends as "over".
	if c, ok := g.drawOne(); ok {
		current.Hand = append(current.Hand, c)
	} else {
		g.finishByOver()
		return
	}
	g.advanceTurn()
}

func (g *Game) drawOne() (Card, bool) {
	if len(g.DrawPile) == 0 {
		return Card{}, false
	}
	last := len(g.DrawPile) - 1
	c := g.DrawPile[last]
	g.DrawPile = g.DrawPile[:last]
	return c, true
}

func (g *Game) advanceTurn() {
	if len(g.Players) == 0 {
		return
	}
	g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
	g.normalizeTurnIndex()
}

func (g *Game) normalizeTurnIndex() {
	if len(g.Players) == 0 {
		g.TurnIndex = 0
		return
	}
	// Ensure TurnIndex points at an active player.
	for i := 0; i < len(g.Players); i++ {
		idx := (g.TurnIndex + i) % len(g.Players)
		p := g.Players[idx]
		if p != nil && !p.Eliminated {
			g.TurnIndex = idx
			return
		}
	}
	// All eliminated: keep 0.
	g.TurnIndex = 0
}

func (g *Game) finishByOver() {
	g.Status = GameStatusFinished
	if g.ChestScore >= g.GoalScore {
		g.Winner = WinnerGood
	} else {
		g.Winner = WinnerImpostor
	}
}

func (g *Game) checkEndConditions() error {
	if g.Status != GameStatusInGame {
		return nil
	}
	// If all impostors are eliminated => good wins.
	aliveImpostors := 0
	alivePlayers := 0
	for _, p := range g.Players {
		if p == nil || p.Eliminated {
			continue
		}
		alivePlayers++
		if p.Role == RoleImpostor {
			aliveImpostors++
		}
	}
	if alivePlayers == 0 {
		g.Status = GameStatusFinished
		g.Winner = WinnerNone
		return nil
	}
	if aliveImpostors == 0 {
		g.Status = GameStatusFinished
		g.Winner = WinnerGood
		return nil
	}
	return nil
}

func (g *Game) mustPlayer(id string) (*Player, error) {
	for _, p := range g.Players {
		if p != nil && p.ID == id {
			return p, nil
		}
	}
	return nil, ErrPlayerNotFound
}

func removeCard(hand *[]Card, index int) (Card, error) {
	if hand == nil {
		return Card{}, ErrInvalidHandIndex
	}
	h := *hand
	if index < 0 || index >= len(h) {
		return Card{}, ErrInvalidHandIndex
	}
	c := h[index]
	copy(h[index:], h[index+1:])
	h = h[:len(h)-1]
	*hand = h
	return c, nil
}

func insertCard(hand []Card, index int, c Card) []Card {
	if index < 0 {
		index = 0
	}
	if index > len(hand) {
		index = len(hand)
	}
	hand = append(hand, Card{})
	copy(hand[index+1:], hand[index:])
	hand[index] = c
	return hand
}
