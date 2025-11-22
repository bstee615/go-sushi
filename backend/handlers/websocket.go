package handlers

import (
	"net/http"

	"github.com/sushi-go-game/backend/models"
)

// WebSocketHandler manages WebSocket connections and message routing
type WebSocketHandler interface {
	HandleConnection(w http.ResponseWriter, r *http.Request)
	BroadcastToGame(gameID string, message models.Message) error
	SendToPlayer(gameID, playerID string, message models.Message) error
}
