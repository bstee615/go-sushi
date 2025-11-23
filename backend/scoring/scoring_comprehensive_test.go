package scoring

import (
	"testing"

	"github.com/sushi-go-game/backend/models"
)

// TestScoreMakiRolls_SingleWinner tests maki scoring with a clear winner
func TestScoreMakiRolls_SingleWinner(t *testing.T) {
	players := []*models.Player{
		{
			ID: "p1",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 3},
				{Type: models.CardTypeMakiRoll, Value: 2},
			},
		},
		{
			ID: "p2",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 2},
			},
		},
		{
			ID: "p3",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 1},
			},
		},
	}

	scores := ScoreMakiRolls(players)

	if scores["p1"] != 6 {
		t.Errorf("Expected p1 to score 6, got %d", scores["p1"])
	}
	if scores["p2"] != 3 {
		t.Errorf("Expected p2 to score 3, got %d", scores["p2"])
	}
	if scores["p3"] != 0 {
		t.Errorf("Expected p3 to score 0, got %d", scores["p3"])
	}
}

// TestScoreMakiRolls_TiedForFirst tests maki scoring with tied winners
func TestScoreMakiRolls_TiedForFirst(t *testing.T) {
	players := []*models.Player{
		{
			ID: "p1",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 3},
			},
		},
		{
			ID: "p2",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 3},
			},
		},
		{
			ID: "p3",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 1},
			},
		},
	}

	scores := ScoreMakiRolls(players)

	if scores["p1"] != 6 {
		t.Errorf("Expected p1 to score 6, got %d", scores["p1"])
	}
	if scores["p2"] != 6 {
		t.Errorf("Expected p2 to score 6, got %d", scores["p2"])
	}
	if scores["p3"] != 0 {
		t.Errorf("Expected p3 to score 0 (no second place when tied for first), got %d", scores["p3"])
	}
}

// TestScoreMakiRolls_TiedForSecond tests maki scoring with tied second place
func TestScoreMakiRolls_TiedForSecond(t *testing.T) {
	players := []*models.Player{
		{
			ID: "p1",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 3},
			},
		},
		{
			ID: "p2",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 2},
			},
		},
		{
			ID: "p3",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 2},
			},
		},
	}

	scores := ScoreMakiRolls(players)

	if scores["p1"] != 6 {
		t.Errorf("Expected p1 to score 6, got %d", scores["p1"])
	}
	if scores["p2"] != 3 {
		t.Errorf("Expected p2 to score 3, got %d", scores["p2"])
	}
	if scores["p3"] != 3 {
		t.Errorf("Expected p3 to score 3, got %d", scores["p3"])
	}
}

// TestScoreMakiRolls_NoMaki tests maki scoring when no one has maki
func TestScoreMakiRolls_NoMaki(t *testing.T) {
	players := []*models.Player{
		{
			ID:         "p1",
			Collection: []models.Card{},
		},
		{
			ID:         "p2",
			Collection: []models.Card{},
		},
	}

	scores := ScoreMakiRolls(players)

	if len(scores) != 0 {
		t.Errorf("Expected no scores, got %v", scores)
	}
}

// TestScoreTempura_CompleteSets tests tempura scoring with complete sets
func TestScoreTempura_CompleteSets(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"No tempura", 0, 0},
		{"One tempura", 1, 0},
		{"Two tempura", 2, 5},
		{"Three tempura", 3, 5},
		{"Four tempura", 4, 10},
		{"Five tempura", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := make([]models.Card, tt.count)
			for i := 0; i < tt.count; i++ {
				cards[i] = models.Card{Type: models.CardTypeTempura}
			}

			score := ScoreTempura(cards)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestScoreSashimi_CompleteSets tests sashimi scoring with complete sets
func TestScoreSashimi_CompleteSets(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"No sashimi", 0, 0},
		{"One sashimi", 1, 0},
		{"Two sashimi", 2, 0},
		{"Three sashimi", 3, 10},
		{"Four sashimi", 4, 10},
		{"Five sashimi", 5, 10},
		{"Six sashimi", 6, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := make([]models.Card, tt.count)
			for i := 0; i < tt.count; i++ {
				cards[i] = models.Card{Type: models.CardTypeSashimi}
			}

			score := ScoreSashimi(cards)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestScoreDumplings tests dumpling scoring
func TestScoreDumplings(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"No dumplings", 0, 0},
		{"One dumpling", 1, 1},
		{"Two dumplings", 2, 3},
		{"Three dumplings", 3, 6},
		{"Four dumplings", 4, 10},
		{"Five dumplings", 5, 15},
		{"Six dumplings", 6, 15},
		{"Ten dumplings", 10, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := make([]models.Card, tt.count)
			for i := 0; i < tt.count; i++ {
				cards[i] = models.Card{Type: models.CardTypeDumpling}
			}

			score := ScoreDumplings(cards)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestScoreNigiri_NoWasabi tests nigiri scoring without wasabi
func TestScoreNigiri_NoWasabi(t *testing.T) {
	tests := []struct {
		name     string
		variant  string
		expected int
	}{
		{"Squid", "Squid", 3},
		{"Salmon", "Salmon", 2},
		{"Egg", "Egg", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := []models.Card{
				{Type: models.CardTypeNigiri, Variant: tt.variant},
			}

			score := ScoreNigiri(cards)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestScoreNigiri_WithWasabi tests nigiri scoring with wasabi multiplier
func TestScoreNigiri_WithWasabi(t *testing.T) {
	tests := []struct {
		name     string
		cards    []models.Card
		expected int
	}{
		{
			name: "Wasabi then Squid",
			cards: []models.Card{
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeNigiri, Variant: "Squid"},
			},
			expected: 9, // 3 * 3
		},
		{
			name: "Wasabi then Salmon",
			cards: []models.Card{
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeNigiri, Variant: "Salmon"},
			},
			expected: 6, // 2 * 3
		},
		{
			name: "Wasabi then Egg",
			cards: []models.Card{
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeNigiri, Variant: "Egg"},
			},
			expected: 3, // 1 * 3
		},
		{
			name: "Two Wasabi then two Nigiri",
			cards: []models.Card{
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeNigiri, Variant: "Squid"},
				{Type: models.CardTypeNigiri, Variant: "Salmon"},
			},
			expected: 15, // 9 + 6
		},
		{
			name: "Wasabi then Nigiri then Nigiri",
			cards: []models.Card{
				{Type: models.CardTypeWasabi},
				{Type: models.CardTypeNigiri, Variant: "Squid"},
				{Type: models.CardTypeNigiri, Variant: "Salmon"},
			},
			expected: 11, // 9 + 2 (only first nigiri gets wasabi)
		},
		{
			name: "Nigiri then Wasabi (wasabi unused)",
			cards: []models.Card{
				{Type: models.CardTypeNigiri, Variant: "Squid"},
				{Type: models.CardTypeWasabi},
			},
			expected: 3, // Just the nigiri, wasabi has nothing to multiply
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ScoreNigiri(tt.cards)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestScoreNigiri_MultipleVariants tests scoring multiple nigiri types together
func TestScoreNigiri_MultipleVariants(t *testing.T) {
	cards := []models.Card{
		{Type: models.CardTypeNigiri, Variant: "Squid"},
		{Type: models.CardTypeNigiri, Variant: "Salmon"},
		{Type: models.CardTypeNigiri, Variant: "Egg"},
	}

	score := ScoreNigiri(cards)
	expected := 6 // Squid(3) + Salmon(2) + Egg(1)
	if score != expected {
		t.Errorf("Expected %d, got %d", expected, score)
	}
}

// TestScorePlayerRound tests complete round scoring
func TestScorePlayerRound(t *testing.T) {
	player := &models.Player{
		ID: "p1",
		Collection: []models.Card{
			{Type: models.CardTypeTempura},
			{Type: models.CardTypeTempura},
			{Type: models.CardTypeSashimi},
			{Type: models.CardTypeSashimi},
			{Type: models.CardTypeSashimi},
			{Type: models.CardTypeDumpling},
			{Type: models.CardTypeDumpling},
			{Type: models.CardTypeNigiri, Variant: "Squid"},
			{Type: models.CardTypeMakiRoll, Value: 3},
		},
	}

	allPlayers := []*models.Player{
		player,
		{
			ID: "p2",
			Collection: []models.Card{
				{Type: models.CardTypeMakiRoll, Value: 1},
			},
		},
	}

	score := ScorePlayerRound(player, allPlayers)
	
	// Expected: Tempura=5, Sashimi=10, Dumpling=3, Nigiri=3, Maki=6 = 27
	expected := 27
	if score != expected {
		t.Errorf("Expected %d, got %d", expected, score)
	}
}

// TestScorePlayerRound_EmptyCollection tests scoring with no cards
func TestScorePlayerRound_EmptyCollection(t *testing.T) {
	player := &models.Player{
		ID:         "p1",
		Collection: []models.Card{},
	}

	allPlayers := []*models.Player{player}

	score := ScorePlayerRound(player, allPlayers)
	if score != 0 {
		t.Errorf("Expected 0, got %d", score)
	}
}

// TestScorePlayerRound_OnlyChopsticks tests that chopsticks don't score
func TestScorePlayerRound_OnlyChopsticks(t *testing.T) {
	player := &models.Player{
		ID: "p1",
		Collection: []models.Card{
			{Type: models.CardTypeChopsticks},
			{Type: models.CardTypeChopsticks},
		},
	}

	allPlayers := []*models.Player{player}

	score := ScorePlayerRound(player, allPlayers)
	if score != 0 {
		t.Errorf("Expected 0 (chopsticks don't score), got %d", score)
	}
}
