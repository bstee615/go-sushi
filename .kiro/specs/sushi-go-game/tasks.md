# Implementation Plan

- [x] 1. Set up project structure and dependencies





  - Create Go module for backend with necessary dependencies (gorilla/websocket, gopter for property testing)
  - Create React TypeScript project with Vite
  - Install frontend dependencies (Tailwind CSS, DaisyUI, fast-check for property testing)
  - Set up project directory structure for backend (handlers, engine, models, scoring) and frontend (components, hooks, utils)
  - _Requirements: All_

- [-] 2. Implement core data models



  - Define Go structs for Game, Player, Card, CardType, RoundPhase, GameState
  - Define TypeScript interfaces for GameState, PlayerState, Card, CardType, RoundPhase
  - Implement card type constants and enums
  - _Requirements: 20.3_

- [-] 2.1 Write property test for game state serialization


  - **Property 14: Game state serialization round-trip**
  - **Validates: Requirements 20.1, 20.2, 20.3, 20.4**

- [ ] 3. Implement card deck and dealing logic
  - Create deck initialization with correct card distribution
  - Implement shuffle function
  - Implement dealing logic that distributes correct number of cards based on player count (10 for 2p, 9 for 3p, 8 for 4p, 7 for 5p)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 3.1 Write property test for card dealing consistency

  - **Property 1: Card dealing consistency**
  - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

- [ ] 4. Implement game session management
  - Create game creation logic with unique ID generation
  - Implement player join logic with validation (max 5 players)
  - Implement player ID assignment
  - Add game start logic (minimum 2 players)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_


- [ ] 4.1 Write property tests for game session management

  - **Property: Unique game ID generation**
  - **Property: Player join validation**
  - **Property: Unique player ID assignment**
  - **Validates: Requirements 1.1, 1.2, 1.4, 1.5**

- [ ] 5. Implement card selection and hand passing mechanics
  - Create card selection logic for players
  - Implement simultaneous reveal mechanism
  - Implement hand passing to the left
  - Add selected card removal from hand before passing
  - Implement round end detection (when last card is played)
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3_

- [ ] 5.1 Write property test for hand passing

  - **Property 2: Hand passing preserves card count**
  - **Validates: Requirements 4.1, 4.2**

- [ ] 5.2 Write property test for simultaneous reveal

  - **Property 15: Simultaneous card reveal**
  - **Validates: Requirements 3.3**

- [ ] 6. Implement Maki Roll scoring
  - Count Maki Roll icons for each player
  - Implement first place scoring (6 points, split on ties)
  - Implement second place scoring (3 points, split on ties)
  - Handle tie-breaking logic with proper point division
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_


- [ ] 6.1 Write property test for Maki Roll scoring

  - **Property 3: Maki roll scoring is zero-sum for top positions**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**

- [ ] 7. Implement set collection scoring (Tempura, Sashimi, Dumplings)
  - Implement Tempura scoring (5 points per pair)
  - Implement Sashimi scoring (10 points per set of 3)
  - Implement Dumpling scoring with progression table (1→1, 2→3, 3→6, 4→10, 5+→15)
  - _Requirements: 6.1, 6.2, 6.3, 7.1, 7.2, 7.3, 8.1, 8.2, 8.3, 8.4, 8.5_


- [ ] 7.1 Write property test for Tempura scoring

  - **Property 4: Tempura scoring is proportional to pairs**
  - **Validates: Requirements 6.1, 6.2, 6.3**

- [ ] 7.2 Write property test for Sashimi scoring

  - **Property 5: Sashimi scoring is proportional to triples**
  - **Validates: Requirements 7.1, 7.2, 7.3**

- [ ] 7.3 Write property test for Dumpling scoring

  - **Property 6: Dumpling scoring follows defined progression**
  - **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**

- [ ] 8. Implement Nigiri and Wasabi scoring
  - Implement base Nigiri scoring (Squid=3, Salmon=2, Egg=1)
  - Implement Wasabi state tracking (available/used)
  - Implement Wasabi tripling logic for next Nigiri
  - Handle unused Wasabi (0 points)
  - _Requirements: 9.1, 9.2, 9.3, 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ] 8.1 Write property test for Nigiri base scoring

  - **Property 7: Nigiri base scoring is correct**
  - **Validates: Requirements 9.1, 9.2, 9.3**

- [ ] 8.2 Write property test for Wasabi tripling

  - **Property 8: Wasabi triples Nigiri value exactly once**
  - **Validates: Requirements 10.1, 10.2, 10.3, 10.5**

- [ ] 9. Implement Chopsticks special card logic
  - Add Chopsticks activation logic
  - Implement two-card selection when Chopsticks are used
  - Implement Chopsticks return to passed hand
  - Add both selected cards to player collection
  - _Requirements: 3.5, 11.1, 11.2, 11.3, 11.4_


- [ ] 9.1 Write property test for Chopsticks mechanics

  - **Property 9: Chopsticks enables exactly two card selection**
  - **Validates: Requirements 11.1, 11.2, 11.3, 11.4**

- [ ] 10. Implement Pudding scoring
  - Track Pudding cards across all three rounds
  - Implement most Pudding scoring (6 points, all tied players get points)
  - Implement fewest Pudding penalty (-6 points, all tied players lose points)
  - Add special case for 2-player games (no penalty)
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6_

- [ ] 10.1 Write property test for Pudding scoring

  - **Property 10: Pudding scoring is symmetric for most and least**
  - **Validates: Requirements 12.2, 12.3, 12.4, 12.5**

- [ ] 10.2 Write property test for Pudding persistence

  - **Property 16: Collection persistence across rounds**
  - **Validates: Requirements 16.4**

- [ ] 11. Implement round progression and game flow
  - Initialize round counter to 1 at game start
  - Implement round end detection and scoring trigger
  - Increment round counter after each round
  - Clear non-Pudding cards between rounds
  - Trigger final scoring after round 3
  - End game after round 3
  - _Requirements: 4.4, 16.1, 16.2, 16.3, 16.4_

- [ ] 11.1 Write property test for round progression

  - **Property 11: Round progression is sequential**
  - **Validates: Requirements 16.1, 16.2, 16.3**

- [ ] 12. Implement winner determination and final scoring
  - Calculate final scores including Pudding points
  - Identify player with highest score as winner
  - Implement tiebreaker using Pudding card count
  - Generate final rankings
  - _Requirements: 17.1, 17.2, 17.3, 17.4_

- [ ] 12.1 Write property test for score accumulation

  - **Property 12: Score accumulation is monotonic per round**
  - **Validates: Requirements 13.2**

- [ ] 12.2 Write property test for winner determination

  - **Property 17: Winner determination is deterministic**
  - **Validates: Requirements 17.1, 17.2, 17.3**

- [ ] 13. Implement validation logic
  - Validate card selection (card must be in player's hand)
  - Validate Chopsticks usage (must have Chopsticks in collection)
  - Validate turn order (prevent out-of-turn moves)
  - Add error logging for invalid state
  - Prevent state corruption on validation failures
  - _Requirements: 19.1, 19.2, 19.3, 19.4_


- [ ] 13.1 Write property test for validation

  - **Property 13: Invalid card selection is rejected**
  - **Validates: Requirements 19.1**

- [ ] 14. Implement state management and serialization
  - Create state manager for storing active games
  - Implement JSON serialization for complete game state
  - Implement JSON deserialization with validation
  - Add mutex protection for concurrent access
  - Implement state snapshot generation for clients
  - _Requirements: 20.1, 20.2, 20.3, 20.4_

- [ ] 15. Checkpoint - Ensure all backend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 16. Implement WebSocket server and message handling
  - Set up WebSocket server with gorilla/websocket
  - Implement connection handler
  - Create message routing logic
  - Implement broadcast to all players in a game
  - Implement send to specific player
  - Handle client disconnections with game pause
  - _Requirements: 18.1, 18.2, 18.3, 18.4_

- [ ] 17. Define WebSocket message protocol
  - Define message types (join_game, start_game, select_card, game_state, card_revealed, round_end, game_end, error)
  - Create message payload structures
  - Implement message serialization/deserialization
  - _Requirements: 18.1, 18.2, 18.3_

- [ ] 18. Implement real-time game state updates
  - Push state updates on card selection (without revealing card)
  - Broadcast simultaneous reveal to all players
  - Push state updates on round end
  - Push state updates on game end
  - _Requirements: 18.1, 18.2, 18.3_

- [ ] 19. Set up React frontend project structure
  - Create component structure (GameBoard, Hand, Collection, ScoreDisplay, CardRenderer)
  - Set up routing if needed
  - Configure Tailwind CSS and DaisyUI
  - Set up TypeScript types
  - _Requirements: 13.1, 13.2, 13.3, 13.4, 14.1, 14.2, 14.3, 14.4, 15.1, 15.2, 15.3_

- [ ] 20. Implement WebSocket client
  - Create WebSocket connection hook
  - Implement message sending
  - Implement message receiving and parsing
  - Add reconnection logic with exponential backoff
  - Handle connection errors
  - _Requirements: 18.3_

- [ ] 21. Implement Canvas card renderer
  - Create Canvas component for rendering cards
  - Draw card artwork and borders
  - Display card type and point values
  - Implement card hover effects
  - Optimize rendering with requestAnimationFrame
  - _Requirements: 14.1, 14.2, 14.3_

- [ ] 22. Implement Hand component
  - Display all cards in player's hand using Canvas
  - Arrange cards in fan layout
  - Handle card selection on click
  - Show Chopsticks usage option when available
  - Prevent selection after card is chosen
  - _Requirements: 3.1, 3.2, 14.1, 14.4_

- [ ] 23. Implement Collection component
  - Display collected cards grouped by type
  - Show Wasabi-Nigiri pairings
  - Keep similar cards together
  - Animate card additions
  - _Requirements: 15.1, 15.2, 15.3_

- [ ] 24. Implement ScoreDisplay component
  - Display current round scores for all players
  - Display cumulative scores
  - Show score breakdown by card type
  - Display final scores with Pudding points
  - Highlight winner
  - _Requirements: 13.1, 13.2, 13.3, 13.4, 17.4_

- [ ] 25. Implement GameBoard component
  - Coordinate all child components
  - Display current round information
  - Show all players and their states
  - Handle game flow UI
  - Display waiting states
  - _Requirements: 13.1, 13.2_

- [ ] 26. Implement game lobby and join flow
  - Create game creation UI
  - Create game join UI with game ID input
  - Display waiting room with player list
  - Add start game button (enabled when 2+ players)
  - _Requirements: 1.1, 1.2, 1.3_

- [ ] 27. Connect frontend to backend via WebSocket
  - Implement game creation flow
  - Implement game join flow
  - Handle card selection and send to backend
  - Receive and apply state updates
  - Handle error messages from backend
  - _Requirements: 18.1, 18.2, 18.3_

- [ ] 28. Implement error handling and user feedback
  - Display connection errors
  - Display validation errors from backend
  - Show loading states
  - Handle disconnection with reconnection UI
  - Add fallback rendering for Canvas failures
  - _Requirements: 19.1, 19.2, 19.3_

- [ ] 29. Add game flow animations and transitions
  - Animate card selection
  - Animate card reveal
  - Animate hand passing
  - Animate score updates
  - Add round transition animations
  - _Requirements: 3.3, 4.1_

- [ ] 30. Implement responsive design
  - Make layout work on different screen sizes
  - Optimize Canvas rendering for different resolutions
  - Test on mobile devices
  - Adjust card sizes for smaller screens
  - _Requirements: 14.1, 14.4_

- [ ] 31. Write frontend property tests

  - Test card rendering with random game states
  - Test score display calculations with random scores
  - Test UI state management with random actions
  - _Requirements: 13.1, 13.2, 14.1_

- [ ] 32. Final checkpoint - End-to-end testing
  - Ensure all tests pass, ask the user if questions arise.
  - Test complete game flow from creation to winner
  - Test multiplayer with multiple browser windows
  - Test all card types and scoring
  - Test edge cases (disconnection, invalid moves)
  - Verify all requirements are met
