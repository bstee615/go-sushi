package models

import "time"

// CardType represents the type of a card
type CardType string

const (
	CardTypeMakiRoll   CardType = "maki_roll"
	CardTypeTempura    CardType = "tempura"
	CardTypeSashimi    CardType = "sashimi"
	CardTypeDumpling   CardType = "dumpling"
	CardTypeNigiri     CardType = "nigiri"
	CardTypeWasabi     CardType = "wasabi"
	CardTypeChopsticks CardType = "chopsticks"
	CardTypePudding    CardType = "pudding"
)

// RoundPhase represents the current phase of a round
type RoundPhase string

const (
	PhaseWaitingForPlayers RoundPhase = "waiting"
	PhaseSelecting         RoundPhase = "selecting"
	PhaseRevealing         RoundPhase = "revealing"
	PhasePassing           RoundPhase = "passing"
	PhaseScoring           RoundPhase = "scoring"
	PhaseRoundEnd          RoundPhase = "round_end"
	PhaseGameEnd           RoundPhase = "game_end"
)

// Card represents a single card in the game
type Card struct {
	ID      string   `json:"id"`
	Type    CardType `json:"type"`
	Variant string   `json:"variant,omitempty"` // For Nigiri types (Squid, Salmon, Egg)
	Value   int      `json:"value,omitempty"`   // Maki roll count or base points
}

// Player represents a player in the game
type Player struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Hand          []Card `json:"hand"`
	Collection    []Card `json:"collection"`
	PuddingCards  []Card `json:"pudding_cards"`
	Score         int    `json:"score"`
	RoundScores   []int  `json:"round_scores"`
	HasChopsticks bool   `json:"has_chopsticks"`
	SelectedCard  *int   `json:"selected_card,omitempty"`
}

// Game represents a complete game session
type Game struct {
	ID           string     `json:"id"`
	Players      []*Player  `json:"players"`
	Deck         []Card     `json:"deck"`
	CurrentRound int        `json:"current_round"`
	RoundPhase   RoundPhase `json:"round_phase"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GameState represents the state visible to clients
type GameState struct {
	GameID       string        `json:"game_id"`
	Players      []PlayerState `json:"players"`
	CurrentRound int           `json:"current_round"`
	Phase        RoundPhase    `json:"phase"`
}

// PlayerState represents a player's state visible to clients
type PlayerState struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HandSize    int    `json:"hand_size"`
	Collection  []Card `json:"collection"`
	Score       int    `json:"score"`
	HasSelected bool   `json:"has_selected"`
}
