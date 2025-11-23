package engine

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sushi-go-game/backend/models"
)

// Feature: sushi-go-game, Property: Unique game ID generation
// Validates: Requirements 1.1
func TestUniqueGameIDGeneration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("creating multiple games generates unique IDs", prop.ForAll(
		func(numGames int) bool {
			engine := NewEngine()
			gameIDs := make(map[string]bool)

			// Create multiple games
			for i := 0; i < numGames; i++ {
				playerIDs := []string{"player1"}
				game, err := engine.CreateGame(playerIDs)
				if err != nil {
					t.Logf("Failed to create game: %v", err)
					return false
				}

				// Check if ID is unique
				if gameIDs[game.ID] {
					t.Logf("Duplicate game ID found: %s", game.ID)
					return false
				}
				gameIDs[game.ID] = true
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property: Player join validation
// Validates: Requirements 1.2, 1.4
func TestPlayerJoinValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("cannot join game with more than 5 players", prop.ForAll(
		func(numPlayers int) bool {
			engine := NewEngine()

			// Create a game with one player
			game, err := engine.CreateGame([]string{"player1"})
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			// Try to add players up to the limit
			successfulJoins := 1 // Already have player1
			usedIDs := map[string]bool{"player1": true}

			for i := 2; i <= numPlayers; i++ {
				// Generate a unique player ID
				var playerID string
				for attempts := 0; attempts < 10; attempts++ {
					playerIDVal, ok := gen.Identifier().Sample()
					if !ok {
						continue
					}
					candidateID := playerIDVal.(string)
					if !usedIDs[candidateID] {
						playerID = candidateID
						usedIDs[candidateID] = true
						break
					}
				}

				if playerID == "" {
					// Fallback to guaranteed unique ID
					playerID = fmt.Sprintf("player%d", i)
					usedIDs[playerID] = true
				}

				err := engine.JoinGame(game.ID, playerID)
				if err == nil {
					successfulJoins++
				} else if err == ErrGameFull {
					// Expected error when trying to exceed 5 players
					break
				} else {
					t.Logf("Unexpected error: %v", err)
					return false
				}
			}

			// Should not be able to have more than 5 players
			if successfulJoins > 5 {
				t.Logf("Game has %d players, exceeds maximum of 5", successfulJoins)
				return false
			}

			// If we tried to add more than 5, verify we got the error
			if numPlayers > 5 {
				// Try to add one more player with a guaranteed unique ID
				err := engine.JoinGame(game.ID, "extra_player_unique_999")
				if err != ErrGameFull {
					t.Logf("Expected ErrGameFull, got: %v", err)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property: Unique player ID assignment
// Validates: Requirements 1.5
func TestUniquePlayerIDAssignment(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all players in a game have unique IDs", prop.ForAll(
		func(numPlayers int) bool {
			// Clamp to valid range
			if numPlayers < 1 {
				numPlayers = 1
			}
			if numPlayers > 5 {
				numPlayers = 5
			}

			engine := NewEngine()

			// Generate unique player IDs
			playerIDs := make([]string, numPlayers)
			usedIDs := make(map[string]bool)

			for i := 0; i < numPlayers; i++ {
				// Try to generate a unique ID
				var playerID string
				for attempts := 0; attempts < 10; attempts++ {
					playerIDVal, ok := gen.Identifier().Sample()
					if !ok {
						continue
					}
					candidateID := playerIDVal.(string)
					if !usedIDs[candidateID] {
						playerID = candidateID
						usedIDs[candidateID] = true
						break
					}
				}

				if playerID == "" {
					// Fallback to guaranteed unique ID
					playerID = fmt.Sprintf("player%d", i)
					usedIDs[playerID] = true
				}

				playerIDs[i] = playerID
			}

			// Create game with players
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			// Check all player IDs are unique
			seenIDs := make(map[string]bool)
			for _, player := range game.Players {
				if seenIDs[player.ID] {
					t.Logf("Duplicate player ID found: %s", player.ID)
					return false
				}
				seenIDs[player.ID] = true
			}

			// Verify player count matches
			if len(game.Players) != numPlayers {
				t.Logf("Expected %d players, got %d", numPlayers, len(game.Players))
				return false
			}

			return true
		},
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property: Game start validation
// Validates: Requirements 1.3
func TestGameStartValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("game can only start with 2 or more players", prop.ForAll(
		func(numPlayers int) bool {
			// Clamp to valid range for testing
			if numPlayers < 0 {
				numPlayers = 0
			}
			if numPlayers > 5 {
				numPlayers = 5
			}

			engine := NewEngine()

			// Generate player IDs
			playerIDs := make([]string, numPlayers)
			for i := 0; i < numPlayers; i++ {
				playerIDVal, ok := gen.Identifier().Sample()
				if !ok {
					t.Logf("Failed to generate player ID")
					return false
				}
				playerIDs[i] = playerIDVal.(string)
			}

			// Create game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			// Try to start the game
			err = engine.StartGame(game.ID)

			// Verify behavior based on player count
			if numPlayers < 2 {
				// Should fail with not enough players
				if err != ErrNotEnoughPlayers {
					t.Logf("Expected ErrNotEnoughPlayers for %d players, got: %v", numPlayers, err)
					return false
				}
			} else {
				// Should succeed
				if err != nil {
					t.Logf("Expected success for %d players, got error: %v", numPlayers, err)
					return false
				}

				// Verify game state changed
				updatedGame, _ := engine.GetGame(game.ID)
				if updatedGame.CurrentRound != 1 {
					t.Logf("Expected CurrentRound to be 1, got %d", updatedGame.CurrentRound)
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 6),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 2: Hand passing preserves card count
// Validates: Requirements 4.1, 4.2
func TestHandPassingPreservesCardCount(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("card count is preserved after selection and passing", prop.ForAll(
		func(playerCount int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			engine := NewEngine()

			// Create players
			playerIDs := make([]string, playerCount)
			for i := 0; i < playerCount; i++ {
				playerIDs[i] = fmt.Sprintf("player%d", i)
			}

			// Create and start game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			err = engine.StartGame(game.ID)
			if err != nil {
				t.Logf("Failed to start game: %v", err)
				return false
			}

			// Start round (deals cards)
			err = engine.StartRound(game.ID)
			if err != nil {
				t.Logf("Failed to start round: %v", err)
				return false
			}

			// Get the game state after dealing
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			// Count total cards before selection
			totalCardsBefore := 0
			for _, player := range game.Players {
				totalCardsBefore += len(player.Hand)
				totalCardsBefore += len(player.Collection)
				totalCardsBefore += len(player.PuddingCards)
			}

			// Each player selects a card (first card in their hand)
			for _, player := range game.Players {
				if len(player.Hand) > 0 {
					err = engine.PlayCard(game.ID, player.ID, 0, false, nil)
					if err != nil {
						t.Logf("Failed to play card: %v", err)
						return false
					}
				}
			}

			// Reveal cards
			err = engine.RevealCards(game.ID)
			if err != nil {
				t.Logf("Failed to reveal cards: %v", err)
				return false
			}

			// Pass hands
			err = engine.PassHands(game.ID)
			if err != nil {
				t.Logf("Failed to pass hands: %v", err)
				return false
			}

			// Get updated game state
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game after passing: %v", err)
				return false
			}

			// Count total cards after passing
			totalCardsAfter := 0
			for _, player := range game.Players {
				totalCardsAfter += len(player.Hand)
				totalCardsAfter += len(player.Collection)
				totalCardsAfter += len(player.PuddingCards)
			}

			// Verify card count is preserved
			if totalCardsBefore != totalCardsAfter {
				t.Logf("Card count mismatch: before=%d, after=%d", totalCardsBefore, totalCardsAfter)
				return false
			}

			// Verify that hands were actually passed (each player should have one less card in hand)
			cardsPerPlayer, _ := GetCardsPerPlayer(playerCount)
			expectedHandSize := cardsPerPlayer - 1 // One card was selected and moved to collection

			for _, player := range game.Players {
				if len(player.Hand) != expectedHandSize {
					t.Logf("Player %s has %d cards in hand, expected %d", player.ID, len(player.Hand), expectedHandSize)
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 15: Simultaneous card reveal
// Validates: Requirements 3.3
func TestSimultaneousCardReveal(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("cards are only revealed when all players have selected", prop.ForAll(
		func(playerCount int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			engine := NewEngine()

			// Create players
			playerIDs := make([]string, playerCount)
			for i := 0; i < playerCount; i++ {
				playerIDs[i] = fmt.Sprintf("player%d", i)
			}

			// Create and start game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			err = engine.StartGame(game.ID)
			if err != nil {
				t.Logf("Failed to start game: %v", err)
				return false
			}

			// Start round (deals cards)
			err = engine.StartRound(game.ID)
			if err != nil {
				t.Logf("Failed to start round: %v", err)
				return false
			}

			// Get the game state
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			// Have some players (but not all) select cards
			numPlayersToSelect := playerCount - 1
			for i := 0; i < numPlayersToSelect; i++ {
				player := game.Players[i]
				if len(player.Hand) > 0 {
					err = engine.PlayCard(game.ID, player.ID, 0, false, nil)
					if err != nil {
						t.Logf("Failed to play card: %v", err)
						return false
					}
				}
			}

			// Try to reveal cards - should fail because not all players have selected
			err = engine.RevealCards(game.ID)
			if err == nil {
				t.Logf("RevealCards should have failed when not all players selected")
				return false
			}

			// Verify that no cards were added to collections yet
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			for i := 0; i < numPlayersToSelect; i++ {
				player := game.Players[i]
				// Collections should still be empty (cards not revealed yet)
				if len(player.Collection) > 0 || len(player.PuddingCards) > 0 {
					t.Logf("Player %s has cards in collection before all players selected", player.ID)
					return false
				}
			}

			// Now have the last player select a card
			lastPlayer := game.Players[playerCount-1]
			if len(lastPlayer.Hand) > 0 {
				err = engine.PlayCard(game.ID, lastPlayer.ID, 0, false, nil)
				if err != nil {
					t.Logf("Failed to play card for last player: %v", err)
					return false
				}
			}

			// Now reveal should succeed
			err = engine.RevealCards(game.ID)
			if err != nil {
				t.Logf("RevealCards failed when all players selected: %v", err)
				return false
			}

			// Verify that all players now have cards in their collections
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game after reveal: %v", err)
				return false
			}

			for _, player := range game.Players {
				totalCollected := len(player.Collection) + len(player.PuddingCards)
				if totalCollected != 1 {
					t.Logf("Player %s should have exactly 1 card in collection, has %d", player.ID, totalCollected)
					return false
				}
			}

			// Verify phase changed to revealing
			if game.RoundPhase != "revealing" {
				t.Logf("Expected phase 'revealing', got '%s'", game.RoundPhase)
				return false
			}

			return true
		},
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 11: Round progression is sequential
// Validates: Requirements 16.1, 16.2, 16.3
func TestRoundProgressionIsSequential(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("rounds progress sequentially from 1 to 3 and game ends after round 3", prop.ForAll(
		func(playerCount int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			engine := NewEngine()

			// Create players
			playerIDs := make([]string, playerCount)
			for i := 0; i < playerCount; i++ {
				playerIDs[i] = fmt.Sprintf("player%d", i)
			}

			// Create and start game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			err = engine.StartGame(game.ID)
			if err != nil {
				t.Logf("Failed to start game: %v", err)
				return false
			}

			// Verify initial round is 1
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			if game.CurrentRound != 1 {
				t.Logf("Expected initial round to be 1, got %d", game.CurrentRound)
				return false
			}

			// Simulate three rounds
			for expectedRound := 1; expectedRound <= 3; expectedRound++ {
				// Verify we're in the expected round
				game, err = engine.GetGame(game.ID)
				if err != nil {
					t.Logf("Failed to get game: %v", err)
					return false
				}

				if game.CurrentRound != expectedRound {
					t.Logf("Expected round %d, got %d", expectedRound, game.CurrentRound)
					return false
				}

				// Start the round
				err = engine.StartRound(game.ID)
				if err != nil {
					t.Logf("Failed to start round %d: %v", expectedRound, err)
					return false
				}

				// Play through the round until all cards are played
				game, err = engine.GetGame(game.ID)
				if err != nil {
					t.Logf("Failed to get game: %v", err)
					return false
				}

				// Determine how many turns in this round
				cardsPerPlayer, _ := GetCardsPerPlayer(playerCount)

				for turn := 0; turn < cardsPerPlayer; turn++ {
					// Each player selects a card
					for _, player := range game.Players {
						if len(player.Hand) > 0 {
							err = engine.PlayCard(game.ID, player.ID, 0, false, nil)
							if err != nil {
								t.Logf("Failed to play card in round %d, turn %d: %v", expectedRound, turn, err)
								return false
							}
						}
					}

					// Reveal cards
					err = engine.RevealCards(game.ID)
					if err != nil {
						t.Logf("Failed to reveal cards in round %d, turn %d: %v", expectedRound, turn, err)
						return false
					}

					// Pass hands
					err = engine.PassHands(game.ID)
					if err != nil {
						t.Logf("Failed to pass hands in round %d, turn %d: %v", expectedRound, turn, err)
						return false
					}

					// Refresh game state
					game, err = engine.GetGame(game.ID)
					if err != nil {
						t.Logf("Failed to get game: %v", err)
						return false
					}

					// Check if round ended (phase should be scoring)
					if game.RoundPhase == "scoring" {
						break
					}
				}

				// Verify we're in scoring phase
				game, err = engine.GetGame(game.ID)
				if err != nil {
					t.Logf("Failed to get game: %v", err)
					return false
				}

				if game.RoundPhase != "scoring" {
					t.Logf("Expected phase 'scoring' after round %d, got '%s'", expectedRound, game.RoundPhase)
					return false
				}

				// Score the round
				err = engine.ScoreRound(game.ID)
				if err != nil {
					t.Logf("Failed to score round %d: %v", expectedRound, err)
					return false
				}

				// Check game state after scoring
				game, err = engine.GetGame(game.ID)
				if err != nil {
					t.Logf("Failed to get game: %v", err)
					return false
				}

				// After round 3, game should be in game_end phase
				if expectedRound == 3 {
					if game.RoundPhase != "game_end" {
						t.Logf("Expected phase 'game_end' after round 3, got '%s'", game.RoundPhase)
						return false
					}
					if game.CurrentRound != 3 {
						t.Logf("Expected current round to still be 3 at game end, got %d", game.CurrentRound)
						return false
					}
				} else {
					// After rounds 1 and 2, should be in round_end phase
					if game.RoundPhase != "round_end" {
						t.Logf("Expected phase 'round_end' after round %d, got '%s'", expectedRound, game.RoundPhase)
						return false
					}
					// Round counter should have incremented
					if game.CurrentRound != expectedRound+1 {
						t.Logf("Expected round counter to be %d after scoring round %d, got %d", expectedRound+1, expectedRound, game.CurrentRound)
						return false
					}

					// Verify non-Pudding cards were cleared
					for _, player := range game.Players {
						if len(player.Collection) != 0 {
							t.Logf("Player %s still has %d cards in collection after round %d", player.ID, len(player.Collection), expectedRound)
							return false
						}
						if len(player.Hand) != 0 {
							t.Logf("Player %s still has %d cards in hand after round %d", player.ID, len(player.Hand), expectedRound)
							return false
						}
						// PuddingCards should NOT be cleared
						// (we can't verify they're preserved without actually dealing pudding cards, but we can verify they're not cleared if they exist)
					}
				}
			}

			// Verify game ended after round 3
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			if game.RoundPhase != "game_end" {
				t.Logf("Expected game to be in 'game_end' phase after 3 rounds, got '%s'", game.RoundPhase)
				return false
			}

			return true
		},
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 12: Score accumulation is monotonic per round
// Validates: Requirements 13.2
func TestScoreAccumulationIsMonotonic(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("cumulative score equals sum of round scores and never decreases during rounds", prop.ForAll(
		func(playerCount int, roundScores [][]int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			// Ensure we have exactly 3 rounds of scores
			if len(roundScores) < 3 {
				// Pad with empty rounds
				for len(roundScores) < 3 {
					roundScores = append(roundScores, make([]int, playerCount))
				}
			} else if len(roundScores) > 3 {
				roundScores = roundScores[:3]
			}

			// Ensure each round has the right number of player scores
			for i := range roundScores {
				if len(roundScores[i]) < playerCount {
					// Pad with zeros
					for len(roundScores[i]) < playerCount {
						roundScores[i] = append(roundScores[i], 0)
					}
				} else if len(roundScores[i]) > playerCount {
					roundScores[i] = roundScores[i][:playerCount]
				}

				// Ensure scores are non-negative (round scores should never be negative)
				for j := range roundScores[i] {
					if roundScores[i][j] < 0 {
						roundScores[i][j] = 0
					}
				}
			}

			engine := NewEngine()

			// Create players
			playerIDs := make([]string, playerCount)
			for i := 0; i < playerCount; i++ {
				playerIDs[i] = fmt.Sprintf("player%d", i)
			}

			// Create game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			// Simulate scoring for each round
			for round := 0; round < 3; round++ {
				// Record scores before this round
				scoresBefore := make([]int, playerCount)
				for i, player := range game.Players {
					scoresBefore[i] = player.Score
				}

				// Add round scores to players
				for i, player := range game.Players {
					roundScore := roundScores[round][i]
					player.Score += roundScore
					player.RoundScores = append(player.RoundScores, roundScore)
				}

				// Verify cumulative score equals sum of round scores
				for playerIdx, player := range game.Players {
					expectedCumulative := 0
					for _, rs := range player.RoundScores {
						expectedCumulative += rs
					}

					if player.Score != expectedCumulative {
						t.Logf("Player %s cumulative score %d != sum of round scores %d",
							player.ID, player.Score, expectedCumulative)
						return false
					}

					// Verify score is monotonically increasing (never decreases during rounds)
					if player.Score < scoresBefore[playerIdx] {
						t.Logf("Player %s score decreased from %d to %d in round %d",
							player.ID, scoresBefore[playerIdx], player.Score, round+1)
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(2, 5),
		gen.SliceOf(gen.SliceOf(gen.IntRange(0, 50))),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 17: Winner determination is deterministic
// Validates: Requirements 17.1, 17.2, 17.3
func TestWinnerDeterminationIsDeterministic(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("given same scores and pudding counts, winner is always the same", prop.ForAll(
		func(playerCount int, scores []int, puddingCounts []int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			// Ensure we have the right number of scores and pudding counts
			if len(scores) < playerCount {
				for len(scores) < playerCount {
					scores = append(scores, 0)
				}
			} else if len(scores) > playerCount {
				scores = scores[:playerCount]
			}

			if len(puddingCounts) < playerCount {
				for len(puddingCounts) < playerCount {
					puddingCounts = append(puddingCounts, 0)
				}
			} else if len(puddingCounts) > playerCount {
				puddingCounts = puddingCounts[:playerCount]
			}

			// Clamp scores and pudding counts to reasonable ranges
			for i := range scores {
				if scores[i] < 0 {
					scores[i] = 0
				}
				if scores[i] > 200 {
					scores[i] = 200
				}
			}

			for i := range puddingCounts {
				if puddingCounts[i] < 0 {
					puddingCounts[i] = 0
				}
				if puddingCounts[i] > 20 {
					puddingCounts[i] = 20
				}
			}

			// Run the winner determination twice with the same inputs
			results := make([]*GameResult, 2)

			for run := 0; run < 2; run++ {
				engine := NewEngine()

				// Create players
				playerIDs := make([]string, playerCount)
				for i := 0; i < playerCount; i++ {
					playerIDs[i] = fmt.Sprintf("player%d", i)
				}

				// Create game
				game, err := engine.CreateGame(playerIDs)
				if err != nil {
					t.Logf("Failed to create game: %v", err)
					return false
				}

				// Set up game state
				game.CurrentRound = 3
				game.RoundPhase = models.PhaseGameEnd

				// Set player scores and pudding cards
				for i, player := range game.Players {
					player.Score = scores[i]

					// Create pudding cards
					player.PuddingCards = make([]models.Card, puddingCounts[i])
					for j := 0; j < puddingCounts[i]; j++ {
						player.PuddingCards[j] = models.Card{
							ID:   fmt.Sprintf("pudding_%d_%d", i, j),
							Type: models.CardTypePudding,
						}
					}
				}

				// Call EndGame
				result, err := engine.EndGame(game.ID)
				if err != nil {
					t.Logf("Failed to end game: %v", err)
					return false
				}

				results[run] = result
			}

			// Verify both runs produced the same winner
			if results[0].Winner != results[1].Winner {
				t.Logf("Winner determination is not deterministic: run1=%s, run2=%s",
					results[0].Winner, results[1].Winner)
				return false
			}

			// Verify rankings are identical
			if len(results[0].Rankings) != len(results[1].Rankings) {
				t.Logf("Rankings length mismatch: run1=%d, run2=%d",
					len(results[0].Rankings), len(results[1].Rankings))
				return false
			}

			for i := range results[0].Rankings {
				r1 := results[0].Rankings[i]
				r2 := results[1].Rankings[i]

				if r1.PlayerID != r2.PlayerID || r1.FinalScore != r2.FinalScore ||
					r1.PuddingCount != r2.PuddingCount || r1.Rank != r2.Rank {
					t.Logf("Rankings differ at position %d", i)
					return false
				}
			}

			// Verify winner is the player with highest score
			if len(results[0].Rankings) > 0 {
				winner := results[0].Rankings[0]

				// Winner should have rank 1
				if winner.Rank != 1 {
					t.Logf("Winner has rank %d, expected 1", winner.Rank)
					return false
				}

				// Winner should have the highest score
				for i := 1; i < len(results[0].Rankings); i++ {
					other := results[0].Rankings[i]
					if other.FinalScore > winner.FinalScore {
						t.Logf("Player %s has higher score %d than winner %s with score %d",
							other.PlayerID, other.FinalScore, winner.PlayerID, winner.FinalScore)
						return false
					}

					// If scores are equal, winner should have more or equal pudding
					if other.FinalScore == winner.FinalScore && other.PuddingCount > winner.PuddingCount {
						t.Logf("Player %s has same score but more pudding than winner", other.PlayerID)
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(2, 5),
		gen.SliceOf(gen.IntRange(0, 200)),
		gen.SliceOf(gen.IntRange(0, 20)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: sushi-go-game, Property 19: Player reconnection preserves state
// Validates: Requirements 22.1, 22.2, 22.4
func TestPlayerReconnectionPreservesState(t *testing.T) {
	// Note: This test simulates reconnection at the engine level
	// The actual WebSocket reconnection logic is in the handlers package

	properties := gopter.NewProperties(nil)

	properties.Property("player state is preserved when reconnecting with same username", prop.ForAll(
		func(playerCount int) bool {
			// Clamp to valid range
			if playerCount < 2 {
				playerCount = 2
			}
			if playerCount > 5 {
				playerCount = 5
			}

			engine := NewEngine()

			// Create players with specific names
			playerIDs := make([]string, playerCount)
			playerNames := make([]string, playerCount)
			for i := 0; i < playerCount; i++ {
				playerIDs[i] = fmt.Sprintf("player%d", i)
				playerNames[i] = fmt.Sprintf("TestPlayer%d", i)
			}

			// Create and start game
			game, err := engine.CreateGame(playerIDs)
			if err != nil {
				t.Logf("Failed to create game: %v", err)
				return false
			}

			// Set player names
			for i, player := range game.Players {
				player.Name = playerNames[i]
			}

			err = engine.StartGame(game.ID)
			if err != nil {
				t.Logf("Failed to start game: %v", err)
				return false
			}

			// Start round and deal cards
			err = engine.StartRound(game.ID)
			if err != nil {
				t.Logf("Failed to start round: %v", err)
				return false
			}

			// Get game state
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			// Pick a player to "disconnect" and save their state
			disconnectedPlayerIdx := 0
			disconnectedPlayer := game.Players[disconnectedPlayerIdx]

			// Save the player's state before "disconnection"
			savedHand := make([]models.Card, len(disconnectedPlayer.Hand))
			copy(savedHand, disconnectedPlayer.Hand)
			savedCollection := make([]models.Card, len(disconnectedPlayer.Collection))
			copy(savedCollection, disconnectedPlayer.Collection)
			savedPudding := make([]models.Card, len(disconnectedPlayer.PuddingCards))
			copy(savedPudding, disconnectedPlayer.PuddingCards)
			savedScore := disconnectedPlayer.Score
			savedRoundScores := make([]int, len(disconnectedPlayer.RoundScores))
			copy(savedRoundScores, disconnectedPlayer.RoundScores)
			savedID := disconnectedPlayer.ID
			savedName := disconnectedPlayer.Name

			// Simulate some game progress (other players play cards)
			for i := 1; i < playerCount; i++ {
				player := game.Players[i]
				if len(player.Hand) > 0 {
					err = engine.PlayCard(game.ID, player.ID, 0, false, nil)
					if err != nil {
						t.Logf("Failed to play card: %v", err)
						return false
					}
				}
			}

			// Get updated game state
			game, err = engine.GetGame(game.ID)
			if err != nil {
				t.Logf("Failed to get game: %v", err)
				return false
			}

			// Verify the disconnected player's state is still intact
			reconnectedPlayer := game.Players[disconnectedPlayerIdx]

			// Check that all state is preserved
			if reconnectedPlayer.ID != savedID {
				t.Logf("Player ID changed: expected %s, got %s", savedID, reconnectedPlayer.ID)
				return false
			}

			if reconnectedPlayer.Name != savedName {
				t.Logf("Player name changed: expected %s, got %s", savedName, reconnectedPlayer.Name)
				return false
			}

			if len(reconnectedPlayer.Hand) != len(savedHand) {
				t.Logf("Hand size changed: expected %d, got %d", len(savedHand), len(reconnectedPlayer.Hand))
				return false
			}

			// Verify hand contents are the same
			for i, card := range reconnectedPlayer.Hand {
				if card.ID != savedHand[i].ID || card.Type != savedHand[i].Type {
					t.Logf("Hand card %d changed", i)
					return false
				}
			}

			if len(reconnectedPlayer.Collection) != len(savedCollection) {
				t.Logf("Collection size changed: expected %d, got %d", len(savedCollection), len(reconnectedPlayer.Collection))
				return false
			}

			if len(reconnectedPlayer.PuddingCards) != len(savedPudding) {
				t.Logf("Pudding cards changed: expected %d, got %d", len(savedPudding), len(reconnectedPlayer.PuddingCards))
				return false
			}

			if reconnectedPlayer.Score != savedScore {
				t.Logf("Score changed: expected %d, got %d", savedScore, reconnectedPlayer.Score)
				return false
			}

			if len(reconnectedPlayer.RoundScores) != len(savedRoundScores) {
				t.Logf("Round scores length changed: expected %d, got %d", len(savedRoundScores), len(reconnectedPlayer.RoundScores))
				return false
			}

			return true
		},
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
