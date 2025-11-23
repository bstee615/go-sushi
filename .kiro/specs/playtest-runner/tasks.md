# Implementation Plan

- [x] 1. Set up playtest runner package structure


  - Create `backend/playtest/runner` directory
  - Create main runner file with package declaration
  - Set up Go module dependencies (gorilla/websocket, gopter for property tests, yaml.v3)
  - _Requirements: All_

- [x] 2. Implement YAML parser


  - [x] 2.1 Define PlaytestDefinition and Turn structs with YAML tags


    - Create struct definitions for playtest data model
    - Add YAML parsing tags for all fields
    - _Requirements: 1.1_

  - [x] 2.2 Implement ParsePlaytest function

    - Write function to read YAML file and unmarshal into structs
    - Add error handling for file I/O and YAML parsing
    - _Requirements: 1.1, 1.3_

  - [ ]* 2.3 Write property test for YAML parsing
    - **Property 1: YAML parsing preserves structure**
    - **Validates: Requirements 1.1**

  - [ ]* 2.4 Write property test for invalid YAML handling
    - **Property 3: Invalid YAML produces errors**
    - **Validates: Requirements 1.3**

  - [ ]* 2.5 Write unit tests for YAML parser
    - Test parsing of example two-players-one-turn.yaml
    - Test various message types (join_game, start_game, select_card)
    - Test empty string handling
    - _Requirements: 1.4, 1.5_





- [ ] 3. Implement Variable Store
  - [x] 3.1 Create VariableStore struct and methods

    - Implement NewVariableStore, Set, Get methods
    - Use map[string]interface{} for storage
    - _Requirements: 3.1, 3.4_

  - [ ] 3.2 Implement variable pattern recognition
    - Write function to detect `<variableName>` pattern in strings
    - Use regex or string matching
    - _Requirements: 1.2_


  - [ ]* 3.3 Write property test for variable pattern recognition
    - **Property 2: Variable pattern recognition**
    - **Validates: Requirements 1.2**

  - [ ] 3.4 Implement Substitute method for variable replacement
    - Walk message template recursively
    - Replace variable patterns with stored values
    - Return error for missing variables
    - Preserve empty strings
    - _Requirements: 3.2, 3.3_

  - [ ]* 3.5 Write property test for JSON field extraction
    - **Property 6: JSON field extraction and storage**
    - **Validates: Requirements 3.1, 3.4**

  - [ ]* 3.6 Write property test for variable substitution
    - **Property 7: Variable substitution correctness**
    - **Validates: Requirements 3.2**

  - [ ]* 3.7 Write property test for missing variable detection
    - **Property 8: Missing variable detection**
    - **Validates: Requirements 3.3**





  - [x]* 3.8 Write unit tests for Variable Store

    - Test Set and Get operations
    - Test substitution with various templates
    - Test error cases
    - _Requirements: 3.1, 3.2, 3.3, 3.4_


- [ ] 4. Implement Client Simulator
  - [ ] 4.1 Create ClientSimulator struct
    - Define struct with ID, connection, messages, and gameState fields
    - _Requirements: 2.1, 2.4_


  - [ ] 4.2 Implement NewClientSimulator with WebSocket connection
    - Establish WebSocket connection to server
    - Handle connection errors
    - _Requirements: 2.1, 2.5_


  - [ ] 4.3 Implement SendMessage method
    - Marshal message to JSON
    - Send via WebSocket connection
    - Log sent message
    - _Requirements: 2.2, 4.4_

  - [ ] 4.4 Implement ReceiveMessage method
    - Read from WebSocket connection
    - Store received message in messages slice
    - Update gameState if message type is game_state
    - Log received message
    - _Requirements: 2.3, 4.4, 6.1, 6.3_

  - [ ] 4.5 Implement Close method
    - Close WebSocket connection gracefully
    - _Requirements: 2.1_






  - [ ]* 4.6 Write property test for client state isolation
    - **Property 5: Client state isolation**
    - **Validates: Requirements 2.4**


  - [ ]* 4.7 Write property test for client state management
    - **Property 14: Client state management**
    - **Validates: Requirements 6.1, 6.2, 6.3**



  - [ ]* 4.8 Write unit tests for Client Simulator
    - Test message sending and receiving
    - Test state storage
    - Test connection handling
    - _Requirements: 2.2, 2.3, 6.1_



- [ ] 5. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 6. Implement Test Runner
  - [ ] 6.1 Create TestRunner struct
    - Define struct with clients map, variable store, and server URL


    - _Requirements: 2.1, 3.5_

  - [ ] 6.2 Implement NewTestRunner constructor
    - Initialize clients map and variable store
    - Store server URL
    - _Requirements: 2.1, 3.5_

  - [ ] 6.3 Implement executeTurn method
    - Get or create client for turn's client ID
    - Substitute variables in message template
    - Send message via client
    - Receive and process response
    - Extract gameId from response and store as globalGame
    - Handle errors and halt on failure
    - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.2, 4.1, 4.2, 4.3_

  - [ ] 6.4 Implement RunPlaytest method
    - Parse YAML file
    - Execute turns sequentially
    - Wait for response after each turn
    - Log execution progress
    - Handle errors and report failures
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 6.5 Implement printResults method




    - Print test name and pass/fail status
    - Print execution time
    - On failure, print final state of all clients

    - _Requirements: 4.5, 6.4_

  - [ ]* 6.6 Write property test for unique client connections
    - **Property 4: Unique clients create unique connections**
    - **Validates: Requirements 2.1**


  - [ ]* 6.7 Write property test for variable persistence
    - **Property 9: Variable persistence across turns**
    - **Validates: Requirements 3.5**



  - [ ]* 6.8 Write property test for sequential execution
    - **Property 10: Sequential turn execution**
    - **Validates: Requirements 4.1, 4.2**

  - [ ]* 6.9 Write property test for turn logging
    - **Property 11: Turn logging completeness**
    - **Validates: Requirements 4.4**

  - [ ]* 6.10 Write integration test for two-players-one-turn scenario
    - Run complete playtest against test server
    - Verify all turns execute successfully
    - Verify final game state
    - _Requirements: All_

- [ ] 7. Implement CLI interface
  - [x] 7.1 Create main.go for playtest runner command




    - Parse command-line arguments (file path or directory)
    - Handle --server-url flag for backend URL


    - _Requirements: 5.1, 5.2_

  - [ ] 7.2 Implement single file execution
    - Load and run specified playtest file
    - Print results to stdout
    - Exit with appropriate code
    - _Requirements: 5.1, 5.3_






  - [ ] 7.3 Implement directory execution
    - Find all .yaml files in directory
    - Execute each playtest
    - Aggregate and report results


    - _Requirements: 5.2, 5.3_

  - [ ] 7.4 Add connection error handling
    - Detect when backend server is not running
    - Report clear error message
    - _Requirements: 5.5_

  - [ ]* 7.5 Write property test for directory batch execution
    - **Property 12: Directory batch execution**
    - **Validates: Requirements 5.2**

  - [ ]* 7.6 Write property test for exit code correctness
    - **Property 13: Exit code correctness**
    - **Validates: Requirements 5.3**

  - [ ]* 7.7 Write unit tests for CLI interface
    - Test file path handling
    - Test directory path handling
    - Test exit codes
    - _Requirements: 5.1, 5.2, 5.3_

- [ ] 8. Add optional state snapshot feature
  - [ ] 8.1 Add --verbose flag to CLI
    - When enabled, print state snapshot after each turn
    - _Requirements: 6.5_

  - [ ] 8.2 Implement state snapshot formatting
    - Format client states in readable JSON
    - Include all messages and current game state
    - _Requirements: 6.2, 6.5_

  - [ ]* 8.3 Write unit tests for state snapshot feature
    - Test snapshot formatting
    - Test verbose flag behavior
    - _Requirements: 6.5_

- [ ] 9. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 10. Create documentation and examples
  - [ ] 10.1 Create README.md for playtest runner
    - Document usage and command-line options
    - Provide examples of YAML playtest files
    - Explain variable substitution syntax
    - _Requirements: All_

  - [ ] 10.2 Add additional example playtest files
    - Create example for three-player game
    - Create example for full round completion
    - Create example demonstrating variable usage
    - _Requirements: All_
