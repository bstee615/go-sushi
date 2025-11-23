# Requirements Document

## Introduction

This document specifies the requirements for implementing Sushi Go!, a fast-paced card drafting and set-collection game. The system shall support 2 to 5 players competing to create the best combination of sushi dishes over three rounds. The implementation consists of a Go backend server managing game state and logic, and a React frontend with Tailwind CSS, DaisyUI, and Canvas for rendering the game interface.

## Glossary

- **Game System**: The complete software implementation including backend and frontend components
- **Backend Server**: The Go-based server that manages game state, validates moves, and enforces rules
- **Frontend Client**: The React-based user interface that displays game state and accepts player input
- **Card Draft**: The process where players simultaneously select one card from their hand
- **Hand**: The collection of cards currently held by a player during a round
- **Round**: One complete cycle of card drafting until all cards are played
- **Game Session**: Three complete rounds of play
- **Score Tracker**: The visual representation of player scores using train and tray cards
- **Maki Roll**: A type of sushi card used for competitive scoring (also called Sushi Roll)
- **Nigiri**: A type of sushi card with three variants (Squid, Salmon, Egg) worth different points
- **Wasabi Card**: A modifier card that triples the points of the next Nigiri played
- **Chopsticks Card**: A special card allowing a player to select two cards in a future turn
- **Pudding Card**: A dessert card scored only at the end of the game
- **Set Collection**: Cards that score points when collected in specific quantities (Tempura, Sashimi, Dumplings)

## Requirements

### Requirement 1

**User Story:** As a player, I want to create or join a game session, so that I can play Sushi Go! with other players.

#### Acceptance Criteria

1. WHEN a player creates a new game session, THEN the Backend Server SHALL generate a unique game identifier and initialize the game state
2. WHEN a player joins an existing game session, THEN the Backend Server SHALL add the player to the session if the player count is less than 5
3. WHEN the player count reaches the minimum of 2 players, THEN the Game System SHALL allow the game to start
4. WHEN a player attempts to join a full game session, THEN the Backend Server SHALL reject the join request and return an error message
5. WHEN a game session is created, THEN the Backend Server SHALL assign each player a unique player identifier

### Requirement 2

**User Story:** As a player, I want the game to deal the correct number of cards at the start of each round, so that gameplay follows the official rules.

#### Acceptance Criteria

1. WHEN a round starts with 2 players, THEN the Backend Server SHALL deal 10 cards to each player
2. WHEN a round starts with 3 players, THEN the Backend Server SHALL deal 9 cards to each player
3. WHEN a round starts with 4 players, THEN the Backend Server SHALL deal 8 cards to each player
4. WHEN a round starts with 5 players, THEN the Backend Server SHALL deal 7 cards to each player
5. WHEN cards are dealt, THEN the Backend Server SHALL shuffle the deck before dealing

### Requirement 3

**User Story:** As a player, I want to select one card from my hand each turn, so that I can build my collection strategically.

#### Acceptance Criteria

1. WHEN it is a player's turn to select a card, THEN the Frontend Client SHALL display all cards in the player's current hand
2. WHEN a player selects a card from their hand, THEN the Frontend Client SHALL mark the card as selected and prevent further selection
3. WHEN all players have selected their cards, THEN the Backend Server SHALL reveal all selected cards simultaneously
4. WHEN cards are revealed, THEN the Backend Server SHALL add each selected card to the respective player's collection
5. WHEN a player has Chopsticks in their collection and uses them, THEN the Backend Server SHALL allow the player to select two cards instead of one

### Requirement 4

**User Story:** As a player, I want hands to be passed to the next player after each turn, so that the card drafting mechanic works correctly.

#### Acceptance Criteria

1. WHEN all players have revealed their selected cards, THEN the Backend Server SHALL pass each player's remaining hand to the player on their left
2. WHEN a hand is passed, THEN the Backend Server SHALL remove the selected card from the hand before passing
3. WHEN the last card in a hand is played, THEN the Backend Server SHALL end the current round
4. WHEN a round ends, THEN the Backend Server SHALL trigger scoring for that round

### Requirement 5

**User Story:** As a player, I want Maki Roll cards to be scored correctly at the end of each round, so that competitive scoring is accurate.

#### Acceptance Criteria

1. WHEN a round ends, THEN the Backend Server SHALL count the total Maki Roll icons for each player
2. WHEN one player has the most Maki Roll icons, THEN the Backend Server SHALL award 6 points to that player
3. WHEN multiple players tie for the most Maki Roll icons, THEN the Backend Server SHALL split 6 points evenly among them (rounded down)
4. WHEN one player has the second-most Maki Roll icons and no tie for first, THEN the Backend Server SHALL award 3 points to that player
5. WHEN multiple players tie for second-most Maki Roll icons, THEN the Backend Server SHALL split 3 points evenly among them (rounded down)

### Requirement 6

**User Story:** As a player, I want Tempura cards to be scored correctly, so that set collection rewards are accurate.

#### Acceptance Criteria

1. WHEN a round ends and a player has collected 2 Tempura cards, THEN the Backend Server SHALL award 5 points for that set
2. WHEN a round ends and a player has collected 4 Tempura cards, THEN the Backend Server SHALL award 10 points (two sets of 2)
3. WHEN a round ends and a player has an odd number of Tempura cards, THEN the Backend Server SHALL award 0 points for the unpaired card

### Requirement 7

**User Story:** As a player, I want Sashimi cards to be scored correctly, so that set collection rewards are accurate.

#### Acceptance Criteria

1. WHEN a round ends and a player has collected 3 Sashimi cards, THEN the Backend Server SHALL award 10 points for that set
2. WHEN a round ends and a player has collected 6 Sashimi cards, THEN the Backend Server SHALL award 20 points (two sets of 3)
3. WHEN a round ends and a player has fewer than 3 Sashimi cards, THEN the Backend Server SHALL award 0 points for those cards

### Requirement 8

**User Story:** As a player, I want Dumpling cards to be scored correctly with increasing value, so that collecting more dumplings is rewarded.

#### Acceptance Criteria

1. WHEN a round ends and a player has 1 Dumpling card, THEN the Backend Server SHALL award 1 point
2. WHEN a round ends and a player has 2 Dumpling cards, THEN the Backend Server SHALL award 3 points
3. WHEN a round ends and a player has 3 Dumpling cards, THEN the Backend Server SHALL award 6 points
4. WHEN a round ends and a player has 4 Dumpling cards, THEN the Backend Server SHALL award 10 points
5. WHEN a round ends and a player has 5 or more Dumpling cards, THEN the Backend Server SHALL award 15 points

### Requirement 9

**User Story:** As a player, I want Nigiri cards to be scored correctly with their base values, so that my sushi collection is properly valued.

#### Acceptance Criteria

1. WHEN a round ends and a player has a Squid Nigiri card, THEN the Backend Server SHALL award 3 points for that card
2. WHEN a round ends and a player has a Salmon Nigiri card, THEN the Backend Server SHALL award 2 points for that card
3. WHEN a round ends and a player has an Egg Nigiri card, THEN the Backend Server SHALL award 1 point for that card

### Requirement 10

**User Story:** As a player, I want Wasabi cards to triple the value of the next Nigiri I play, so that I can maximize my scoring strategy.

#### Acceptance Criteria

1. WHEN a player plays a Wasabi card, THEN the Backend Server SHALL mark that Wasabi as available for the next Nigiri
2. WHEN a player plays a Nigiri card and has an available Wasabi, THEN the Backend Server SHALL triple the Nigiri's point value
3. WHEN a Nigiri is placed on a Wasabi, THEN the Backend Server SHALL mark that Wasabi as used
4. WHEN a round ends and a Wasabi has no Nigiri on it, THEN the Backend Server SHALL award 0 points for that Wasabi
5. WHEN scoring a Squid Nigiri with Wasabi, THEN the Backend Server SHALL award 9 points instead of 3

### Requirement 11

**User Story:** As a player, I want to use Chopsticks to select two cards in a single turn, so that I can adapt my strategy during the game.

#### Acceptance Criteria

1. WHEN a player has Chopsticks in their collection and activates them, THEN the Backend Server SHALL allow the player to select two cards from their current hand
2. WHEN a player uses Chopsticks, THEN the Backend Server SHALL return the Chopsticks card to the hand being passed
3. WHEN a player selects two cards using Chopsticks, THEN the Backend Server SHALL add both cards to the player's collection
4. WHEN Chopsticks are returned to a hand, THEN the Backend Server SHALL include them in the hand passed to the next player

### Requirement 12

**User Story:** As a player, I want Pudding cards to be scored at the end of the game, so that long-term strategy is rewarded.

#### Acceptance Criteria

1. WHEN the third round ends, THEN the Backend Server SHALL count each player's total Pudding cards across all three rounds
2. WHEN the game has more than 2 players and one player has the most Pudding cards, THEN the Backend Server SHALL award 6 points to that player
3. WHEN the game has more than 2 players and multiple players tie for the most Pudding cards, THEN the Backend Server SHALL award 6 points to each tied player
4. WHEN the game has more than 2 players and one player has the fewest Pudding cards, THEN the Backend Server SHALL deduct 6 points from that player
5. WHEN the game has more than 2 players and multiple players tie for the fewest Pudding cards, THEN the Backend Server SHALL deduct 6 points from each tied player
6. WHEN the game has exactly 2 players, THEN the Backend Server SHALL ignore the penalty for fewest Pudding cards

### Requirement 13

**User Story:** As a player, I want to see my current score updated after each round, so that I can track my progress throughout the game.

#### Acceptance Criteria

1. WHEN a round ends, THEN the Frontend Client SHALL display each player's score for that round
2. WHEN a round ends, THEN the Frontend Client SHALL display each player's cumulative total score
3. WHEN the game ends, THEN the Frontend Client SHALL display the final scores including Pudding points
4. WHEN scores are displayed, THEN the Frontend Client SHALL show a breakdown of points by card type

### Requirement 14

**User Story:** As a player, I want to see all cards in my current hand clearly displayed, so that I can make informed decisions.

#### Acceptance Criteria

1. WHEN it is a player's turn, THEN the Frontend Client SHALL render all cards in the player's hand using Canvas
2. WHEN displaying cards, THEN the Frontend Client SHALL show card artwork, type, and point values
3. WHEN a player hovers over a card, THEN the Frontend Client SHALL highlight that card
4. WHEN cards are displayed, THEN the Frontend Client SHALL arrange them in a readable fan layout

### Requirement 15

**User Story:** As a player, I want to see my collected cards organized by type, so that I can easily assess my current strategy.

#### Acceptance Criteria

1. WHEN a player has collected cards, THEN the Frontend Client SHALL display them grouped by card type
2. WHEN displaying collected cards, THEN the Frontend Client SHALL show Wasabi cards with any Nigiri placed on them
3. WHEN displaying collected cards, THEN the Frontend Client SHALL keep similar cards together for easier scoring visibility

### Requirement 16

**User Story:** As a player, I want the game to automatically progress through three rounds, so that the game structure is enforced.

#### Acceptance Criteria

1. WHEN a game session starts, THEN the Backend Server SHALL initialize the round counter to 1
2. WHEN a round ends, THEN the Backend Server SHALL increment the round counter
3. WHEN the third round ends, THEN the Backend Server SHALL trigger final scoring and end the game
4. WHEN a new round starts, THEN the Backend Server SHALL clear all cards except Pudding cards from player collections

### Requirement 17

**User Story:** As a player, I want the game to determine a winner based on total score, so that the outcome is clear.

#### Acceptance Criteria

1. WHEN the game ends, THEN the Backend Server SHALL calculate each player's final score including Pudding points
2. WHEN final scores are calculated, THEN the Backend Server SHALL identify the player with the highest score as the winner
3. WHEN multiple players tie for the highest score, THEN the Backend Server SHALL use Pudding card count as the tiebreaker
4. WHEN the winner is determined, THEN the Frontend Client SHALL display the winner and final rankings

### Requirement 18

**User Story:** As a player, I want real-time updates when other players make moves, so that the game flows smoothly.

#### Acceptance Criteria

1. WHEN any player selects a card, THEN the Backend Server SHALL notify all players that a selection has been made without revealing the card
2. WHEN all players have selected cards, THEN the Backend Server SHALL broadcast the reveal to all players simultaneously
3. WHEN game state changes, THEN the Backend Server SHALL push updates to all connected Frontend Clients
4. WHEN a player disconnects, THEN the Backend Server SHALL notify other players and pause the game

### Requirement 19

**User Story:** As a developer, I want the Backend Server to validate all game moves, so that cheating is prevented and rules are enforced.

#### Acceptance Criteria

1. WHEN a player attempts to select a card not in their hand, THEN the Backend Server SHALL reject the move and return an error
2. WHEN a player attempts to select multiple cards without Chopsticks, THEN the Backend Server SHALL reject the move and return an error
3. WHEN a player attempts to make a move out of turn, THEN the Backend Server SHALL reject the move and return an error
4. WHEN invalid game state is detected, THEN the Backend Server SHALL log the error and prevent state corruption

### Requirement 20

**User Story:** As a developer, I want the game state to be serializable, so that games can be saved and restored.

#### Acceptance Criteria

1. WHEN game state is requested, THEN the Backend Server SHALL serialize the complete game state to JSON format
2. WHEN a serialized game state is provided, THEN the Backend Server SHALL restore the game to that exact state
3. WHEN serializing game state, THEN the Backend Server SHALL include all player hands, collections, scores, and round information
4. WHEN deserializing game state, THEN the Backend Server SHALL validate the state before restoring

### Requirement 21

**User Story:** As a player, I want the system to automatically generate memorable game IDs and player names, so that I can easily share and remember game sessions.

#### Acceptance Criteria

1. WHEN a new game session is created, THEN the Backend Server SHALL generate a game identifier in the format region-flower-number using famous Japanese regions and flowers
2. WHEN a game identifier is generated, THEN the Backend Server SHALL use recognizable Japanese region names such as Tokyo, Kyoto, Osaka, Hokkaido, or Okinawa
3. WHEN a game identifier is generated, THEN the Backend Server SHALL use recognizable Japanese flower names such as Sakura, Ume, Tsubaki, Ajisai, or Kiku
4. WHEN a player joins without providing a name, THEN the Backend Server SHALL assign a random name from famous sushi chefs, historical figures, or pop culture characters
5. WHEN a random player name is generated, THEN the Backend Server SHALL select from recognizable names such as Jiro Ono, Naruto, Totoro, Miyamoto Musashi, or other highly famous figures

### Requirement 22

**User Story:** As a player, I want to reconnect to my game session using my username, so that I can resume playing if I get disconnected.

#### Acceptance Criteria

1. WHEN a player joins a game with a username that already exists in that game, THEN the Backend Server SHALL load the player into their existing player state
2. WHEN a player reconnects using their username, THEN the Backend Server SHALL restore their hand, collection, score, and all game progress
3. WHEN a player reconnects, THEN the Backend Server SHALL notify other players that the player has reconnected
4. WHEN a player reconnects, THEN the Frontend Client SHALL display the current game state with all their cards and progress intact
5. WHEN a player attempts to join with a username that exists in a different game, THEN the Backend Server SHALL treat them as a new player in the new game
