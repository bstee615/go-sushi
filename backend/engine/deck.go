package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sushi-go-game/backend/models"
)

// InitializeDeck creates a new deck with the correct card distribution for Sushi Go!
func InitializeDeck() []models.Card {
	deck := []models.Card{}
	cardID := 0

	// Maki Rolls: 6 cards with 1 icon, 12 cards with 2 icons, 8 cards with 3 icons
	for i := 0; i < 6; i++ {
		deck = append(deck, models.Card{
			ID:    fmt.Sprintf("maki_%d", cardID),
			Type:  models.CardTypeMakiRoll,
			Value: 1,
		})
		cardID++
	}
	for i := 0; i < 12; i++ {
		deck = append(deck, models.Card{
			ID:    fmt.Sprintf("maki_%d", cardID),
			Type:  models.CardTypeMakiRoll,
			Value: 2,
		})
		cardID++
	}
	for i := 0; i < 8; i++ {
		deck = append(deck, models.Card{
			ID:    fmt.Sprintf("maki_%d", cardID),
			Type:  models.CardTypeMakiRoll,
			Value: 3,
		})
		cardID++
	}

	// Tempura: 14 cards
	for i := 0; i < 14; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("tempura_%d", i),
			Type: models.CardTypeTempura,
		})
	}

	// Sashimi: 14 cards
	for i := 0; i < 14; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("sashimi_%d", i),
			Type: models.CardTypeSashimi,
		})
	}

	// Dumplings: 14 cards
	for i := 0; i < 14; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("dumpling_%d", i),
			Type: models.CardTypeDumpling,
		})
	}

	// Nigiri: 5 Squid (3 points), 10 Salmon (2 points), 5 Egg (1 point)
	for i := 0; i < 5; i++ {
		deck = append(deck, models.Card{
			ID:      fmt.Sprintf("nigiri_squid_%d", i),
			Type:    models.CardTypeNigiri,
			Variant: "Squid",
			Value:   3,
		})
	}
	for i := 0; i < 10; i++ {
		deck = append(deck, models.Card{
			ID:      fmt.Sprintf("nigiri_salmon_%d", i),
			Type:    models.CardTypeNigiri,
			Variant: "Salmon",
			Value:   2,
		})
	}
	for i := 0; i < 5; i++ {
		deck = append(deck, models.Card{
			ID:      fmt.Sprintf("nigiri_egg_%d", i),
			Type:    models.CardTypeNigiri,
			Variant: "Egg",
			Value:   1,
		})
	}

	// Wasabi: 6 cards
	for i := 0; i < 6; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("wasabi_%d", i),
			Type: models.CardTypeWasabi,
		})
	}

	// Chopsticks: 4 cards
	for i := 0; i < 4; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("chopsticks_%d", i),
			Type: models.CardTypeChopsticks,
		})
	}

	// Pudding: 10 cards
	for i := 0; i < 10; i++ {
		deck = append(deck, models.Card{
			ID:   fmt.Sprintf("pudding_%d", i),
			Type: models.CardTypePudding,
		})
	}

	return deck
}

// ShuffleDeck shuffles the deck using Fisher-Yates algorithm
func ShuffleDeck(deck []models.Card) []models.Card {
	// Create a new random source with current time as seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Create a copy to avoid modifying the original
	shuffled := make([]models.Card, len(deck))
	copy(shuffled, deck)

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// GetCardsPerPlayer returns the number of cards each player should receive based on player count
func GetCardsPerPlayer(playerCount int) (int, error) {
	switch playerCount {
	case 2:
		return 10, nil
	case 3:
		return 9, nil
	case 4:
		return 8, nil
	case 5:
		return 7, nil
	default:
		return 0, fmt.Errorf("invalid player count: %d (must be 2-5)", playerCount)
	}
}

// DealCards deals cards to players based on player count
// Returns the updated players with cards in their hands and the remaining deck
func DealCards(deck []models.Card, players []*models.Player) ([]*models.Player, []models.Card, error) {
	playerCount := len(players)
	cardsPerPlayer, err := GetCardsPerPlayer(playerCount)
	if err != nil {
		return nil, nil, err
	}

	totalCardsNeeded := cardsPerPlayer * playerCount
	if len(deck) < totalCardsNeeded {
		return nil, nil, fmt.Errorf("not enough cards in deck: need %d, have %d", totalCardsNeeded, len(deck))
	}

	// Deal cards to each player
	cardIndex := 0
	for _, player := range players {
		player.Hand = make([]models.Card, cardsPerPlayer)
		for i := 0; i < cardsPerPlayer; i++ {
			player.Hand[i] = deck[cardIndex]
			cardIndex++
		}
	}

	// Return remaining deck
	remainingDeck := deck[cardIndex:]
	return players, remainingDeck, nil
}
