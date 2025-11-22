package scoring

import "github.com/sushi-go-game/backend/models"

// ScoringSystem handles all scoring logic for different card types
type ScoringSystem interface {
	ScoreMakiRolls(players []*models.Player) map[string]int
	ScoreTempura(cards []models.Card) int
	ScoreSashimi(cards []models.Card) int
	ScoreDumplings(cards []models.Card) int
	ScoreNigiri(cards []models.Card, wasabiCards []models.Card) int
	ScorePudding(players []*models.Player, playerCount int) map[string]int
}
