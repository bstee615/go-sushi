package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sushi-go-game/backend/engine"
	"github.com/sushi-go-game/backend/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Client represents a connected WebSocket client
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	gameID   string
	playerID string
}

// WSHandler implements WebSocketHandler interface
type WSHandler struct {
	engine         *engine.Engine
	clients        map[string]*Client            // playerID -> Client
	games          map[string]map[string]*Client // gameID -> playerID -> Client
	allConnections map[*Client]bool              // All connected clients (including those not in games)
	mu             sync.RWMutex
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(engine *engine.Engine) *WSHandler {
	return &WSHandler{
		engine:         engine,
		clients:        make(map[string]*Client),
		games:          make(map[string]map[string]*Client),
		allConnections: make(map[*Client]bool),
	}
}

// HandleConnection handles new WebSocket connections
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Register this connection
	h.mu.Lock()
	h.allConnections[client] = true
	h.mu.Unlock()

	log.Printf("New client connected, total connections: %d", len(h.allConnections))

	// Start goroutines for reading and writing
	go h.readPump(client)
	go h.writePump(client)
}

// readPump reads messages from the WebSocket connection
func (h *WSHandler) readPump(client *Client) {
	defer func() {
		h.removeClient(client)
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.handleMessage(client, message)
	}
}

// writePump writes messages to the WebSocket connection
func (h *WSHandler) writePump(client *Client) {
	defer client.conn.Close()

	for message := range client.send {
		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to write message: %v", err)
			return
		}
	}
}

// handleMessage processes incoming messages from clients
func (h *WSHandler) handleMessage(client *Client, message []byte) {
	log.Printf("Received message from client %s: %s", client.playerID, string(message))

	var msg models.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		h.sendError(client, "Invalid message format")
		return
	}

	log.Printf("Message type: %s, PlayerID: %s, GameID: %s", msg.Type, client.playerID, client.gameID)

	switch msg.Type {
	case models.MsgTypeJoinGame:
		h.handleJoinGame(client, msg.Payload)
	case models.MsgTypeStartGame:
		h.handleStartGame(client, msg.Payload)
	case models.MsgTypeSelectCard:
		log.Printf("Handling select_card for player %s in game %s", client.playerID, client.gameID)
		h.handleSelectCard(client, msg.Payload)
	case models.MsgTypeWithdrawCard:
		log.Printf("Handling withdraw_card for player %s in game %s", client.playerID, client.gameID)
		h.handleWithdrawCard(client, msg.Payload)
	case models.MsgTypeKickPlayer:
		log.Printf("Handling kick_player for player %s in game %s", client.playerID, client.gameID)
		h.handleKickPlayer(client, msg.Payload)
	case models.MsgTypeListGames:
		log.Printf("Handling list_games")
		h.handleListGames(client)
	case models.MsgTypeDeleteGame:
		log.Printf("Handling delete_game")
		h.handleDeleteGame(client, msg.Payload)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		h.sendError(client, "Unknown message type")
	}
}

// handleJoinGame handles join_game messages
func (h *WSHandler) handleJoinGame(client *Client, payload json.RawMessage) {
	var data struct {
		GameID     string `json:"gameId"`
		PlayerName string `json:"playerName"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		h.sendError(client, "Invalid join_game payload")
		return
	}

	// Generate random player name if not provided
	playerName := data.PlayerName
	if playerName == "" {
		playerName = engine.GeneratePlayerName()
	}

	var game *models.Game
	var err error
	var playerID string
	var isReconnection bool

	if data.GameID == "" {
		// Create new game
		playerID = engine.GenerateRandomID()
		game, err = h.engine.CreateGame([]string{playerID})
		if err != nil {
			h.sendError(client, "Failed to create game: "+err.Error())
			return
		}

		// Set player name
		if len(game.Players) > 0 {
			game.Players[0].Name = playerName
		}
	} else {
		// Try to join existing game
		game, err = h.engine.GetGame(data.GameID)
		if err != nil {
			h.sendError(client, "Failed to get game: "+err.Error())
			return
		}

		// Check if player with this name already exists (reconnection)
		var existingPlayer *models.Player
		for _, player := range game.Players {
			if player.Name == playerName {
				existingPlayer = player
				break
			}
		}

		if existingPlayer != nil {
			// Reconnection: use existing player ID
			playerID = existingPlayer.ID
			isReconnection = true
			log.Printf("Player %s reconnecting to game %s", playerName, data.GameID)
		} else {
			// New player joining
			playerID = engine.GenerateRandomID()
			err = h.engine.JoinGame(data.GameID, playerID)
			if err != nil {
				h.sendError(client, "Failed to join game: "+err.Error())
				return
			}

			// Refresh game state
			game, err = h.engine.GetGame(data.GameID)
			if err != nil {
				h.sendError(client, "Failed to get game: "+err.Error())
				return
			}

			// Set player name
			for _, player := range game.Players {
				if player.ID == playerID {
					player.Name = playerName
					break
				}
			}
		}
	}

	client.playerID = playerID
	client.gameID = game.ID

	// Register client
	h.mu.Lock()
	// Remove old client connection if reconnecting
	if isReconnection {
		if oldClient, exists := h.clients[playerID]; exists {
			close(oldClient.send)
		}
	}
	h.clients[playerID] = client
	if h.games[game.ID] == nil {
		h.games[game.ID] = make(map[string]*Client)
	}
	h.games[game.ID][playerID] = client
	gameWasCreated := data.GameID == ""
	h.mu.Unlock()

	// Broadcast updated game state to all players in the game
	h.broadcastGameState(game.ID)

	// If a new game was created, broadcast updated games list to all clients
	if gameWasCreated {
		h.broadcastGamesList()
	}

	if isReconnection {
		log.Printf("Player %s successfully reconnected to game %s", playerName, game.ID)
	}
}

// handleStartGame handles start_game messages
func (h *WSHandler) handleStartGame(client *Client, payload json.RawMessage) {
	var data struct {
		GameID string `json:"gameId"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		h.sendError(client, "Invalid start_game payload")
		return
	}

	// Start the game
	if err := h.engine.StartGame(data.GameID); err != nil {
		h.sendError(client, "Failed to start game: "+err.Error())
		return
	}

	// Start the first round
	if err := h.engine.StartRound(data.GameID); err != nil {
		h.sendError(client, "Failed to start round: "+err.Error())
		return
	}

	// Broadcast updated game state
	h.broadcastGameState(data.GameID)
}

// handleSelectCard handles select_card messages
func (h *WSHandler) handleSelectCard(client *Client, payload json.RawMessage) {
	log.Printf("handleSelectCard: Starting for player %s", client.playerID)
	var data models.SelectCardPayload

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("handleSelectCard: Failed to unmarshal payload: %v", err)
		h.sendError(client, "Invalid select_card payload")
		return
	}

	log.Printf("handleSelectCard: Player %s selecting card index %d", client.playerID, data.CardIndex)

	// Play the card
	if err := h.engine.PlayCard(client.gameID, client.playerID, data.CardIndex, data.UseChopsticks); err != nil {
		log.Printf("handleSelectCard: PlayCard failed: %v", err)
		h.sendError(client, "Failed to select card: "+err.Error())
		return
	}

	log.Printf("handleSelectCard: Card selected successfully, broadcasting state")

	// Broadcast updated game state (without revealing the card)
	h.broadcastGameState(client.gameID)

	// Check if all players have selected cards
	game, err := h.engine.GetGame(client.gameID)
	if err != nil {
		log.Printf("handleSelectCard: Failed to get game: %v", err)
		return
	}

	allSelected := true
	for _, player := range game.Players {
		if player.SelectedCard == nil {
			allSelected = false
			break
		}
	}

	log.Printf("handleSelectCard: All players selected? %v", allSelected)

	if allSelected {
		log.Printf("handleSelectCard: All players selected, revealing cards")
		// Reveal cards
		if err := h.engine.RevealCards(client.gameID); err != nil {
			log.Printf("Failed to reveal cards: %v", err)
			return
		}

		log.Printf("handleSelectCard: Cards revealed, broadcasting")
		// Broadcast card reveal
		h.broadcastGameState(client.gameID)

		log.Printf("handleSelectCard: Passing hands")
		// Pass hands
		if err := h.engine.PassHands(client.gameID); err != nil {
			log.Printf("Failed to pass hands: %v", err)
			return
		}

		log.Printf("handleSelectCard: Hands passed, getting updated game state")
		// Get updated game state
		game, err = h.engine.GetGame(client.gameID)
		if err != nil {
			log.Printf("handleSelectCard: Failed to get game after passing: %v", err)
			return
		}

		// Check if round ended
		if game.RoundPhase == models.PhaseScoring {
			// Score the round
			if err := h.engine.ScoreRound(client.gameID); err != nil {
				log.Printf("Failed to score round: %v", err)
				return
			}

			// Get updated game state
			game, err = h.engine.GetGame(client.gameID)
			if err != nil {
				return
			}

			// Check if game ended
			if game.RoundPhase == models.PhaseGameEnd {
				// Calculate final results
				result, err := h.engine.EndGame(client.gameID)
				if err != nil {
					log.Printf("Failed to end game: %v", err)
					return
				}

				// Broadcast game end
				h.broadcastGameEnd(client.gameID, result)
			} else {
				// Broadcast round end
				h.broadcastRoundEnd(client.gameID)

				// Start next round
				if err := h.engine.StartRound(client.gameID); err != nil {
					log.Printf("Failed to start next round: %v", err)
					return
				}
			}
		}

		log.Printf("handleSelectCard: Broadcasting final game state")
		// Broadcast updated game state
		h.broadcastGameState(client.gameID)
		log.Printf("handleSelectCard: Complete")
	}
	log.Printf("handleSelectCard: Finished for player %s", client.playerID)
}

// handleWithdrawCard handles withdraw_card messages
func (h *WSHandler) handleWithdrawCard(client *Client, payload json.RawMessage) {
	log.Printf("handleWithdrawCard: Starting for player %s", client.playerID)

	// Withdraw the card selection
	if err := h.engine.WithdrawCard(client.gameID, client.playerID); err != nil {
		log.Printf("handleWithdrawCard: WithdrawCard failed: %v", err)
		h.sendError(client, "Failed to withdraw card: "+err.Error())
		return
	}

	log.Printf("handleWithdrawCard: Card withdrawn successfully, broadcasting state")

	// Broadcast updated game state
	h.broadcastGameState(client.gameID)

	log.Printf("handleWithdrawCard: Finished for player %s", client.playerID)
}

// handleKickPlayer handles kick_player messages
func (h *WSHandler) handleKickPlayer(client *Client, payload json.RawMessage) {
	log.Printf("handleKickPlayer: Starting for player %s", client.playerID)

	var data struct {
		PlayerID string `json:"playerId"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("handleKickPlayer: Failed to unmarshal payload: %v", err)
		h.sendError(client, "Invalid kick_player payload")
		return
	}

	// Get the game to check if it's in waiting phase
	game, err := h.engine.GetGame(client.gameID)
	if err != nil {
		log.Printf("handleKickPlayer: Failed to get game: %v", err)
		h.sendError(client, "Failed to get game: "+err.Error())
		return
	}

	// Only allow kicking in waiting phase
	if game.RoundPhase != models.PhaseWaitingForPlayers {
		h.sendError(client, "Can only kick players before game starts")
		return
	}

	// Send player_kicked message to the kicked player
	h.mu.RLock()
	if kickedClient, exists := h.clients[data.PlayerID]; exists {
		kickMsg := models.Message{
			Type:    models.MsgTypePlayerKicked,
			Payload: json.RawMessage(mustMarshal(map[string]string{"message": "You have been kicked from the game"})),
		}
		h.sendToClient(kickedClient, kickMsg)
	}
	h.mu.RUnlock()

	// Remove the player from the game (but don't close connection - let client handle that)
	h.mu.Lock()
	if gameClients, ok := h.games[client.gameID]; ok {
		delete(gameClients, data.PlayerID)
	}
	if _, exists := h.clients[data.PlayerID]; exists {
		delete(h.clients, data.PlayerID)
	}
	h.mu.Unlock()

	// Remove player from game engine
	if err := h.engine.RemovePlayer(client.gameID, data.PlayerID); err != nil {
		log.Printf("handleKickPlayer: RemovePlayer failed: %v", err)
		h.sendError(client, "Failed to kick player: "+err.Error())
		return
	}

	log.Printf("handleKickPlayer: Player %s kicked successfully, broadcasting state", data.PlayerID)

	// Broadcast updated game state
	h.broadcastGameState(client.gameID)

	log.Printf("handleKickPlayer: Finished for player %s", client.playerID)
}

// handleListGames handles list_games messages
func (h *WSHandler) handleListGames(client *Client) {
	games := h.engine.ListGames()

	msg := models.Message{
		Type:    models.MsgTypeListGames,
		Payload: json.RawMessage(mustMarshal(map[string]interface{}{"games": games})),
	}

	h.sendToClient(client, msg)
}

// handleDeleteGame handles delete_game messages
func (h *WSHandler) handleDeleteGame(client *Client, payload json.RawMessage) {
	var data struct {
		GameID string `json:"gameId"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("handleDeleteGame: Failed to unmarshal payload: %v", err)
		h.sendError(client, "Invalid delete_game payload")
		return
	}

	// Notify all players in the game that it's being deleted
	h.mu.RLock()
	if gameClients, ok := h.games[data.GameID]; ok {
		deleteMsg := models.Message{
			Type:    models.MsgTypeGameDeleted,
			Payload: json.RawMessage(mustMarshal(map[string]string{"message": "This game has been deleted"})),
		}
		for _, gameClient := range gameClients {
			h.sendToClient(gameClient, deleteMsg)
		}
	}
	h.mu.RUnlock()

	// Delete the game from the engine first
	if err := h.engine.DeleteGame(data.GameID); err != nil {
		log.Printf("handleDeleteGame: DeleteGame failed: %v", err)
		h.sendError(client, "Failed to delete game: "+err.Error())
		return
	}

	// Remove game from tracking (but don't close connections - let clients handle that)
	h.mu.Lock()
	if gameClients, ok := h.games[data.GameID]; ok {
		for playerID := range gameClients {
			delete(h.clients, playerID)
		}
		delete(h.games, data.GameID)
	}
	h.mu.Unlock()

	log.Printf("handleDeleteGame: Game %s deleted successfully", data.GameID)

	// Broadcast updated games list to all connected clients
	h.broadcastGamesList()
}

// broadcastGameState sends the current game state to all players in a game
func (h *WSHandler) broadcastGameState(gameID string) {
	log.Printf("broadcastGameState: Starting for game %s", gameID)
	game, err := h.engine.GetGame(gameID)
	if err != nil {
		log.Printf("Failed to get game: %v", err)
		return
	}

	log.Printf("broadcastGameState: Got game, phase=%s, round=%d", game.RoundPhase, game.CurrentRound)

	h.mu.RLock()
	clients := h.games[gameID]
	h.mu.RUnlock()

	log.Printf("broadcastGameState: Broadcasting to %d clients", len(clients))

	// Send personalized game state to each player
	for playerID, client := range clients {
		log.Printf("broadcastGameState: Building state for player %s", playerID)
		gameState := h.buildGameState(game, playerID)
		msg := models.Message{
			Type:    models.MsgTypeGameState,
			Payload: json.RawMessage(mustMarshal(gameState)),
		}
		log.Printf("broadcastGameState: Sending to player %s", playerID)
		h.sendToClient(client, msg)
	}
	log.Printf("broadcastGameState: Complete for game %s", gameID)
}

// broadcastRoundEnd sends round_end message to all players
func (h *WSHandler) broadcastRoundEnd(gameID string) {
	game, err := h.engine.GetGame(gameID)
	if err != nil {
		return
	}

	payload := map[string]interface{}{
		"round": game.CurrentRound,
	}

	msg := models.Message{
		Type:    models.MsgTypeRoundEnd,
		Payload: json.RawMessage(mustMarshal(payload)),
	}

	h.BroadcastToGame(gameID, msg)
}

// broadcastGameEnd sends game_end message to all players
func (h *WSHandler) broadcastGameEnd(gameID string, result *engine.GameResult) {
	msg := models.Message{
		Type:    models.MsgTypeGameEnd,
		Payload: json.RawMessage(mustMarshal(result)),
	}

	h.BroadcastToGame(gameID, msg)
}

// broadcastGamesList sends updated games list to all connected clients
func (h *WSHandler) broadcastGamesList() {
	games := h.engine.ListGames()

	msg := models.Message{
		Type:    models.MsgTypeListGames,
		Payload: json.RawMessage(mustMarshal(map[string]interface{}{"games": games})),
	}

	// Send to all connected clients
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.allConnections {
		h.sendToClient(client, msg)
	}

	log.Printf("Broadcasted games list to %d clients", len(h.allConnections))
}

// buildGameState creates a game state for a specific player
func (h *WSHandler) buildGameState(game *models.Game, playerID string) map[string]interface{} {
	players := make([]map[string]interface{}, len(game.Players))
	var myHand []models.Card

	for i, player := range game.Players {
		hasSelected := player.SelectedCard != nil

		players[i] = map[string]interface{}{
			"id":            player.ID,
			"name":          player.Name,
			"handSize":      len(player.Hand),
			"collection":    player.Collection,
			"puddingCards":  player.PuddingCards,
			"score":         player.Score,
			"hasSelected":   hasSelected,
			"roundScores":   player.RoundScores,
			"hasChopsticks": player.HasChopsticks,
		}

		// Include hand only for the requesting player
		if player.ID == playerID {
			myHand = player.Hand
		}
	}

	return map[string]interface{}{
		"gameId":       game.ID,
		"players":      players,
		"currentRound": game.CurrentRound,
		"phase":        game.RoundPhase,
		"myPlayerId":   playerID,
		"myHand":       myHand,
	}
}

// BroadcastToGame sends a message to all players in a game
func (h *WSHandler) BroadcastToGame(gameID string, message models.Message) error {
	h.mu.RLock()
	clients := h.games[gameID]
	h.mu.RUnlock()

	for _, client := range clients {
		h.sendToClient(client, message)
	}

	return nil
}

// SendToPlayer sends a message to a specific player
func (h *WSHandler) SendToPlayer(gameID, playerID string, message models.Message) error {
	h.mu.RLock()
	client := h.clients[playerID]
	h.mu.RUnlock()

	if client == nil {
		return nil
	}

	h.sendToClient(client, message)
	return nil
}

// sendToClient sends a message to a client
func (h *WSHandler) sendToClient(client *Client, message models.Message) {
	// Recover from panic if channel is closed
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in sendToClient: %v", r)
		}
	}()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		// Client's send channel is full or closed, skip
		log.Printf("Failed to send to client, channel full or closed")
	}
}

// sendError sends an error message to a client
func (h *WSHandler) sendError(client *Client, errorMsg string) {
	msg := models.Message{
		Type:    models.MsgTypeError,
		Payload: json.RawMessage(mustMarshal(map[string]string{"error": errorMsg})),
	}
	h.sendToClient(client, msg)
}

// removeClient removes a client from the handler
func (h *WSHandler) removeClient(client *Client) {
	h.mu.Lock()

	// Remove from all connections
	delete(h.allConnections, client)

	if client.playerID != "" {
		delete(h.clients, client.playerID)
	}

	gameToDelete := ""
	if client.gameID != "" && client.playerID != "" {
		if gameClients, ok := h.games[client.gameID]; ok {
			delete(gameClients, client.playerID)
			if len(gameClients) == 0 {
				delete(h.games, client.gameID)
				gameToDelete = client.gameID
			}
		}
	}

	log.Printf("Client disconnected, total connections: %d", len(h.allConnections))

	h.mu.Unlock()

	// If game has no more clients, delete it from the engine
	if gameToDelete != "" {
		log.Printf("All players disconnected from game %s, deleting game", gameToDelete)
		if err := h.engine.DeleteGame(gameToDelete); err != nil {
			log.Printf("Failed to delete empty game %s: %v", gameToDelete, err)
		} else {
			// Broadcast updated games list to all connected clients
			h.broadcastGamesList()
		}
	}
}

// mustMarshal marshals data to JSON or panics
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal: %v", err)
		return []byte("{}")
	}
	return data
}
