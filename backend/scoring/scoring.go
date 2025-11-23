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

// ScoreNigiri calculates the score for Nigiri cards
// Base scoring: Squid=3, Salmon=2, Egg=1
// Wasabi cards are handled separately (not in this base implementation)
func ScoreNigiri(cards []models.Card) int {
	score := 0
	for _, card := range cards {
		if card.Type != models.CardTypeNigiri {
			continue
		}
		
		switch card.Variant {
		case "Squid":
			score += 3
		case "Salmon":
			score += 2
		case "Egg":
			score += 1
		}
	}
	return score
}
