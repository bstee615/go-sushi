# Requirements Document

## Introduction

The playtest runner system enables automated testing of game flows by simulating multiple WebSocket clients interacting with the backend server. Test scenarios are defined in YAML files that specify sequences of client actions and expected server responses. The runner executes these scenarios, captures state, and validates behavior.

## Glossary

- **Playtest**: A test scenario defined in YAML that simulates a sequence of client-server interactions
- **Runner**: The Go program that executes playtest scenarios against the backend server
- **Client Simulator**: A simulated WebSocket client that sends messages and receives responses
- **Variable Store**: An in-memory map that stores values from server responses for use in subsequent messages
- **Turn**: A single client action within a playtest scenario

## Requirements

### Requirement 1

**User Story:** As a developer, I want to define test scenarios in YAML files, so that I can specify game flows in a readable format without writing Go code.

#### Acceptance Criteria

1. WHEN a YAML file is parsed THEN the system SHALL load all turn definitions with client identifiers and message payloads
2. WHEN a message field contains a variable reference (e.g., `<globalGame>`) THEN the system SHALL recognize it as a placeholder for runtime substitution
3. WHEN the YAML structure is invalid THEN the system SHALL return a descriptive error indicating the parsing failure
4. THE system SHALL support message types including join_game, start_game, and select_card
5. THE system SHALL allow empty string values to indicate server-generated values (e.g., empty gameId for new game creation)

### Requirement 2

**User Story:** As a developer, I want the runner to simulate multiple WebSocket clients, so that I can test multi-player game scenarios.

#### Acceptance Criteria

1. WHEN a playtest references multiple client identifiers THEN the system SHALL create separate WebSocket connections for each unique client
2. WHEN a client sends a message THEN the system SHALL use that client's specific WebSocket connection
3. WHEN a client receives a response THEN the system SHALL store the response in that client's state
4. THE system SHALL maintain separate state for each simulated client throughout the test execution
5. WHEN a WebSocket connection fails THEN the system SHALL report the connection error and halt test execution

### Requirement 3

**User Story:** As a developer, I want to capture and store values from server responses, so that I can reference them in subsequent test steps.

#### Acceptance Criteria

1. WHEN a server response contains a gameId field THEN the system SHALL store it in the variable store with key "globalGame"
2. WHEN a message template contains a variable reference THEN the system SHALL substitute the stored value before sending
3. WHEN a variable is referenced but not found in the store THEN the system SHALL report an error indicating the missing variable
4. THE system SHALL support storing any JSON field from server responses
5. THE system SHALL maintain the variable store across all turns in a single playtest execution

### Requirement 4

**User Story:** As a developer, I want the runner to execute test scenarios sequentially, so that I can verify game state progresses correctly through each step.

#### Acceptance Criteria

1. WHEN the runner executes a playtest THEN the system SHALL process turns in the order defined in the YAML file
2. WHEN a turn completes THEN the system SHALL wait for the server response before proceeding to the next turn
3. WHEN a turn fails THEN the system SHALL halt execution and report the failure with context
4. THE system SHALL log each turn's sent message and received response for debugging
5. WHEN all turns complete successfully THEN the system SHALL report test success

### Requirement 5

**User Story:** As a developer, I want to run playtest files from the command line, so that I can integrate tests into my development workflow.

#### Acceptance Criteria

1. WHEN the runner is invoked with a test file path THEN the system SHALL load and execute that specific playtest
2. WHEN the runner is invoked with a directory path THEN the system SHALL execute all YAML files in that directory
3. WHEN execution completes THEN the system SHALL exit with code 0 for success or non-zero for failure
4. THE system SHALL print test results to stdout in a readable format
5. WHEN the backend server is not running THEN the system SHALL report a connection error with clear messaging

### Requirement 6

**User Story:** As a developer, I want the runner to save client state after each turn, so that I can inspect game progression and debug failures.

#### Acceptance Criteria

1. WHEN a client receives a game_state message THEN the system SHALL store the complete game state in that client's state
2. WHEN a turn completes THEN the system SHALL make the current client states available for inspection
3. THE system SHALL preserve all received messages in chronological order for each client
4. WHEN a test fails THEN the system SHALL output the final state of all clients
5. THE system SHALL support optional state snapshots after each turn for detailed debugging
