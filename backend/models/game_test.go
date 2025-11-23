package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: sushi-go-game, Property 14: Game state serialization round-trip
// Validates: Requirements 20.1, 20.2, 20.3, 20.4
func TestGameStateSerializationRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("serializing and deserializing game state preserves all data", prop.ForAll(
		func(id string, round int, phase RoundPhase, createdAt time.Time) bool {
			// Create a game with random data
			game := createTestGame(id, round, phase, createdAt)

			// Serialize the game to JSON
			jsonData, err := json.Marshal(game)
			if err != nil {
				t.Logf("Failed to marshal game: %v", err)
				return false
			}

			// Deserialize back to a Game struct
			var deserializedGame Game
			err = json.Unmarshal(jsonData, &deserializedGame)
			if err != nil {
				t.Logf("Failed to unmarshal game: %v", err)
				return false
			}

			// Verify all fields are preserved
			if !gamesEqual(&deserializedGame, game) {
				t.Logf("Games are not equal after round-trip")
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.IntRange(1, 3),
		genRoundPhase(),
		gen.Time(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// createTestGame creates a test game with the given parameters
func createTestGame(id string, round int, phase RoundPhase, createdAt time.Time) *Game {
	// Create test players
	players := []*Player{
		{
			ID:            "player1",
			Name:          "Alice",
			Hand:          []Card{{ID: "c1", Type: CardTypeMakiRoll, Value: 2}},
			Collection:    []Card{{ID: "c2", Type: CardTypeTempura}},
			PuddingCards:  []Card{{ID: "c3", Type: CardTypePudding}},
			Score:         10,
			RoundScores:   []int{5, 5},
			HasChopsticks: true,
			SelectedCard:  nil,
		},
		{
			ID:            "player2",
			Name:          "Bob",
			Hand:          []Card{{ID: "c4", Type: CardTypeNigiri, Variant: "Squid", Value: 3}},
			Collection:    []Card{{ID: "c5", Type: CardTypeWasabi}},
			PuddingCards:  []Card{},
			Score:         8,
			RoundScores:   []int{3, 5},
			HasChopsticks: false,
			SelectedCard:  intPtr(0),
		},
	}

	// Create test deck
	deck := []Card{
		{ID: "d1", Type: CardTypeSashimi},
		{ID: "d2", Type: CardTypeDumpling},
		{ID: "d3", Type: CardTypeChopsticks},
	}

	return &Game{
		ID:           id,
		Players:      players,
		Deck:         deck,
		CurrentRound: round,
		RoundPhase:   phase,
		CreatedAt:    createdAt,
	}
}

// genRoundPhase generates a random RoundPhase
func genRoundPhase() gopter.Gen {
	return gen.OneConstOf(
		PhaseWaitingForPlayers,
		PhaseSelecting,
		PhaseRevealing,
		PhasePassing,
		PhaseScoring,
		PhaseRoundEnd,
		PhaseGameEnd,
	)
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// gamesEqual checks if two games are equal
func gamesEqual(g1, g2 *Game) bool {
	if g1.ID != g2.ID {
		return false
	}
	if g1.CurrentRound != g2.CurrentRound {
		return false
	}
	if g1.RoundPhase != g2.RoundPhase {
		return false
	}
	if !g1.CreatedAt.Equal(g2.CreatedAt) {
		return false
	}
	if len(g1.Players) != len(g2.Players) {
		return false
	}
	for i := range g1.Players {
		if !playersEqual(g1.Players[i], g2.Players[i]) {
			return false
		}
	}
	if len(g1.Deck) != len(g2.Deck) {
		return false
	}
	for i := range g1.Deck {
		if !cardsEqual(g1.Deck[i], g2.Deck[i]) {
			return false
		}
	}
	return true
}

// playersEqual checks if two players are equal
func playersEqual(p1, p2 *Player) bool {
	if p1.ID != p2.ID || p1.Name != p2.Name || p1.Score != p2.Score || p1.HasChopsticks != p2.HasChopsticks {
		return false
	}
	if len(p1.Hand) != len(p2.Hand) || len(p1.Collection) != len(p2.Collection) || len(p1.PuddingCards) != len(p2.PuddingCards) {
		return false
	}
	for i := range p1.Hand {
		if !cardsEqual(p1.Hand[i], p2.Hand[i]) {
			return false
		}
	}
	for i := range p1.Collection {
		if !cardsEqual(p1.Collection[i], p2.Collection[i]) {
			return false
		}
	}
	for i := range p1.PuddingCards {
		if !cardsEqual(p1.PuddingCards[i], p2.PuddingCards[i]) {
			return false
		}
	}
	if len(p1.RoundScores) != len(p2.RoundScores) {
		return false
	}
	for i := range p1.RoundScores {
		if p1.RoundScores[i] != p2.RoundScores[i] {
			return false
		}
	}
	if (p1.SelectedCard == nil) != (p2.SelectedCard == nil) {
		return false
	}
	if p1.SelectedCard != nil && *p1.SelectedCard != *p2.SelectedCard {
		return false
	}
	return true
}

// cardsEqual checks if two cards are equal
func cardsEqual(c1, c2 Card) bool {
	return c1.ID == c2.ID && c1.Type == c2.Type && c1.Variant == c2.Variant && c1.Value == c2.Value
}
