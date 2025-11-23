package runner

import (
	"fmt"

	"github.com/sushi-go-game/backend/models"
)

// PlaytestDealer deals cards according to a predefined script
type PlaytestDealer struct {
	deals map[int][][]models.Card // round -> player index -> cards
}

// NewPlaytestDealer creates a dealer with predefined card deals
func NewPlaytestDealer(deals map[int][][]models.Card) *PlaytestDealer {
	return &PlaytestDealer{
		deals: deals,
	}
}

// DealCards deals cards according to the predefined script
// Players are dealt cards in the order they appear in the players slice
func (d *PlaytestDealer) DealCards(players []*models.Player, round int, cardsPerHand int) error {
	roundDeals, ok := d.deals[round]
	if !ok {
		return fmt.Errorf("no card deals defined for round %d", round)
	}

	if len(roundDeals) != len(players) {
		return fmt.Errorf("round %d: expected %d players but have deals for %d", round, len(players), len(roundDeals))
	}

	for i, player := range players {
		player.Hand = roundDeals[i]
	}

	return nil
}

// ParseCardSpec parses a card specification string into a Card
// Format: "type" or "type:variant" or "type:variant:value"
// Examples: "tempura", "nigiri:squid", "maki_roll::2"
func ParseCardSpec(spec string, id string) (models.Card, error) {
	// Parse the spec
	parts := []string{}
	current := ""
	for _, ch := range spec {
		if ch == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	parts = append(parts, current)

	if len(parts) == 0 || parts[0] == "" {
		return models.Card{}, fmt.Errorf("invalid card spec: %s", spec)
	}

	card := models.Card{
		ID:   id,
		Type: models.CardType(parts[0]),
	}

	// Set variant if provided
	if len(parts) > 1 && parts[1] != "" {
		card.Variant = parts[1]
	}

	// Set value if provided
	if len(parts) > 2 && parts[2] != "" {
		var value int
		_, err := fmt.Sscanf(parts[2], "%d", &value)
		if err != nil {
			return models.Card{}, fmt.Errorf("invalid value in card spec %s: %w", spec, err)
		}
		card.Value = value
	}

	// Set default values based on type
	switch card.Type {
	case models.CardTypeNigiri:
		if card.Variant == "" {
			return models.Card{}, fmt.Errorf("nigiri card must specify variant (squid, salmon, or egg)")
		}
		if card.Value == 0 {
			switch card.Variant {
			case "squid", "Squid":
				card.Value = 3
			case "salmon", "Salmon":
				card.Value = 2
			case "egg", "Egg":
				card.Value = 1
			default:
				return models.Card{}, fmt.Errorf("unknown nigiri variant: %s", card.Variant)
			}
		}
	case models.CardTypeMakiRoll:
		if card.Value == 0 {
			card.Value = 1 // Default to 1 maki
		}
	}

	return card, nil
}
