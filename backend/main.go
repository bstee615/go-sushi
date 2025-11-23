package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sushi-go-game/backend/server"
)

func main() {
	// Command-line flags for game configuration
	numRounds := flag.Int("rounds", 3, "Number of rounds per game (default: 3)")
	cardsPerHand := flag.Int("cards", 10, "Number of cards dealt per hand (default: 10)")
	port := flag.String("port", ":8080", "Server port (default: :8080)")
	flag.Parse()

	// Create server with configuration
	options := &server.ServerOptions{
		GameConfig: &server.GameConfig{
			NumRounds:    *numRounds,
			CardsPerHand: *cardsPerHand,
		},
	}

	srv, err := server.NewServer(*port, options)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Server starting on port %d\n", srv.Port)
	fmt.Printf("Game configuration: %d rounds, %d cards per hand\n", *numRounds, *cardsPerHand)
	if err := srv.Start(); err != nil {
		log.Fatal("Server error: ", err)
	}
}
