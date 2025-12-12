package usecase

import "errors"

var (
	ErrLobbyNotFound       = errors.New("lobby not found")
	ErrLobbyCodeCollision  = errors.New("lobby code collision")
	ErrPlayerNotInLobby    = errors.New("player not in lobby")
	ErrLobbyAlreadyStarted = errors.New("lobby already started")
)
