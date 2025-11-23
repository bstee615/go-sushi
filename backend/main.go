package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sushi-go-game/backend/engine"
	"github.com/sushi-go-game/backend/handlers"
)

func main() {
	// Initialize game engine
	gameEngine := engine.NewEngine()
	
	// Initialize WebSocket handler
	wsHandler := handlers.NewWSHandler(gameEngine)
	
	// Set up routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	http.HandleFunc("/ws", wsHandler.HandleConnection)

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
