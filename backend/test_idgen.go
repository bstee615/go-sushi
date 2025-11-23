package main

import (
	"fmt"

	"github.com/sushi-go-game/backend/engine"
)

func main() {
	fmt.Println("Testing Game ID and Player Name Generation")
	fmt.Println("==========================================")
	fmt.Println()

	// Generate 10 game IDs
	fmt.Println("Generated Game IDs:")
	for i := 0; i < 10; i++ {
		gameID := engine.GenerateGameID()
		fmt.Printf("  %d. %s\n", i+1, gameID)
	}

	fmt.Println()

	// Generate 10 player names
	fmt.Println("Generated Player Names:")
	for i := 0; i < 10; i++ {
		playerName := engine.GeneratePlayerName()
		fmt.Printf("  %d. %s\n", i+1, playerName)
	}
}
