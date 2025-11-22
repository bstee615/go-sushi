package engine

import "github.com/sushi-go-game/backend/models"

// GameEngine manages game state and enforces rules
type GameEngine interface {
	CreateGame(playerIDs []string) (*models.Game, error)
	StartRound(gameID string) error
	PlayCard(gameID, playerID string, cardIndex int, useChopsticks bool) error
	RevealCards(gameID string) error
	PassHands(gameID string) error
	ScoreRound(gameID string) error
	EndGame(gameID string) (*GameResult, error)
}

// GameResult represents the final result of a game
type GameResult struct {
	Winner   string
	Rankings []PlayerRanking
}

// PlayerRanking represents a player's final ranking
type PlayerRanking struct {
	PlayerID     string
	PlayerName   string
	FinalScore   int
	PuddingCount int
}
