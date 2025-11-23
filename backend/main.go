package main

import (
	"fmt"
	"log"

	"github.com/sushi-go-game/backend/server"
)

func main() {
	// Create server on port 8080
	srv, err := server.NewServer(":8080")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Server starting on port %d\n", srv.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Server error: ", err)
	}
}
