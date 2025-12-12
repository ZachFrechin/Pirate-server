package domain

import "errors"

var (
	ErrInvalidState      = errors.New("invalid game state")
	ErrPlayerNotFound    = errors.New("player not found")
	ErrNotPlayersTurn    = errors.New("not player's turn")
	ErrInvalidHandIndex  = errors.New("invalid hand index")
	ErrInvalidCardType   = errors.New("invalid card type")
	ErrTargetNotFound    = errors.New("target player not found")
	ErrTargetInvalid     = errors.New("invalid target")
	ErrOnlyGoodCanCall   = errors.New("only good players can call over")
	ErrCannotStart       = errors.New("cannot start game")
	ErrGameFinished      = errors.New("game already finished")
	ErrPlayerEliminated  = errors.New("player is eliminated")
	ErrNotEnoughPlayers  = errors.New("not enough players")
	ErrTooManyPlayers    = errors.New("too many players")
	ErrAlreadyInGame     = errors.New("game already started")
	ErrDeckEmpty         = errors.New("draw pile is empty")
	ErrDuplicatePlayerID = errors.New("duplicate player id")
)
