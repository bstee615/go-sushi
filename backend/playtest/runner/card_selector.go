package runner

import (
	"encoding/json"
	"fmt"

	"github.com/sushi-go-game/backend/models"
)

// ResolveCardSelection resolves a card selection which can be either:
// - An integer (card index)
// - A string (card spec like "tempura" or "nigiri:squid")
func ResolveCardSelection(selection interface{}, hand []models.Card) (int, error) {
	switch v := selection.(type) {
	case int:
		// Direct index
		if v < 0 || v >= len(hand) {
			return 0, fmt.Errorf("card index %d out of range (hand size: %d)", v, len(hand))
		}
		return v, nil

	case float64:
		// JSON numbers come as float64
		index := int(v)
		if index < 0 || index >= len(hand) {
			return 0, fmt.Errorf("card index %d out of range (hand size: %d)", index, len(hand))
		}
		return index, nil

	case string:
		// Card spec - find matching card in hand
		return findCardBySpec(v, hand)

	default:
		return 0, fmt.Errorf("invalid card selection type: %T", selection)
	}
}

// findCardBySpec finds the first card in the hand matching the given spec
func findCardBySpec(spec string, hand []models.Card) (int, error) {
	// Parse the spec to get the card type and variant
	targetCard, err := ParseCardSpec(spec, "temp")
	if err != nil {
		return 0, fmt.Errorf("invalid card spec: %w", err)
	}

	// Find matching card in hand
	for i, card := range hand {
		if matchesSpec(card, targetCard) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("no card matching '%s' found in hand", spec)
}

// matchesSpec checks if a card matches the specification
func matchesSpec(card, spec models.Card) bool {
	if card.Type != spec.Type {
		return false
	}

	// For nigiri, check variant
	if card.Type == models.CardTypeNigiri {
		if spec.Variant != "" && card.Variant != spec.Variant {
			return false
		}
	}

	// For maki rolls, check value if specified
	if card.Type == models.CardTypeMakiRoll {
		if spec.Value != 0 && card.Value != spec.Value {
			return false
		}
	}

	return true
}

// ProcessSelectCardPayload processes a select_card payload and resolves card selection
func ProcessSelectCardPayload(payload map[string]interface{}, gameState map[string]interface{}) (map[string]interface{}, error) {
	// Check if cardIndex needs resolution
	cardIndexRaw, hasCardIndex := payload["cardIndex"]
	if !hasCardIndex {
		return payload, nil // No cardIndex, return as-is
	}

	// If it's already a number, no need to resolve
	if _, ok := cardIndexRaw.(float64); ok {
		return payload, nil
	}
	if _, ok := cardIndexRaw.(int); ok {
		return payload, nil
	}

	// It's a string spec, need to resolve it
	cardSpec, ok := cardIndexRaw.(string)
	if !ok {
		return nil, fmt.Errorf("cardIndex must be a number or string, got %T", cardIndexRaw)
	}

	// Get the player's hand from game state
	myHand, ok := gameState["myHand"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot resolve card spec: myHand not found in game state")
	}

	// Convert to []models.Card
	hand := make([]models.Card, len(myHand))
	for i, cardData := range myHand {
		cardJSON, err := json.Marshal(cardData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal card: %w", err)
		}
		if err := json.Unmarshal(cardJSON, &hand[i]); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card: %w", err)
		}
	}

	// Resolve the card spec to an index
	index, err := findCardBySpec(cardSpec, hand)
	if err != nil {
		return nil, err
	}

	// Update the payload with the resolved index
	result := make(map[string]interface{})
	for k, v := range payload {
		result[k] = v
	}
	result["cardIndex"] = index

	return result, nil
}
