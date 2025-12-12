package domain

// Role represents the hidden team for a player.
type Role string

const (
	RoleGood     Role = "good"
	RoleImpostor Role = "impostor"
)

// GameStatus tracks the lifecycle of a lobby/game.
type GameStatus string

const (
	GameStatusLobby    GameStatus = "lobby"
	GameStatusInGame   GameStatus = "in_game"
	GameStatusFinished GameStatus = "finished"
)

// Winner indicates who won when the game is finished.
type Winner string

const (
	WinnerNone     Winner = "none"
	WinnerGood     Winner = "good"
	WinnerImpostor Winner = "impostor"
)
