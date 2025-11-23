package engine

import (
	"regexp"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: sushi-go-game, Property 18: Game ID format is consistent
func TestGameIDFormat(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Generated game IDs match region-flower-number format", prop.ForAll(
		func(seed int) bool {
			gameID := GenerateGameID()

			// Check format: region-flower-number
			parts := strings.Split(gameID, "-")
			if len(parts) != 3 {
				t.Logf("Game ID %s does not have 3 parts", gameID)
				return false
			}

			region := parts[0]
			flower := parts[1]
			number := parts[2]

			// Check region is in the list
			validRegion := false
			for _, r := range japaneseRegions {
				if r == region {
					validRegion = true
					break
				}
			}
			if !validRegion {
				t.Logf("Region %s is not in the valid list", region)
				return false
			}

			// Check flower is in the list
			validFlower := false
			for _, f := range japaneseFlowers {
				if f == flower {
					validFlower = true
					break
				}
			}
			if !validFlower {
				t.Logf("Flower %s is not in the valid list", flower)
				return false
			}

			// Check number is 1-999
			numberRegex := regexp.MustCompile(`^[1-9]\d{0,2}$`)
			if !numberRegex.MatchString(number) {
				t.Logf("Number %s is not in range 1-999", number)
				return false
			}

			return true
		},
		gen.Int(), // Use int generator to trigger multiple test runs
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that player names are from the expected lists
func TestPlayerNameGeneration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Generated player names are from valid lists", prop.ForAll(
		func(seed int) bool {
			playerName := GeneratePlayerName()

			// Check if name is in any of the lists
			allNames := make([]string, 0)
			allNames = append(allNames, sushiChefs...)
			allNames = append(allNames, popCultureCharacters...)
			allNames = append(allNames, historicalFigures...)

			for _, name := range allNames {
				if name == playerName {
					return true
				}
			}

			t.Logf("Player name %s is not in any valid list", playerName)
			return false
		},
		gen.Int(), // Use int generator to trigger multiple test runs
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test uniqueness of generated game IDs
func TestGameIDUniqueness(t *testing.T) {
	// Generate 1000 game IDs and check for uniqueness
	generated := make(map[string]bool)
	duplicates := 0

	for i := 0; i < 1000; i++ {
		id := GenerateGameID()
		if generated[id] {
			duplicates++
		}
		generated[id] = true
	}

	// With 10 regions * 10 flowers * 999 numbers = 99,900 possible IDs,
	// we expect very few duplicates in 1000 generations
	if duplicates > 5 {
		t.Errorf("Too many duplicate IDs generated: %d out of 1000", duplicates)
	}
}
