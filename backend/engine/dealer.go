package engine

import "github.com/sushi-go-game/backend/models"

// CardDealer is an interface for dealing cards to players
// This allows custom card dealing for testing purposes
type CardDealer interface {
	DealCards(players []*models.Player, round int) error
}

// DefaultDealer uses the standard shuffled deck dealing
type DefaultDealer struct{}

func (d *DefaultDealer) DealCards(players []*models.Player, round int) error {
	// Initialize and shuffle deck
	deck := InitializeDeck()
	deck = ShuffleDeck(deck)

	// Deal cards to players
	_, _, err := DealCards(deck, players)
	return err
}
