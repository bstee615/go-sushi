# Sushi Go! Test Frontend

A simple, barebones HTML/JavaScript frontend for testing the Sushi Go! game backend.

## Features

- Pure HTML/CSS/JavaScript (no build tools required)
- WebSocket connection to game server
- Join and start games
- View game state and player information
- Select cards during gameplay
- Real-time message logging

## Usage

### 1. Start the Backend Server

```bash
cd backend
go run main.go
```

The server will start on `http://localhost:8080`

### 2. Open the Test Frontend

Simply open `index.html` in your web browser:

```bash
# On Windows
start test-frontend/index.html

# On macOS
open test-frontend/index.html

# On Linux
xdg-open test-frontend/index.html
```

Or just double-click the `index.html` file.

### 3. Connect and Play

#### Creating a New Game

1. **Connect**: Click "Connect" to establish WebSocket connection
2. **Enter Name**: Type your player name
3. **Create Game**: Click "Create New Game" button
4. **Share Game ID**: The game ID will appear below the buttons - click "Copy" to share it with other players
5. **Wait for Players**: At least 2 players are needed to start
6. **Start Game**: Once enough players have joined, click "Start Game"
7. **Play**: Click on cards in your hand to select them during the selection phase

#### Joining an Existing Game

1. **Connect**: Click "Connect" to establish WebSocket connection
2. **Enter Name**: Type your player name
3. **Enter Game ID**: Paste or type the game ID you want to join
4. **Join Game**: Click "Join Existing Game" button
5. **Wait**: Wait for the host to start the game

## Testing Multiple Players

To test multiplayer functionality:

1. **First Player**: Open `index.html` and click "Create New Game"
2. **Copy Game ID**: Click the "Copy" button next to the displayed game ID
3. **Second Player**: Open `index.html` in another browser window/tab
4. **Join**: Paste the game ID and click "Join Existing Game"
5. **Start**: Once both players are connected, click "Start Game" from any window

## UI Components

- **Connection Status**: Shows whether you're connected to the server
- **Controls**: Connect, join game, and start game buttons
- **Your Hand**: Displays your current hand of cards (click to select)
- **Players**: Shows all players, their scores, and collection status
- **Your Collection**: Shows cards you've collected this round
- **Message Log**: Real-time log of all WebSocket messages

## Message Types

The client handles these message types:

- `game_state`: Updates the game state display
- `card_revealed`: Shows when cards are revealed
- `round_end`: Displays round end information
- `game_end`: Shows final game results
- `error`: Displays error messages

## Limitations

This is a minimal test client with some limitations:

- Cannot see actual card details in hand (only card count)
- No chopsticks support yet
- Basic styling only
- No animations or advanced UI features

For a full-featured game experience, use the main React frontend in the `frontend/` directory.

## File Structure

```
test-frontend/
├── index.html    # Main HTML page with UI structure and styles
├── game.js       # Game logic and WebSocket communication
└── README.md     # This file
```

## Troubleshooting

**Cannot connect to server:**
- Make sure the backend is running on `localhost:8080`
- Check the server URL in the input field
- Look at the browser console for error messages

**Cards not showing:**
- The game state only includes card count, not actual cards
- This is expected behavior for the current implementation
- Cards will show in the collection after being played

**Multiple players not working:**
- Make sure all players use the same game ID
- Each player needs a unique name
- All players must be connected before starting the game
