package scoring

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sushi-go-game/backend/models"
)

// Feature: sushi-go-game, Property 7: Nigiri base scoring is correct
// Validates: Requirements 9.1, 9.2, 9.3
func TestProperty_NigiriBaseScoring(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Nigiri base scoring is correct", prop.ForAll(
		func(squids, salmons, eggs int) bool {
			// Generate a collection of Nigiri cards
			cards := []models.Card{}
			
			// Add Squid Nigiri cards
			for i := 0; i < squids; i++ {
				cards = append(cards, models.Card{
					Type:    models.CardTypeNigiri,
					Variant: "Squid",
				})
			}
			
			// Add Salmon Nigiri cards
			for i := 0; i < salmons; i++ {
				cards = append(cards, models.Card{
					Type:    models.CardTypeNigiri,
					Variant: "Salmon",
				})
			}
			
			// Add Egg Nigiri cards
			for i := 0; i < eggs; i++ {
				cards = append(cards, models.Card{
					Type:    models.CardTypeNigiri,
					Variant: "Egg",
				})
			}
			
			// Calculate expected score
			expectedScore := (squids * 3) + (salmons * 2) + (eggs * 1)
			
			// Calculate actual score
			actualScore := ScoreNigiri(cards)
			
			// Verify the score matches
			return actualScore == expectedScore
		},
		gen.IntRange(0, 10), // squids
		gen.IntRange(0, 10), // salmons
		gen.IntRange(0, 10), // eggs
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
