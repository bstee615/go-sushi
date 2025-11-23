package models

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MsgTypeJoinGame     MessageType = "join_game"
	MsgTypeStartGame    MessageType = "start_game"
	MsgTypeSelectCard   MessageType = "select_card"
	MsgTypeGameState    MessageType = "game_state"
	MsgTypeCardRevealed MessageType = "card_revealed"
	MsgTypeRoundEnd     MessageType = "round_end"
	MsgTypeGameEnd      MessageType = "game_end"
	MsgTypeError        MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SelectCardPayload represents the payload for card selection
type SelectCardPayload struct {
	CardIndex       int  `json:"cardIndex"`
	UseChopsticks   bool `json:"useChopsticks"`
	SecondCardIndex *int `json:"secondCardIndex,omitempty"`
}
