package engine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sushi-go-game/backend/models"
	"github.com/sushi-go-game/backend/scoring"
)

var (
	ErrGameNotFound        = errors.New("game not found")
	ErrGameFull            = errors.New("game is full (max 5 players)")
	ErrNotEnoughPlayers    = errors.New("not enough players to start (minimum 2)")
	ErrTooManyPlayers      = errors.New("too many players (maximum 5)")
	ErrPlayerAlreadyJoined = errors.New("player already in game")
)

// Engine is the concrete implementation of GameEngine
type Engine struct {
	games  map[string]*models.Game
	dealer CardDealer
	mu     sync.RWMutex
}

// NewEngine creates a new game engine with default dealer
func NewEngine() *Engine {
	return &Engine{
		games:  make(map[string]*models.Game),
		dealer: &DefaultDealer{},
	}
}

// NewEngineWithDealer creates a new game engine with a custom dealer
func NewEngineWithDealer(dealer CardDealer) *Engine {
	if dealer == nil {
		dealer = &DefaultDealer{}
	}
	return &Engine{
		games:  make(map[string]*models.Game),
		dealer: dealer,
	}
}

// CreateGame creates a new game session with unique ID generation
func (e *Engine) CreateGame(playerIDs []string) (*models.Game, error) {
	if len(playerIDs) > 5 {
		return nil, ErrTooManyPlayers
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Generate unique game ID
	gameID := e.generateUniqueGameID()

	// Create players with unique IDs
	players := make([]*models.Player, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		player := &models.Player{
			ID:            playerID,
			Name:          fmt.Sprintf("Player %s", playerID),
			Hand:          []models.Card{},
			Collection:    []models.Card{},
			PuddingCards:  []models.Card{},
			Score:         0,
			RoundScores:   []int{},
			HasChopsticks: false,
			SelectedCard:  nil,
		}
		players = append(players, player)
	}

	// Create the game
	game := &models.Game{
		ID:           gameID,
		Players:      players,
		Deck:         []models.Card{},
		CurrentRound: 0,
		RoundPhase:   models.PhaseWaitingForPlayers,
		CreatedAt:    time.Now(),
	}

	e.games[gameID] = game
	return game, nil
}

// JoinGame adds a player to an existing game
func (e *Engine) JoinGame(gameID, playerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Check if game is full
	if len(game.Players) >= 5 {
		return ErrGameFull
	}

	// Check if player already in game
	for _, p := range game.Players {
		if p.ID == playerID {
			return ErrPlayerAlreadyJoined
		}
	}

	// Add player to game
	player := &models.Player{
		ID:            playerID,
		Name:          fmt.Sprintf("Player %s", playerID),
		Hand:          []models.Card{},
		Collection:    []models.Card{},
		PuddingCards:  []models.Card{},
		Score:         0,
		RoundScores:   []int{},
		HasChopsticks: false,
		SelectedCard:  nil,
	}

	game.Players = append(game.Players, player)
	return nil
}

// StartGame starts a game if minimum player count is met
func (e *Engine) StartGame(gameID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Check minimum player count
	if len(game.Players) < 2 {
		return ErrNotEnoughPlayers
	}

	// Initialize the game
	game.CurrentRound = 1
	game.RoundPhase = models.PhaseSelecting

	return nil
}

// GetGame retrieves a game by ID
func (e *Engine) GetGame(gameID string) (*models.Game, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	game, exists := e.games[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}

	return game, nil
}

// generateUniqueGameID generates a unique game identifier
func (e *Engine) generateUniqueGameID() string {
	for {
		id := GenerateGameID()
		if _, exists := e.games[id]; !exists {
			return id
		}
	}
}

// GenerateRandomID generates a random hex string
func GenerateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// generateRandomID is a wrapper for GenerateRandomID (for backward compatibility)
func generateRandomID() string {
	return GenerateRandomID()
}

// StartRound starts a new round by dealing cards to all players
func (e *Engine) StartRound(gameID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Use the dealer to deal cards
	err := e.dealer.DealCards(game.Players, game.CurrentRound)
	if err != nil {
		return err
	}

	game.RoundPhase = models.PhaseSelecting

	// Clear selected cards
	for _, player := range game.Players {
		player.SelectedCard = nil
	}

	return nil
}

// PlayCard allows a player to select a card from their hand
func (e *Engine) PlayCard(gameID, playerID string, cardIndex int, useChopsticks bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Find the player
	var player *models.Player
	for _, p := range game.Players {
		if p.ID == playerID {
			player = p
			break
		}
	}
	if player == nil {
		return errors.New("player not found")
	}

	// Validate card index
	if cardIndex < 0 || cardIndex >= len(player.Hand) {
		return errors.New("invalid card index")
	}

	// Check if player already selected a card
	if player.SelectedCard != nil {
		return errors.New("player has already selected a card")
	}

	// If using chopsticks, validate and mark for special handling
	if useChopsticks {
		if !player.HasChopsticks {
			return errors.New("player does not have chopsticks")
		}
		// Chopsticks usage will be handled in RevealCards
		// For now, just mark that chopsticks were used
		player.HasChopsticks = false
	}

	// Store the selected card index
	player.SelectedCard = &cardIndex

	return nil
}

// WithdrawCard allows a player to withdraw their card selection
func (e *Engine) WithdrawCard(gameID, playerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Find the player
	var player *models.Player
	for _, p := range game.Players {
		if p.ID == playerID {
			player = p
			break
		}
	}
	if player == nil {
		return errors.New("player not found")
	}

	// Check if player has selected a card
	if player.SelectedCard == nil {
		return errors.New("player has not selected a card")
	}

	// Clear the selected card
	player.SelectedCard = nil

	return nil
}

// RevealCards reveals all selected cards and adds them to player collections
func (e *Engine) RevealCards(gameID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Check if all players have selected cards
	for _, player := range game.Players {
		if player.SelectedCard == nil {
			return errors.New("not all players have selected cards")
		}
	}

	// Reveal and add cards to collections
	for _, player := range game.Players {
		if player.SelectedCard != nil {
			cardIndex := *player.SelectedCard
			if cardIndex >= 0 && cardIndex < len(player.Hand) {
				selectedCard := player.Hand[cardIndex]

				// Add to appropriate collection
				if selectedCard.Type == models.CardTypePudding {
					player.PuddingCards = append(player.PuddingCards, selectedCard)
				} else {
					player.Collection = append(player.Collection, selectedCard)
				}

				// Update Chopsticks status
				if selectedCard.Type == models.CardTypeChopsticks {
					player.HasChopsticks = true
				}
			}
		}
	}

	game.RoundPhase = models.PhaseRevealing
	return nil
}

// PassHands passes each player's hand to the player on their left
func (e *Engine) PassHands(gameID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	numPlayers := len(game.Players)
	if numPlayers == 0 {
		return errors.New("no players in game")
	}

	// Remove selected cards from hands first
	for _, player := range game.Players {
		if player.SelectedCard != nil {
			cardIndex := *player.SelectedCard
			if cardIndex >= 0 && cardIndex < len(player.Hand) {
				// Remove the selected card from hand
				player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)
			}
			// Clear the selection
			player.SelectedCard = nil
		}
	}

	// Check if round is over (all hands are empty or have only 1 card left)
	roundOver := true
	for _, player := range game.Players {
		if len(player.Hand) > 1 {
			roundOver = false
			break
		}
	}

	if roundOver {
		// Play the last card if there is one
		for _, player := range game.Players {
			if len(player.Hand) == 1 {
				lastCard := player.Hand[0]
				if lastCard.Type == models.CardTypePudding {
					player.PuddingCards = append(player.PuddingCards, lastCard)
				} else {
					player.Collection = append(player.Collection, lastCard)
				}
				if lastCard.Type == models.CardTypeChopsticks {
					player.HasChopsticks = true
				}
				player.Hand = []models.Card{}
			}
		}
		game.RoundPhase = models.PhaseScoring
		return nil
	}

	// Save current hands
	savedHands := make([][]models.Card, numPlayers)
	for i, player := range game.Players {
		savedHands[i] = player.Hand
	}

	// Pass hands to the left (player i gets hand from player i-1)
	for i := 0; i < numPlayers; i++ {
		prevIndex := (i - 1 + numPlayers) % numPlayers
		game.Players[i].Hand = savedHands[prevIndex]
	}

	game.RoundPhase = models.PhaseSelecting
	return nil
}

// ScoreRound scores the current round and prepares for the next round or game end
func (e *Engine) ScoreRound(gameID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Verify we're in the scoring phase
	if game.RoundPhase != models.PhaseScoring {
		return errors.New("game is not in scoring phase")
	}

	// Calculate scores for this round
	for _, player := range game.Players {
		roundScore := scorePlayerRound(player, game.Players)
		player.Score += roundScore
		player.RoundScores = append(player.RoundScores, roundScore)
	}

	// Mark round as ended
	game.RoundPhase = models.PhaseRoundEnd

	// Check if this was the final round
	if game.CurrentRound >= 3 {
		// Game is over, trigger final scoring
		game.RoundPhase = models.PhaseGameEnd
		return nil
	}

	// Prepare for next round
	// Clear non-Pudding cards from player collections
	for _, player := range game.Players {
		player.Collection = []models.Card{}
		player.HasChopsticks = false
		player.Hand = []models.Card{}
		player.SelectedCard = nil
		// Note: PuddingCards are NOT cleared - they persist across rounds
	}

	// Increment round counter
	game.CurrentRound++

	return nil
}

// EndGame calculates final scores and determines the winner
func (e *Engine) EndGame(gameID string) (*GameResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}

	// Verify game is in end phase
	if game.RoundPhase != models.PhaseGameEnd {
		return nil, errors.New("game is not in end phase")
	}

	// Calculate Pudding scores
	puddingScores := calculatePuddingScores(game.Players)

	// Add Pudding scores to player final scores
	for _, player := range game.Players {
		if puddingScore, ok := puddingScores[player.ID]; ok {
			player.Score += puddingScore
		}
	}

	// Create rankings based on final scores
	rankings := make([]PlayerRanking, len(game.Players))
	for i, player := range game.Players {
		rankings[i] = PlayerRanking{
			PlayerID:     player.ID,
			PlayerName:   player.Name,
			FinalScore:   player.Score,
			PuddingCount: len(player.PuddingCards),
		}
	}

	// Sort rankings by score (descending), then by pudding count (descending) for tiebreaker
	sortRankings(rankings)

	// Assign rank numbers
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	// Winner is the first player in rankings
	winnerID := ""
	if len(rankings) > 0 {
		winnerID = rankings[0].PlayerID
	}

	result := &GameResult{
		Winner:   winnerID,
		Rankings: rankings,
	}

	return result, nil
}

// calculatePuddingScores calculates Pudding scores for all players
// Returns a map of player ID to Pudding score (can be positive or negative)
func calculatePuddingScores(players []*models.Player) map[string]int {
	scores := make(map[string]int)

	// Special case: 2-player games have no penalty for fewest Pudding
	if len(players) == 2 {
		// Find player with most Pudding
		maxPudding := -1
		for _, player := range players {
			puddingCount := len(player.PuddingCards)
			if puddingCount > maxPudding {
				maxPudding = puddingCount
			}
		}

		// Award 6 points to player(s) with most Pudding
		for _, player := range players {
			if len(player.PuddingCards) == maxPudding && maxPudding > 0 {
				scores[player.ID] = 6
			}
		}
		return scores
	}

	// For 3+ players: find most and fewest Pudding counts
	maxPudding := -1
	minPudding := 1000000 // Large number

	for _, player := range players {
		puddingCount := len(player.PuddingCards)
		if puddingCount > maxPudding {
			maxPudding = puddingCount
		}
		if puddingCount < minPudding {
			minPudding = puddingCount
		}
	}

	// Award 6 points to all players with most Pudding
	for _, player := range players {
		puddingCount := len(player.PuddingCards)
		if puddingCount == maxPudding {
			scores[player.ID] = 6
		}
	}

	// Deduct 6 points from all players with fewest Pudding
	for _, player := range players {
		puddingCount := len(player.PuddingCards)
		if puddingCount == minPudding {
			if existingScore, ok := scores[player.ID]; ok {
				scores[player.ID] = existingScore - 6
			} else {
				scores[player.ID] = -6
			}
		}
	}

	return scores
}

// sortRankings sorts player rankings by score (descending), then by pudding count (descending)
func sortRankings(rankings []PlayerRanking) {
	// Simple bubble sort for clarity
	n := len(rankings)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// Compare scores first
			if rankings[j].FinalScore < rankings[j+1].FinalScore {
				rankings[j], rankings[j+1] = rankings[j+1], rankings[j]
			} else if rankings[j].FinalScore == rankings[j+1].FinalScore {
				// If scores are equal, use pudding count as tiebreaker
				if rankings[j].PuddingCount < rankings[j+1].PuddingCount {
					rankings[j], rankings[j+1] = rankings[j+1], rankings[j]
				}
			}
		}
	}
}

// scorePlayerRound calculates the total score for a player's collection in a round
func scorePlayerRound(player *models.Player, allPlayers []*models.Player) int {
	return scoring.ScorePlayerRound(player, allPlayers)
}
