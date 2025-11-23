package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// TestRunner orchestrates playtest execution
type TestRunner struct {
	clients      map[string]*ClientSimulator
	store        *VariableStore
	serverURL    string
	verbose      bool
	playtest     *PlaytestDefinition
	clientIDMap  map[string]string // client name -> player ID
}

// NewTestRunner creates a new test runner
func NewTestRunner(serverURL string) *TestRunner {
	return &TestRunner{
		clients:     make(map[string]*ClientSimulator),
		store:       NewVariableStore(),
		serverURL:   serverURL,
		verbose:     false,
		clientIDMap: make(map[string]string),
	}
}

// SetVerbose enables or disables verbose output
func (r *TestRunner) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// executeTurn executes a single turn in the playtest
func (r *TestRunner) executeTurn(turn Turn, turnNum int) error {
	// Get or create client
	client, exists := r.clients[turn.Client]
	if !exists {
		var err error
		client, err = NewClientSimulator(turn.Client, r.serverURL)
		if err != nil {
			return fmt.Errorf("turn %d: failed to create client %s: %w", turnNum, turn.Client, err)
		}
		r.clients[turn.Client] = client
	}

	// Substitute variables in the message
	message, err := r.store.Substitute(turn.Message)
	if err != nil {
		return fmt.Errorf("turn %d: variable substitution failed: %w", turnNum, err)
	}

	// Process card selection if this is a select_card message
	if msgType, ok := message["type"].(string); ok && msgType == "select_card" {
		if payload, ok := message["payload"].(map[string]interface{}); ok {
			// Get the most recent game state
			gameState := client.GetGameState()
			
			// Only process if we have a hand to work with
			if myHand, ok := gameState["myHand"].([]interface{}); ok && len(myHand) > 0 {
				processedPayload, err := ProcessSelectCardPayload(payload, gameState)
				if err != nil {
					return fmt.Errorf("turn %d: failed to process card selection: %w", turnNum, err)
				}
				message["payload"] = processedPayload
			}
			// If no hand yet, leave cardIndex as-is (might be a number already)
		}
	}

	// Send the message
	if err := client.SendMessage(message); err != nil {
		return fmt.Errorf("turn %d: failed to send message: %w", turnNum, err)
	}

	// Wait for response (with timeout)
	response, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		return fmt.Errorf("turn %d: failed to receive response: %w", turnNum, err)
	}

	// Parse response and extract variables
	var parsed map[string]interface{}
	if err := json.Unmarshal(response, &parsed); err == nil {
		if payload, ok := parsed["payload"].(map[string]interface{}); ok {
			r.store.ExtractAndStore(payload)
		}
	}

	return nil
}

// RunPlaytest executes a complete playtest scenario
func (r *TestRunner) RunPlaytest(filepath string) error {
	log.Printf("Running playtest: %s", filepath)

	// Parse the playtest file
	playtest, err := ParsePlaytest(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse playtest: %w", err)
	}

	r.playtest = playtest

	// Execute each turn sequentially
	for i, turn := range playtest.Turns {
		log.Printf("Executing turn %d: Client %s", i+1, turn.Client)
		if err := r.executeTurn(turn, i+1); err != nil {
			return err
		}
		
		// Print state snapshot if verbose mode is enabled
		if r.verbose {
			r.PrintStateSnapshot()
		}
	}

	log.Printf("Playtest completed successfully")
	return nil
}

// Close closes all client connections
func (r *TestRunner) Close() {
	for _, client := range r.clients {
		client.Close()
	}
}

// PrintResults prints the test results and client states
func (r *TestRunner) PrintResults(success bool, err error) {
	if success {
		fmt.Println("✓ Test PASSED")
	} else {
		fmt.Println("✗ Test FAILED")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		
		// Print final state of all clients
		fmt.Println("\nFinal Client States:")
		for clientID, client := range r.clients {
			fmt.Printf("\nClient %s:\n", clientID)
			gameState := client.GetGameState()
			if len(gameState) > 0 {
				stateJSON, _ := json.MarshalIndent(gameState, "  ", "  ")
				fmt.Printf("  Game State: %s\n", string(stateJSON))
			}
			messages := client.GetMessages()
			fmt.Printf("  Messages Received: %d\n", len(messages))
		}
	}
}

// PrintStateSnapshot prints the current state of all clients
func (r *TestRunner) PrintStateSnapshot() {
	fmt.Println("\n--- State Snapshot ---")
	for clientID, client := range r.clients {
		fmt.Printf("\nClient %s:\n", clientID)
		
		gameState := client.GetGameState()
		if len(gameState) > 0 {
			stateJSON, _ := json.MarshalIndent(gameState, "  ", "  ")
			fmt.Printf("  Game State: %s\n", string(stateJSON))
		} else {
			fmt.Println("  Game State: (none)")
		}
		
		messages := client.GetMessages()
		fmt.Printf("  Total Messages: %d\n", len(messages))
	}
	fmt.Println("--- End Snapshot ---\n")
}
