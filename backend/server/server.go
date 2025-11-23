package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/sushi-go-game/backend/engine"
	"github.com/sushi-go-game/backend/handlers"
)

// Server represents a game server instance
type Server struct {
	engine   *engine.Engine
	handler  *handlers.WSHandler
	listener net.Listener
	server   *http.Server
	Port     int
	URL      string
}

// NewServer creates a new server that listens on the specified address
// Use ":0" for a random port, or ":8080" for a specific port
func NewServer(addr string) (*Server, error) {
	// Create listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	
	// Initialize game engine
	gameEngine := engine.NewEngine()
	
	// Initialize WebSocket handler
	wsHandler := handlers.NewWSHandler(gameEngine)
	
	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	httpServer := &http.Server{
		Handler: mux,
	}

	s := &Server{
		engine:   gameEngine,
		handler:  wsHandler,
		listener: listener,
		server:   httpServer,
		Port:     port,
		URL:      fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
	}

	return s, nil
}

// Start starts the server (blocking)
func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}

// StartBackground starts the server in a background goroutine
func (s *Server) StartBackground() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
