package runner

import (
	"fmt"
	"time"

	"github.com/sushi-go-game/backend/models"
	"github.com/sushi-go-game/backend/server"
)

// StartTestServer starts a new test server on a random port
func StartTestServer(dealsSpec map[int]map[string][]string) (*server.Server, error) {
	var options *server.ServerOptions

	// If custom deals are specified, create a custom dealer
	if len(dealsSpec) > 0 {
		dealer, err := createCustomDealer(dealsSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom dealer: %w", err)
		}
		options = &server.ServerOptions{
			CustomDealer: dealer,
		}
	}

	// Create server on random port (":0")
	srv, err := server.NewServer(":0", options)
	if err != nil {
		return nil, err
	}

	// Start server in background
	srv.StartBackground()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}

// createCustomDealer creates a PlaytestDealer from the YAML deals specification
// Clients are mapped to player indices in alphabetical order
func createCustomDealer(dealsSpec map[int]map[string][]string) (*PlaytestDealer, error) {
	deals := make(map[int][][]models.Card)

	// Get sorted list of client names for consistent ordering
	var clientNames []string
	for _, roundDeals := range dealsSpec {
		for clientName := range roundDeals {
			found := false
			for _, name := range clientNames {
				if name == clientName {
					found = true
					break
				}
			}
			if !found {
				clientNames = append(clientNames, clientName)
			}
		}
	}

	// Sort client names alphabetically for consistent player ordering
	for i := 0; i < len(clientNames); i++ {
		for j := i + 1; j < len(clientNames); j++ {
			if clientNames[i] > clientNames[j] {
				clientNames[i], clientNames[j] = clientNames[j], clientNames[i]
			}
		}
	}

	for round, roundDeals := range dealsSpec {
		playerCards := make([][]models.Card, len(clientNames))

		for i, clientName := range clientNames {
			cardSpecs, ok := roundDeals[clientName]
			if !ok {
				return nil, fmt.Errorf("round %d: no cards specified for client %s", round, clientName)
			}

			cards := make([]models.Card, len(cardSpecs))
			for j, spec := range cardSpecs {
				card, err := ParseCardSpec(spec, fmt.Sprintf("%s_r%d_c%d", clientName, round, j))
				if err != nil {
					return nil, fmt.Errorf("failed to parse card spec '%s' for client %s: %w", spec, clientName, err)
				}
				cards[j] = card
			}

			playerCards[i] = cards
		}

		deals[round] = playerCards
	}

	return NewPlaytestDealer(deals), nil
}
