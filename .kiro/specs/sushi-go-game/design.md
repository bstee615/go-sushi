# Design Document

## Overview

The Sushi Go! game implementation consists of a Go backend server and a React frontend client. The backend manages game state, enforces rules, and provides a WebSocket API for real-time multiplayer communication. The frontend renders the game interface using Canvas for card visualization and provides an interactive UI built with React, Tailwind CSS, and DaisyUI.

The architecture follows a client-server model where the authoritative game state resides on the server, and clients receive state updates via WebSocket connections. This ensures consistency, prevents cheating, and enables real-time multiplayer gameplay.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Game UI    │  │  Canvas      │  │  WebSocket   │      │
│  │  Components  │  │  Renderer    │  │   Client     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ WebSocket
                            │
┌─────────────────────────────────────────────────────────────┐
│                     Backend (Go)                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  WebSocket   │  │    Game      │  │    Card      │      │
│  │   Handler    │  │   Engine     │  │   Scoring    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │    State     │  │  Validation  │                        │
│  │   Manager    │  │    Logic     │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

**Backend:**
- Go 1.21+ for server implementation
- Gorilla WebSocket for real-time communication
- Standard library for HTTP server and JSON handling

**Frontend:**
- React 18+ for UI components
- TypeScript for type safety
- Tailwind CSS for styling
- DaisyUI for component library
- HTML5 Canvas for card rendering
- WebSocket API for real-time updates

## Components and Interfaces

### Backend Components

#### 1. Game Engine

The Game Engine is the core component that manages game state and enforces rules.

**Responsibilities:**
- Initialize new game sessions
- Manage round progression
- Execute card drafting logic
- Validate player moves
- Calculate scores
- Determine winners

**Key Methods:**
```go
type GameEngine interface {
    CreateGame(playerIDs []string) (*Game, error)
    StartRound(gameID string) error
    PlayCard(gameID, playerID string, cardIndex int, useChopsticks bool) error
    RevealCards(gameID string) error
    PassHands(gameID string) error
    ScoreRound(gameID string) error
    EndGame(gameID string) (*GameResult, error)
}
```

#### 2. Card Scoring System

Handles all scoring logic for different card types.

**Responsibilities:**
- Score Maki Rolls with competitive ranking
- Score set collection cards (Tempura, Sashimi, Dumplings)
- Score Nigiri with Wasabi modifiers
- Score Pudding at game end
- Calculate final scores

**Key Methods:**
```go
type ScoringSystem interface {
    ScoreMakiRolls(players []*Player) map[string]int
    ScoreTempura(cards []Card) int
    ScoreSashimi(cards []Card) int
    ScoreDumplings(cards []Card) int
    ScoreNigiri(cards []Card, wasabiCards []Card) int
    ScorePudding(players []*Player, playerCount int) map[string]int
}
```

#### 3. State Manager

Manages game state persistence and retrieval.

**Responsibilities:**
- Store active game sessions
- Serialize/deserialize game state
- Handle concurrent access to game state
- Provide state snapshots for clients

**Key Methods:**
```go
type StateManager interface {
    SaveGame(game *Game) error
    LoadGame(gameID string) (*Game, error)
    DeleteGame(gameID string) error
    GetGameState(gameID string) (*GameState, error)
}
```

#### 4. WebSocket Handler

Manages WebSocket connections and message routing.

**Responsibilities:**
- Accept WebSocket connections
- Route messages to appropriate handlers
- Broadcast state updates to clients
- Handle client disconnections

**Key Methods:**
```go
type WebSocketHandler interface {
    HandleConnection(w http.ResponseWriter, r *http.Request)
    BroadcastToGame(gameID string, message Message) error
    SendToPlayer(gameID, playerID string, message Message) error
}
```

#### 5. Validation Logic

Validates all player actions and game state transitions.

**Responsibilities:**
- Validate card selection
- Validate Chopsticks usage
- Validate game state transitions
- Prevent invalid moves

**Key Methods:**
```go
type Validator interface {
    ValidateCardSelection(game *Game, playerID string, cardIndex int) error
    ValidateChopsticksUse(game *Game, playerID string) error
    ValidateGameState(game *Game) error
}
```

### Frontend Components

#### 1. Game Board Component

Main container for the game interface.

**Responsibilities:**
- Display all players and their scores
- Show current round information
- Coordinate child components
- Handle game flow UI

#### 2. Hand Component

Displays the player's current hand of cards.

**Responsibilities:**
- Render cards using Canvas
- Handle card selection
- Show card hover effects
- Display Chopsticks usage option

#### 3. Collection Component

Shows cards collected by the player.

**Responsibilities:**
- Group cards by type
- Display Wasabi-Nigiri combinations
- Show running score calculations
- Animate card additions

#### 4. Canvas Card Renderer

Renders individual cards on Canvas.

**Responsibilities:**
- Draw card artwork
- Display card values
- Handle card animations
- Optimize rendering performance

#### 5. Score Display Component

Shows player scores and rankings.

**Responsibilities:**
- Display current round scores
- Show cumulative scores
- Display score breakdowns
- Highlight winner

#### 6. WebSocket Client

Manages connection to backend server.

**Responsibilities:**
- Establish WebSocket connection
- Send player actions
- Receive state updates
- Handle reconnection

## Data Models

### Core Game Models

```go
type Game struct {
    ID            string
    Players       []*Player
    Deck          []Card
    CurrentRound  int
    RoundPhase    RoundPhase
    CreatedAt     time.Time
}

type Player struct {
    ID              string
    Name            string
    Hand            []Card
    Collection      []Card
    PuddingCards    []Card
    Score           int
    RoundScores     []int
    HasChopsticks   bool
    SelectedCard    *int
}

type Card struct {
    ID       string
    Type     CardType
    Variant  string  // For Nigiri types (Squid, Salmon, Egg)
    Value    int     // Maki roll count or base points
}

type CardType string

const (
    CardTypeMakiRoll   CardType = "maki_roll"
    CardTypeTempura    CardType = "tempura"
    CardTypeSashimi    CardType = "sashimi"
    CardTypeDumpling   CardType = "dumpling"
    CardTypeNigiri     CardType = "nigiri"
    CardTypeWasabi     CardType = "wasabi"
    CardTypeChopsticks CardType = "chopsticks"
    CardTypePudding    CardType = "pudding"
)

type RoundPhase string

const (
    PhaseWaitingForPlayers RoundPhase = "waiting"
    PhaseSelecting         RoundPhase = "selecting"
    PhaseRevealing         RoundPhase = "revealing"
    PhasePassing           RoundPhase = "passing"
    PhaseScoring           RoundPhase = "scoring"
    PhaseRoundEnd          RoundPhase = "round_end"
    PhaseGameEnd           RoundPhase = "game_end"
)

type GameState struct {
    GameID       string
    Players      []PlayerState
    CurrentRound int
    Phase        RoundPhase
}

type PlayerState struct {
    ID           string
    Name         string
    HandSize     int
    Collection   []Card
    Score        int
    HasSelected  bool
}
```

### WebSocket Message Models

```go
type Message struct {
    Type    MessageType
    Payload json.RawMessage
}

type MessageType string

const (
    MsgTypeJoinGame      MessageType = "join_game"
    MsgTypeStartGame     MessageType = "start_game"
    MsgTypeSelectCard    MessageType = "select_card"
    MsgTypeGameState     MessageType = "game_state"
    MsgTypeCardRevealed  MessageType = "card_revealed"
    MsgTypeRoundEnd      MessageType = "round_end"
    MsgTypeGameEnd       MessageType = "game_end"
    MsgTypeError         MessageType = "error"
)

type SelectCardPayload struct {
    CardIndex     int
    UseChopsticks bool
    SecondCardIndex *int  // Only when using Chopsticks
}
```

### Frontend TypeScript Models

```typescript
interface GameState {
  gameId: string;
  players: PlayerState[];
  currentRound: number;
  phase: RoundPhase;
  myPlayerId: string;
}

interface PlayerState {
  id: string;
  name: string;
  handSize: number;
  collection: Card[];
  score: number;
  hasSelected: boolean;
}

interface Card {
  id: string;
  type: CardType;
  variant?: string;
  value?: number;
}

type CardType = 
  | 'maki_roll'
  | 'tempura'
  | 'sashimi'
  | 'dumpling'
  | 'nigiri'
  | 'wasabi'
  | 'chopsticks'
  | 'pudding';

type RoundPhase = 
  | 'waiting'
  | 'selecting'
  | 'revealing'
  | 'passing'
  | 'scoring'
  | 'round_end'
  | 'game_end';
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Card dealing consistency
*For any* game session with a valid player count (2-5), the total number of cards dealt should equal the player count multiplied by the cards-per-player for that count (10 for 2 players, 9 for 3, 8 for 4, 7 for 5).
**Validates: Requirements 2.1, 2.2, 2.3, 2.4**

### Property 2: Hand passing preserves card count
*For any* round in progress, after cards are selected and hands are passed, the sum of all hand sizes plus the number of cards in all collections should equal the total cards dealt at the start of the round.
**Validates: Requirements 4.1, 4.2**

### Property 3: Maki roll scoring is zero-sum for top positions
*For any* completed round, the total points awarded for Maki rolls should be at most 9 points (6 for first place + 3 for second place), and when ties occur, the split points should sum to the original award amount.
**Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**

### Property 4: Tempura scoring is proportional to pairs
*For any* collection of Tempura cards, the score should equal 5 multiplied by the number of complete pairs (count divided by 2, rounded down).
**Validates: Requirements 6.1, 6.2, 6.3**

### Property 5: Sashimi scoring is proportional to triples
*For any* collection of Sashimi cards, the score should equal 10 multiplied by the number of complete sets of 3 (count divided by 3, rounded down).
**Validates: Requirements 7.1, 7.2, 7.3**

### Property 6: Dumpling scoring follows defined progression
*For any* collection of Dumpling cards, the score should match the official table: 1→1pt, 2→3pts, 3→6pts, 4→10pts, 5+→15pts.
**Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**

### Property 7: Nigiri base scoring is correct
*For any* Nigiri card without Wasabi, the score should be 3 for Squid, 2 for Salmon, and 1 for Egg.
**Validates: Requirements 9.1, 9.2, 9.3**

### Property 8: Wasabi triples Nigiri value exactly once
*For any* Wasabi card, when a Nigiri is placed on it, the Nigiri's score should be exactly triple its base value, and the Wasabi should not affect any subsequent Nigiri.
**Validates: Requirements 10.1, 10.2, 10.3, 10.5**

### Property 9: Chopsticks enables exactly two card selection
*For any* player with Chopsticks in their collection who activates them, the player should be able to select exactly 2 cards from their current hand, and the Chopsticks should return to the passed hand.
**Validates: Requirements 11.1, 11.2, 11.3, 11.4**

### Property 10: Pudding scoring is symmetric for most and least
*For any* game with more than 2 players, the player(s) with the most Pudding cards should gain 6 points each, and the player(s) with the fewest should lose 6 points each, with ties awarding/penalizing all tied players.
**Validates: Requirements 12.2, 12.3, 12.4, 12.5**

### Property 11: Round progression is sequential
*For any* game session, rounds should progress sequentially from 1 to 3, and the game should end after the third round completes.
**Validates: Requirements 16.1, 16.2, 16.3**

### Property 12: Score accumulation is monotonic per round
*For any* player across multiple rounds, the cumulative score should equal the sum of all round scores, and adding a new round score should never decrease the cumulative total (except for Pudding penalties at game end).
**Validates: Requirements 13.2**

### Property 13: Invalid card selection is rejected
*For any* player attempting to select a card not in their current hand, the backend should reject the move and the game state should remain unchanged.
**Validates: Requirements 19.1**

### Property 14: Game state serialization round-trip
*For any* valid game state, serializing to JSON and then deserializing should produce an equivalent game state with all players, cards, scores, and round information preserved.
**Validates: Requirements 20.1, 20.2, 20.3, 20.4**

### Property 15: Simultaneous card reveal
*For any* turn where all players have selected cards, the reveal should happen simultaneously such that no player's selected card is visible to others before all selections are complete.
**Validates: Requirements 3.3**

### Property 16: Collection persistence across rounds
*For any* player's Pudding cards, they should remain in the player's collection across all three rounds and not be cleared when new rounds start.
**Validates: Requirements 16.4**

### Property 17: Winner determination is deterministic
*For any* completed game, given the same final scores and Pudding counts, the winner determination should always produce the same result.
**Validates: Requirements 17.1, 17.2, 17.3**

### Property 18: Game ID format is consistent
*For any* generated game ID, it should match the format region-flower-number where region and flower are from predefined lists of Japanese regions and flowers.
**Validates: Requirements 21.1, 21.2, 21.3**

### Property 19: Player reconnection preserves state
*For any* player who reconnects to a game using their existing username, their complete game state (hand, collection, score, round progress) should be identical to their state before disconnection.
**Validates: Requirements 22.1, 22.2, 22.4**

## Random ID and Name Generation

### Game ID Generator

The system will generate memorable game IDs using the format: `region-flower-number`

**Japanese Regions:**
- Tokyo
- Kyoto
- Osaka
- Hokkaido
- Okinawa
- Nara
- Hiroshima
- Fukuoka
- Nagoya
- Sapporo

**Japanese Flowers:**
- Sakura (Cherry Blossom)
- Ume (Plum Blossom)
- Tsubaki (Camellia)
- Ajisai (Hydrangea)
- Kiku (Chrysanthemum)
- Fuji (Wisteria)
- Botan (Peony)
- Ayame (Iris)
- Momiji (Maple)
- Hasu (Lotus)

**Number Range:** 1-999

**Example Game IDs:**
- tokyo-sakura-42
- kyoto-ume-157
- osaka-kiku-888

### Player Name Generator

When a player joins without providing a name, the system will assign a random name from these categories:

**Famous Sushi Chefs:**
- Jiro Ono
- Masahiro Yoshitake
- Takashi Saito
- Hachiro Mizutani
- Keiji Nakazawa

**Pop Culture/Anime Characters:**
- Naruto
- Totoro
- Goku
- Pikachu
- Luffy
- Sailor Moon
- Doraemon
- Saitama
- Spike
- Ghibli

**Historical Figures:**
- Miyamoto Musashi
- Oda Nobunaga
- Tokugawa Ieyasu
- Minamoto Yoritomo
- Himiko

## Player Reconnection System

### Reconnection Logic

The system will support player reconnection by username within the same game session:

1. **Username Matching:** When a player joins a game, the backend checks if a player with that username already exists in the game
2. **State Restoration:** If a match is found, the player is reconnected to their existing player state instead of creating a new player
3. **WebSocket Update:** The new WebSocket connection is associated with the existing player ID
4. **State Broadcast:** Other players are notified of the reconnection
5. **Game Isolation:** Username matching only applies within the same game - the same username can exist in different games as different players

### Data Model Updates

```go
type Player struct {
    ID              string
    Name            string  // Used for reconnection matching
    Hand            []Card
    Collection      []Card
    PuddingCards    []Card
    Score           int
    RoundScores     []int
    HasChopsticks   bool
    SelectedCard    *int
    Connected       bool    // Track connection status
}
```

### Reconnection Flow

```
1. Player joins game with username "Jiro Ono"
2. Backend checks if "Jiro Ono" exists in game
3. If exists:
   - Load existing player state
   - Update connection status
   - Send full game state to reconnected player
   - Broadcast reconnection to other players
4. If not exists:
   - Create new player with that username
```

## Error Handling

### Backend Error Handling

**Connection Errors:**
- WebSocket connection failures should be logged and clients should receive connection error messages
- Disconnected players should trigger game pause notifications to other players
- Reconnection attempts should restore player state if game is still active

**Validation Errors:**
- Invalid moves should return specific error messages indicating the violation
- Game state should never be corrupted by invalid moves
- All validation errors should be logged for debugging

**State Errors:**
- Concurrent access to game state should be protected by mutexes
- State corruption should be detected and logged
- Unrecoverable state errors should end the game gracefully

**Scoring Errors:**
- Scoring calculation errors should be logged with full game state
- Fallback scoring should use conservative estimates
- Score discrepancies should be flagged for review

### Frontend Error Handling

**Network Errors:**
- WebSocket disconnections should trigger reconnection attempts with exponential backoff
- Failed reconnections should display user-friendly error messages
- Network errors should not crash the application

**Rendering Errors:**
- Canvas rendering failures should fall back to DOM-based card display
- Missing card assets should display placeholder graphics
- Rendering errors should be caught and logged

**State Sync Errors:**
- State mismatches between client and server should trigger full state refresh
- Optimistic UI updates should be rolled back on server rejection
- Sync errors should be logged for debugging

## Testing Strategy

### Unit Testing

**Backend Unit Tests:**
- Test each scoring function with specific card combinations
- Test card dealing logic for all player counts
- Test hand passing mechanics
- Test Wasabi and Chopsticks special logic
- Test validation functions with valid and invalid inputs
- Test state serialization/deserialization

**Frontend Unit Tests:**
- Test React component rendering
- Test WebSocket message handling
- Test UI state management
- Test card selection logic
- Test score display calculations

### Property-Based Testing

The implementation will use property-based testing to verify correctness properties across many randomly generated game states. For Go, we'll use the `gopter` library. For TypeScript/JavaScript, we'll use `fast-check`.

**Configuration:**
- Each property test should run a minimum of 100 iterations
- Tests should generate random but valid game states
- Edge cases (empty hands, maximum cards, ties) should be included in generators

**Property Test Tagging:**
Each property-based test must include a comment tag in this exact format:
```go
// Feature: sushi-go-game, Property 1: Card dealing consistency
```

**Property Test Implementation:**
- Each correctness property listed above must be implemented as a single property-based test
- Tests should generate random game states within valid constraints
- Tests should verify the property holds across all generated inputs
- Failed tests should provide clear counterexamples

**Backend Property Tests:**
- Property 1: Generate games with random player counts, verify card dealing
- Property 2: Generate random rounds, verify card conservation
- Property 3-10: Generate random card collections, verify scoring formulas
- Property 11-17: Generate random game progressions, verify state transitions

**Frontend Property Tests:**
- Generate random game states, verify UI renders correctly
- Generate random card selections, verify state updates
- Generate random score data, verify display calculations

### Integration Testing

**End-to-End Tests:**
- Test complete game flow from creation to winner determination
- Test multiplayer scenarios with simulated players
- Test reconnection and error recovery
- Test all card type combinations

**WebSocket Integration Tests:**
- Test message routing between clients and server
- Test broadcast functionality
- Test concurrent player actions
- Test connection handling

### Test Utilities

**Test Data Generators:**
- Generate random valid game states
- Generate random card collections
- Generate random player actions
- Generate edge case scenarios

**Test Fixtures:**
- Predefined game states for specific scenarios
- Sample card collections for scoring tests
- Mock WebSocket connections
- Mock player clients

## Performance Considerations

**Backend:**
- Use goroutines for concurrent game session handling
- Implement connection pooling for WebSocket connections
- Cache scoring calculations where appropriate
- Use efficient data structures for card collections

**Frontend:**
- Optimize Canvas rendering with requestAnimationFrame
- Implement virtual scrolling for large player lists
- Debounce WebSocket message handling
- Use React.memo for expensive components

## Security Considerations

**Backend:**
- Validate all client inputs
- Prevent state manipulation through message tampering
- Implement rate limiting on WebSocket messages
- Use secure WebSocket connections (WSS) in production

**Frontend:**
- Sanitize all user inputs
- Validate server messages before processing
- Implement CSRF protection
- Use secure WebSocket connections

## Deployment Considerations

**Backend:**
- Containerize Go server with Docker
- Use environment variables for configuration
- Implement health check endpoints
- Support horizontal scaling with shared state store

**Frontend:**
- Build optimized production bundle
- Serve static assets via CDN
- Implement code splitting for faster load times
- Use service workers for offline support

## Future Enhancements

- Persistent game history and statistics
- Spectator mode for watching games
- AI opponents for single-player mode
- Tournament mode with brackets
- Custom card sets and variants
- Mobile-responsive design
- Internationalization support
