# Playtest Runner

A simple test runner for executing automated WebSocket-based game flow tests against the Sushi Go backend.

## Overview

The playtest runner simulates multiple WebSocket clients interacting with the backend server. Test scenarios are defined in YAML files that specify sequences of client actions. The runner executes these scenarios, captures state, and validates behavior.

## Installation

Build the playtest runner:

```bash
cd backend
go build -o playtest.exe ./playtest/cmd
```

## Usage

### Running a Single Test

```bash
.\playtest.exe <path-to-test-file.yaml>
```

Example:
```bash
.\playtest.exe .\playtest\tests\two-players-one-turn.yaml
```

### Running All Tests in a Directory

```bash
.\playtest.exe <path-to-directory>
```

Example:
```bash
.\playtest.exe .\playtest\tests
```

### Command-Line Options

- `--server <URL>`: WebSocket server URL (default: `ws://localhost:8080/ws`)
- `--start-server`: Start a test server on a random port (useful for isolated testing)
- `--verbose`: Print state snapshot after each turn for detailed debugging

Examples:
```bash
# Use a different server
.\playtest.exe --server ws://localhost:9000/ws .\playtest\tests\two-players-one-turn.yaml

# Start a test server automatically
.\playtest.exe --start-server .\playtest\tests\two-players-one-turn.yaml

# Enable verbose output
.\playtest.exe --verbose .\playtest\tests\two-players-one-turn.yaml

# Combine options
.\playtest.exe --start-server --verbose .\playtest\tests\two-players-one-turn.yaml
```

## Test File Format

Test files are written in YAML and define a sequence of turns. Each turn specifies:
- A client identifier (e.g., "A", "B", "Player1")
- A message to send to the server

### Basic Structure

```yaml
turns:
  - client: A
    message:
      type: <message_type>
      payload:
        <field>: <value>
```

### Message Types

The runner supports all WebSocket message types:

- `join_game`: Join or create a game
- `start_game`: Start a game
- `select_card`: Select a card during gameplay

### Variable Substitution

The runner automatically captures values from server responses and makes them available for use in subsequent messages using field path notation.

#### Field Path Syntax

Use `<response.field.path>` to reference any field from the last server response:

- `<response.gameId>`: Extract the gameId field
- `<response.players[0].id>`: Extract the first player's ID
- `<response.myPlayerId>`: Extract your player ID
- `<response.currentRound>`: Extract the current round number

#### Backward Compatibility

- `<globalGame>`: Still supported - automatically set to the `gameId` from any server response

#### Examples

**Basic field access:**
```yaml
turns:
  - client: A
    message:
      type: join_game
      payload:
        gameId: ""
        playerName: "Player A"
  
  # Use the gameId from the previous response
  - client: B
    message:
      type: join_game
      payload:
        gameId: <response.gameId>
        playerName: "Player B"
```

**Array access:**
```yaml
# Access nested fields and array elements
- client: A
  message:
    type: some_action
    payload:
      targetPlayerId: <response.players[0].id>
```

**Multiple field references:**
```yaml
- client: A
  message:
    type: complex_action
    payload:
      gameId: <response.gameId>
      round: <response.currentRound>
      playerId: <response.myPlayerId>
```

### Complete Example

```yaml
turns:
  # Player A creates a new game
  - client: A
    message:
      type: join_game
      payload:
        gameId: ""
        playerName: "Player A"
  
  # Player B joins the game
  - client: B
    message:
      type: join_game
      payload:
        gameId: <globalGame>
        playerName: "Player B"
  
  # Player A starts the game
  - client: A
    message:
      type: start_game
      payload:
        gameId: <globalGame>
  
  # Both players select cards
  - client: A
    message:
      type: select_card
      payload:
        cardIndex: 0
        useChopsticks: false
  
  - client: B
    message:
      type: select_card
      payload:
        cardIndex: 0
        useChopsticks: false
```

## Output

### Success

```
2025/11/22 23:56:43 Running playtest: .\playtest\tests\two-players-one-turn.yaml
2025/11/22 23:56:43 Executing turn 1: Client A
2025/11/22 23:56:43 [Client A] Sending: {"payload":{"gameId":"","playerName":"Player A"},"type":"join_game"}
2025/11/22 23:56:43 [Client A] Received: game_state
...
2025/11/22 23:56:43 Playtest completed successfully
✓ Test PASSED
```

Exit code: 0

### Failure

```
✗ Test FAILED
Error: turn 2: variable substitution failed: variable <globalGame> not found in store (available: )

Final Client States:

Client A:
  Messages Received: 1

Client B:
  Messages Received: 0
```

Exit code: 1

### Verbose Mode

With `--verbose`, the runner prints a state snapshot after each turn:

```
--- State Snapshot ---

Client A:
  Game State: {
    "currentRound": 1,
    "gameId": "abc123",
    "myHand": [...],
    "myPlayerId": "player1",
    "phase": "selecting",
    "players": [...]
  }
  Total Messages: 3
--- End Snapshot ---
```

## Architecture

The playtest runner consists of four main components:

1. **YAML Parser**: Reads and validates playtest definition files
2. **Client Simulator**: Manages WebSocket connections and message exchange
3. **Variable Store**: Stores and retrieves values from server responses
4. **Test Runner**: Orchestrates test execution and reports results

## Tips

- Use descriptive client identifiers (e.g., "PlayerA", "Host", "Guest")
- Start the backend server before running tests
- Use `--verbose` mode to debug failing tests
- Empty `gameId` creates a new game; use `<globalGame>` to join existing games
- The runner waits for a response after each message (5-second timeout)

## Troubleshooting

### Connection Refused

```
Error: turn 1: failed to create client A: failed to connect to ws://localhost:8080/ws: dial tcp [::1]:8080: connectex: No connection could be made...
```

**Solution**: Start the backend server first:
```bash
cd backend
go run main.go
```

### Variable Not Found

```
Error: turn 2: variable substitution failed: variable <globalGame> not found in store
```

**Solution**: Ensure the previous turn received a response containing `gameId`. Check that the server is responding correctly.

### Timeout Waiting for Message

```
Error: turn 1: failed to receive response: timeout waiting for message
```

**Solution**: The server may not be responding. Check server logs and ensure the message format is correct.
