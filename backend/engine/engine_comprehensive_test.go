package engine

import (
	"testing"

	"github.com/sushi-go-game/backend/models"
)

// TestEngineCreateGame tests game creation
func TestEngineCreateGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, err := engine.CreateGame(playerIDs)

	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if game.ID == "" {
		t.Error("Game ID should not be empty")
	}

	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(game.Players))
	}

	if game.RoundPhase != models.PhaseWaitingForPlayers {
		t.Errorf("Expected phase to be waiting, got %s", game.RoundPhase)
	}

	if game.CurrentRound != 0 {
		t.Errorf("Expected round to be 0, got %d", game.CurrentRound)
	}
}

// TestEngineCreateGameTooManyPlayers tests that games can't have more than 5 players
func TestEngineCreateGameTooManyPlayers(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2", "p3", "p4", "p5", "p6"}
	_, err := engine.CreateGame(playerIDs)

	if err != ErrTooManyPlayers {
		t.Errorf("Expected ErrTooManyPlayers, got %v", err)
	}
}

// TestEngineJoinGame tests joining an existing game
func TestEngineJoinGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.JoinGame(game.ID, "p2")
	if err != nil {
		t.Fatalf("Failed to join game: %v", err)
	}

	game, _ = engine.GetGame(game.ID)
	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players after join, got %d", len(game.Players))
	}
}

// TestEngineJoinGameNotFound tests joining non-existent game
func TestEngineJoinGameNotFound(t *testing.T) {
	engine := NewEngine()

	err := engine.JoinGame("non-existent", "p1")
	if err != ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// TestEngineJoinGameFull tests that full games can't accept more players
func TestEngineJoinGameFull(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.JoinGame(game.ID, "p6")
	if err != ErrGameFull {
		t.Errorf("Expected ErrGameFull, got %v", err)
	}
}

// TestEngineJoinGameAlreadyJoined tests that players can't join twice
func TestEngineJoinGameAlreadyJoined(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.JoinGame(game.ID, "p1")
	if err != ErrPlayerAlreadyJoined {
		t.Errorf("Expected ErrPlayerAlreadyJoined, got %v", err)
	}
}

// TestEngineStartGame tests starting a game
func TestEngineStartGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.StartGame(game.ID)
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	game, _ = engine.GetGame(game.ID)
	if game.CurrentRound != 1 {
		t.Errorf("Expected round to be 1, got %d", game.CurrentRound)
	}

	if game.RoundPhase != models.PhaseSelecting {
		t.Errorf("Expected phase to be selecting, got %s", game.RoundPhase)
	}
}

// TestEngineStartGameNotEnoughPlayers tests starting with insufficient players
func TestEngineStartGameNotEnoughPlayers(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.StartGame(game.ID)
	if err != ErrNotEnoughPlayers {
		t.Errorf("Expected ErrNotEnoughPlayers, got %v", err)
	}
}

// TestEngineGetGame tests retrieving a game
func TestEngineGetGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	createdGame, _ := engine.CreateGame(playerIDs)

	retrievedGame, err := engine.GetGame(createdGame.ID)
	if err != nil {
		t.Fatalf("Failed to get game: %v", err)
	}

	if retrievedGame.ID != createdGame.ID {
		t.Errorf("Expected game ID %s, got %s", createdGame.ID, retrievedGame.ID)
	}
}

// TestEngineGetGameNotFound tests retrieving non-existent game
func TestEngineGetGameNotFound(t *testing.T) {
	engine := NewEngine()

	_, err := engine.GetGame("non-existent")
	if err != ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// TestEngineListGames tests listing all games
func TestEngineListGames(t *testing.T) {
	engine := NewEngine()

	// Create multiple games
	engine.CreateGame([]string{"p1", "p2"})
	engine.CreateGame([]string{"p3", "p4"})
	engine.CreateGame([]string{"p5"})

	games := engine.ListGames()
	if len(games) != 3 {
		t.Errorf("Expected 3 games, got %d", len(games))
	}

	// Verify each game has expected fields
	for _, game := range games {
		if game["id"] == nil || game["id"] == "" {
			t.Error("Game should have an id")
		}
		if game["playerCount"] == nil {
			t.Error("Game should have playerCount")
		}
		if game["phase"] == nil {
			t.Error("Game should have phase")
		}
	}
}

// TestEngineListGamesEmpty tests listing when no games exist
func TestEngineListGamesEmpty(t *testing.T) {
	engine := NewEngine()

	games := engine.ListGames()
	if len(games) != 0 {
		t.Errorf("Expected 0 games, got %d", len(games))
	}
}

// TestEngineDeleteGame tests deleting a game
func TestEngineDeleteGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.DeleteGame(game.ID)
	if err != nil {
		t.Fatalf("Failed to delete game: %v", err)
	}

	_, err = engine.GetGame(game.ID)
	if err != ErrGameNotFound {
		t.Errorf("Expected game to be deleted, but it still exists")
	}
}

// TestEngineDeleteGameNotFound tests deleting non-existent game
func TestEngineDeleteGameNotFound(t *testing.T) {
	engine := NewEngine()

	err := engine.DeleteGame("non-existent")
	if err != ErrGameNotFound {
		t.Errorf("Expected ErrGameNotFound, got %v", err)
	}
}

// TestEngineRemovePlayer tests removing a player from a game
func TestEngineRemovePlayer(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)

	err := engine.RemovePlayer(game.ID, "p2")
	if err != nil {
		t.Fatalf("Failed to remove player: %v", err)
	}

	game, _ = engine.GetGame(game.ID)
	if len(game.Players) != 1 {
		t.Errorf("Expected 1 player after removal, got %d", len(game.Players))
	}

	if game.Players[0].ID != "p1" {
		t.Errorf("Expected remaining player to be p1, got %s", game.Players[0].ID)
	}
}

// TestEngineRemovePlayerAfterStart tests that players can't be removed after game starts
func TestEngineRemovePlayerAfterStart(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)

	err := engine.RemovePlayer(game.ID, "p2")
	if err == nil {
		t.Error("Expected error when removing player after game start")
	}
}

// TestEngineStartRound tests starting a round with card dealing
func TestEngineStartRound(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)

	err := engine.StartRound(game.ID)
	if err != nil {
		t.Fatalf("Failed to start round: %v", err)
	}

	game, _ = engine.GetGame(game.ID)

	// Verify all players have cards
	for _, player := range game.Players {
		if len(player.Hand) == 0 {
			t.Errorf("Player %s has no cards after starting round", player.ID)
		}
	}

	if game.RoundPhase != models.PhaseSelecting {
		t.Errorf("Expected phase to be selecting, got %s", game.RoundPhase)
	}
}

// TestEngineUniqueGameIDs tests that generated game IDs are unique
func TestEngineUniqueGameIDs(t *testing.T) {
	engine := NewEngine()

	gameIDs := make(map[string]bool)
	numGames := 100

	for i := 0; i < numGames; i++ {
		game, _ := engine.CreateGame([]string{"p1"})
		if gameIDs[game.ID] {
			t.Errorf("Duplicate game ID generated: %s", game.ID)
		}
		gameIDs[game.ID] = true
	}

	if len(gameIDs) != numGames {
		t.Errorf("Expected %d unique IDs, got %d", numGames, len(gameIDs))
	}
}

// TestEngineWithCustomConfig tests creating engine with custom configuration
func TestEngineWithCustomConfig(t *testing.T) {
	engine := NewEngineWithConfig(nil, 2, 7)

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)

	if game.NumRounds != 2 {
		t.Errorf("Expected 2 rounds, got %d", game.NumRounds)
	}

	if game.CardsPerHand != 7 {
		t.Errorf("Expected 7 cards per hand, got %d", game.CardsPerHand)
	}
}

// TestEngineConcurrentAccess tests concurrent access to engine
func TestEngineConcurrentAccess(t *testing.T) {
	engine := NewEngine()

	// Create game
	game, _ := engine.CreateGame([]string{"p1", "p2"})

	// Concurrently try to join and read
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			engine.GetGame(game.ID)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			engine.ListGames()
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic or deadlock
}

// TestEnginePlayCard tests playing a card
func TestEnginePlayCard(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)
	engine.StartRound(game.ID)

	game, _ = engine.GetGame(game.ID)

	// Play first card from p1's hand
	err := engine.PlayCard(game.ID, "p1", 0, false, nil)
	if err != nil {
		t.Fatalf("Failed to play card: %v", err)
	}

	game, _ = engine.GetGame(game.ID)
	player1 := game.Players[0]

	if player1.SelectedCard == nil {
		t.Error("Player should have a selected card")
	}

	if *player1.SelectedCard != 0 {
		t.Errorf("Expected selected card index to be 0, got %d", *player1.SelectedCard)
	}
}

// TestEngineWithdrawCard tests withdrawing a selected card
func TestEngineWithdrawCard(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)
	engine.StartRound(game.ID)

	// Play a card
	engine.PlayCard(game.ID, "p1", 0, false, nil)

	// Withdraw it
	err := engine.WithdrawCard(game.ID, "p1")
	if err != nil {
		t.Fatalf("Failed to withdraw card: %v", err)
	}

	game, _ = engine.GetGame(game.ID)
	player1 := game.Players[0]

	if player1.SelectedCard != nil {
		t.Error("Player should not have a selected card after withdrawal")
	}
}

// TestEngineRevealCards tests revealing selected cards
func TestEngineRevealCards(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)
	engine.StartRound(game.ID)

	// Both players select cards
	engine.PlayCard(game.ID, "p1", 0, false, nil)
	engine.PlayCard(game.ID, "p2", 0, false, nil)

	// Reveal cards
	err := engine.RevealCards(game.ID)
	if err != nil {
		t.Fatalf("Failed to reveal cards: %v", err)
	}

	game, _ = engine.GetGame(game.ID)

	// Verify cards were moved to collection
	for _, player := range game.Players {
		if len(player.Collection) == 0 {
			t.Errorf("Player %s should have cards in collection after reveal", player.ID)
		}
	}
}

// TestEnginePassHands tests passing hands between players
func TestEnginePassHands(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2", "p3"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)
	engine.StartRound(game.ID)

	game, _ = engine.GetGame(game.ID)
	originalHands := make(map[string][]models.Card)
	for _, player := range game.Players {
		// Copy the hand
		hand := make([]models.Card, len(player.Hand))
		copy(hand, player.Hand)
		originalHands[player.ID] = hand
	}

	// Pass hands
	err := engine.PassHands(game.ID)
	if err != nil {
		t.Fatalf("Failed to pass hands: %v", err)
	}

	game, _ = engine.GetGame(game.ID)

	// Verify hands were passed (each player should have a different hand)
	for _, player := range game.Players {
		originalHand := originalHands[player.ID]
		if len(player.Hand) == len(originalHand) {
			// Check if any cards are different (hands should have changed)
			allSame := true
			for i := range player.Hand {
				if player.Hand[i].ID != originalHand[i].ID {
					allSame = false
					break
				}
			}
			if allSame && len(originalHand) > 0 {
				t.Errorf("Player %s has the same hand after passing", player.ID)
			}
		}
	}
}

// TestEngineScoreRound tests scoring a round
func TestEngineScoreRound(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)
	engine.StartRound(game.ID)

	// Give players some cards in their collection
	game, _ = engine.GetGame(game.ID)
	game.Players[0].Collection = []models.Card{
		{Type: models.CardTypeTempura},
		{Type: models.CardTypeTempura},
	}
	
	// Set game to scoring phase
	game.RoundPhase = models.PhaseScoring

	err := engine.ScoreRound(game.ID)
	if err != nil {
		t.Fatalf("Failed to score round: %v", err)
	}

	game, _ = engine.GetGame(game.ID)

	// Verify scores were updated
	if game.Players[0].Score == 0 {
		t.Error("Player 1 should have a score after scoring round")
	}

	if len(game.Players[0].RoundScores) == 0 {
		t.Error("Player 1 should have round scores recorded")
	}
}

// TestEngineEndGame tests ending a game
func TestEngineEndGame(t *testing.T) {
	engine := NewEngine()

	playerIDs := []string{"p1", "p2"}
	game, _ := engine.CreateGame(playerIDs)
	engine.StartGame(game.ID)

	// Give players some pudding cards for final scoring
	game, _ = engine.GetGame(game.ID)
	game.Players[0].PuddingCards = []models.Card{
		{Type: models.CardTypePudding},
		{Type: models.CardTypePudding},
	}
	game.Players[0].Score = 20
	game.Players[0].Name = "Player 1"

	game.Players[1].PuddingCards = []models.Card{
		{Type: models.CardTypePudding},
	}
	game.Players[1].Score = 15
	game.Players[1].Name = "Player 2"
	
	// Set game to game_end phase
	game.CurrentRound = 3
	game.RoundPhase = models.PhaseGameEnd

	result, err := engine.EndGame(game.ID)
	if err != nil {
		t.Fatalf("Failed to end game: %v", err)
	}

	if result.Winner == "" {
		t.Error("Game should have a winner")
	}

	if len(result.Rankings) != 2 {
		t.Errorf("Expected 2 rankings, got %d", len(result.Rankings))
	}

	// Winner should be player with highest score (after pudding scoring)
	game, _ = engine.GetGame(game.ID)
	if game.RoundPhase != models.PhaseGameEnd {
		t.Errorf("Expected phase to be game_end, got %s", game.RoundPhase)
	}
}
