package runner

import (
	"time"

	"github.com/sushi-go-game/backend/server"
)

// StartTestServer starts a new test server on a random port
func StartTestServer() (*server.Server, error) {
	// Create server on random port (":0")
	srv, err := server.NewServer(":0")
	if err != nil {
		return nil, err
	}

	// Start server in background
	srv.StartBackground()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}
