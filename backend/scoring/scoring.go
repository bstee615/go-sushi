package scoring

import (
	"sort"

	"github.com/sushi-go-game/backend/models"
)

// ScoringSystem handles all scoring logic for different card types
type ScoringSystem interface {
	ScoreMakiRolls(players []*models.Player) map[string]int
	ScoreTempura(cards []models.Card) int
	ScoreSashimi(cards []models.Card) int
	ScoreDumplings(cards []models.Card) int
	ScoreNigiri(cards []models.Card, wasabiCards []models.Card) int
	ScorePudding(players []*models.Player, playerCount int) map[string]int
}

// ScoreMakiRolls calculates Maki Roll scores for all players
// Most maki rolls: 6 points, Second most: 3 points
func ScoreMakiRolls(players []*models.Player) map[string]int {
	scores := make(map[string]int)

	// Count maki rolls for each player
	type playerMaki struct {
		playerID string
		count    int
	}

	makiCounts := []playerMaki{}
	for _, player := range players {
		count := 0
		for _, card := range player.Collection {
			if card.Type == models.CardTypeMakiRoll {
				count += card.Value // Maki rolls have 1, 2, or 3 icons
			}
		}
		if count > 0 {
			makiCounts = append(makiCounts, playerMaki{player.ID, count})
		}
	}

	// Sort by count descending
	sort.Slice(makiCounts, func(i, j int) bool {
		return makiCounts[i].count > makiCounts[j].count
	})

	if len(makiCounts) == 0 {
		return scores
	}

	// Award 6 points to player(s) with most maki
	maxCount := makiCounts[0].count
	firstPlaceCount := 0
	for _, pm := range makiCounts {
		if pm.count == maxCount {
			scores[pm.playerID] = 6
			firstPlaceCount++
		}
	}

	// Award 3 points to player(s) with second most (if not tied for first)
	if firstPlaceCount == 1 && len(makiCounts) > 1 {
		secondMaxCount := makiCounts[1].count
		for _, pm := range makiCounts {
			if pm.count == secondMaxCount && pm.count < maxCount {
				scores[pm.playerID] = 3
			}
		}
	}

	return scores
}

// ScoreTempura calculates the score for Tempura cards
// 2 Tempura = 5 points, 1 Tempura = 0 points
func ScoreTempura(cards []models.Card) int {
	count := 0
	for _, card := range cards {
		if card.Type == models.CardTypeTempura {
			count++
		}
	}
	return (count / 2) * 5
}

// ScoreSashimi calculates the score for Sashimi cards
// 3 Sashimi = 10 points, fewer = 0 points
func ScoreSashimi(cards []models.Card) int {
	count := 0
	for _, card := range cards {
		if card.Type == models.CardTypeSashimi {
			count++
		}
	}
	return (count / 3) * 10
}

// ScoreDumplings calculates the score for Dumpling cards
// 1=1, 2=3, 3=6, 4=10, 5+=15 points
func ScoreDumplings(cards []models.Card) int {
	count := 0
	for _, card := range cards {
		if card.Type == models.CardTypeDumpling {
			count++
		}
	}

	switch {
	case count >= 5:
		return 15
	case count == 4:
		return 10
	case count == 3:
		return 6
	case count == 2:
		return 3
	case count == 1:
		return 1
	default:
		return 0
	}
}

// ScoreNigiri calculates the score for Nigiri cards with Wasabi multiplier
// Base scoring: Squid=3, Salmon=2, Egg=1
// Wasabi triples the next Nigiri played
func ScoreNigiri(cards []models.Card) int {
	score := 0
	wasabiCount := 0

	// Process cards in order (important for Wasabi)
	for _, card := range cards {
		if card.Type == models.CardTypeWasabi {
			wasabiCount++
		} else if card.Type == models.CardTypeNigiri {
			baseScore := 0
			switch card.Variant {
			case "Squid":
				baseScore = 3
			case "Salmon":
				baseScore = 2
			case "Egg":
				baseScore = 1
			}

			// Apply Wasabi multiplier if available
			if wasabiCount > 0 {
				score += baseScore * 3
				wasabiCount--
			} else {
				score += baseScore
			}
		}
	}

	return score
}

// ScorePlayerRound calculates the total score for a player's collection in a round
func ScorePlayerRound(player *models.Player, allPlayers []*models.Player) int {
	score := 0

	// Score individual card types
	score += ScoreTempura(player.Collection)
	score += ScoreSashimi(player.Collection)
	score += ScoreDumplings(player.Collection)
	score += ScoreNigiri(player.Collection)

	// Maki rolls are scored across all players
	makiScores := ScoreMakiRolls(allPlayers)
	if makiScore, ok := makiScores[player.ID]; ok {
		score += makiScore
	}

	return score
}
