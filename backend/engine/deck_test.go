package engine

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sushi-go-game/backend/models"
)

// Feature: sushi-go-game, Property 1: Card dealing consistency
// Validates: Requirements 2.1, 2.2, 2.3, 2.4
func TestCardDealingConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("total cards dealt equals player count times cards per player", prop.ForAll(
		func(playerCount int) bool {
			// Get expected cards per player
			cardsPerPlayer, err := GetCardsPerPlayer(playerCount)
			if err != nil {
				t.Logf("Invalid player count: %d", playerCount)
				return false
			}

			// Initialize and shuffle deck
			deck := InitializeDeck()
			shuffledDeck := ShuffleDeck(deck)

			// Create players
			players := make([]*models.Player, playerCount)
			for i := 0; i < playerCount; i++ {
				players[i] = &models.Player{
					ID:   string(rune('A' + i)),
					Name: string(rune('A' + i)),
					Hand: []models.Card{},
				}
			}

			// Deal cards
			dealtPlayers, remainingDeck, err := DealCards(shuffledDeck, players)
			if err != nil {
				t.Logf("Failed to deal cards: %v", err)
				return false
			}

			// Verify each player has correct number of cards
			for i, player := range dealtPlayers {
				if len(player.Hand) != cardsPerPlayer {
					t.Logf("Player %d has %d cards, expected %d", i, len(player.Hand), cardsPerPlayer)
					return false
				}
			}

			// Verify total cards dealt
			totalCardsDealt := 0
			for _, player := range dealtPlayers {
				totalCardsDealt += len(player.Hand)
			}

			expectedTotal := playerCount * cardsPerPlayer
			if totalCardsDealt != expectedTotal {
				t.Logf("Total cards dealt: %d, expected: %d", totalCardsDealt, expectedTotal)
				return false
			}

			// Verify remaining deck size
			expectedRemaining := len(shuffledDeck) - expectedTotal
			if len(remainingDeck) != expectedRemaining {
				t.Logf("Remaining deck size: %d, expected: %d", len(remainingDeck), expectedRemaining)
				return false
			}

			// Verify no card duplication (all dealt cards are unique)
			cardIDs := make(map[string]bool)
			for _, player := range dealtPlayers {
				for _, card := range player.Hand {
					if cardIDs[card.ID] {
						t.Logf("Duplicate card ID found: %s", card.ID)
						return false
					}
					cardIDs[card.ID] = true
				}
			}

			return true
		},
		gen.IntRange(2, 5), // Valid player counts: 2-5
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that deck initialization creates correct number of cards
func TestDeckInitialization(t *testing.T) {
	deck := InitializeDeck()

	// Total cards in Sushi Go! deck: 108
	// Maki: 26, Tempura: 14, Sashimi: 14, Dumplings: 14
	// Nigiri: 20 (5 Squid + 10 Salmon + 5 Egg)
	// Wasabi: 6, Chopsticks: 4, Pudding: 10
	expectedTotal := 108
	if len(deck) != expectedTotal {
		t.Errorf("Expected %d cards in deck, got %d", expectedTotal, len(deck))
	}

	// Count cards by type
	cardCounts := make(map[models.CardType]int)
	for _, card := range deck {
		cardCounts[card.Type]++
	}

	// Verify counts
	expectedCounts := map[models.CardType]int{
		models.CardTypeMakiRoll:   26,
		models.CardTypeTempura:    14,
		models.CardTypeSashimi:    14,
		models.CardTypeDumpling:   14,
		models.CardTypeNigiri:     20,
		models.CardTypeWasabi:     6,
		models.CardTypeChopsticks: 4,
		models.CardTypePudding:    10,
	}

	for cardType, expectedCount := range expectedCounts {
		if cardCounts[cardType] != expectedCount {
			t.Errorf("Expected %d %s cards, got %d", expectedCount, cardType, cardCounts[cardType])
		}
	}
}

// Test that shuffle actually changes card order
func TestShuffleDeck(t *testing.T) {
	deck := InitializeDeck()
	shuffled := ShuffleDeck(deck)

	// Verify same length
	if len(shuffled) != len(deck) {
		t.Errorf("Shuffled deck has different length: %d vs %d", len(shuffled), len(deck))
	}

	// Verify all cards are present (check IDs)
	originalIDs := make(map[string]bool)
	for _, card := range deck {
		originalIDs[card.ID] = true
	}

	for _, card := range shuffled {
		if !originalIDs[card.ID] {
			t.Errorf("Shuffled deck contains card not in original: %s", card.ID)
		}
	}

	// Note: We can't reliably test that order changed because shuffle is random
	// and there's a small chance it could produce the same order
}

// Test GetCardsPerPlayer function
func TestGetCardsPerPlayer(t *testing.T) {
	tests := []struct {
		playerCount int
		expected    int
		shouldError bool
	}{
		{2, 10, false},
		{3, 9, false},
		{4, 8, false},
		{5, 7, false},
		{1, 0, true},
		{6, 0, true},
		{0, 0, true},
		{-1, 0, true},
	}

	for _, tt := range tests {
		result, err := GetCardsPerPlayer(tt.playerCount)
		if tt.shouldError {
			if err == nil {
				t.Errorf("Expected error for player count %d, got nil", tt.playerCount)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for player count %d: %v", tt.playerCount, err)
			}
			if result != tt.expected {
				t.Errorf("For %d players, expected %d cards, got %d", tt.playerCount, tt.expected, result)
			}
		}
	}
}
