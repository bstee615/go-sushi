# Backend Tests

This directory contains comprehensive tests for the Sushi Go! backend implementation.

## Running Tests

### Run all tests
```bash
cd backend
go test ./...
```

### Run tests with verbose output
```bash
go test ./... -v
```

### Run tests with coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run specific test packages
```bash
# Scoring tests
go test ./scoring -v

# Server tests
go test ./server -v

# Engine tests
go test ./engine -v

# Models tests
go test ./models -v
```

### Run specific tests
```bash
# Run a specific test by name
go test ./scoring -run TestScoreMakiRolls_SingleWinner -v
```

## Test Coverage

### Scoring Tests (`scoring/scoring_comprehensive_test.go`)
- ✅ Maki Roll scoring (single winner, ties for first, ties for second, no maki)
- ✅ Tempura scoring (0-5 cards, set completion)
- ✅ Sashimi scoring (0-6 cards, set completion)
- ✅ Dumpling scoring (0-10 cards, point progression)
- ✅ Nigiri scoring (all variants: Squid, Salmon, Egg)
- ✅ Wasabi multiplier mechanics (3x nigiri points)
- ✅ Complete round scoring with multiple card types
- ✅ Edge cases (empty collection, chopsticks don't score)

### Server Tests (`server/server_test.go`)
- ✅ Server start and stop
- ✅ Health endpoint
- ✅ WebSocket connection handling
- ✅ Multiple concurrent clients
- ✅ Game creation and joining
- ✅ Game listing
- ✅ Custom game configuration
- ✅ Connection close handling
- ✅ Invalid message handling

### Engine Tests (`engine/engine_comprehensive_test.go`)
- ✅ Game creation with player limits (2-5 players)
- ✅ Joining existing games
- ✅ Player management (remove, kick)
- ✅ Game start validation
- ✅ Card dealing and hand management
- ✅ Card selection and withdrawal
- ✅ Card revealing mechanics
- ✅ Hand passing between players
- ✅ Round scoring
- ✅ Game ending and winner determination
- ✅ Concurrent access safety
- ✅ Custom game configuration

### Models Tests (`models/game_test.go`)
- ✅ Game state serialization round-trip

## Continuous Integration

Tests are automatically run on every push and pull request via GitHub Actions (`.github/workflows/go-tests.yml`).

**⚠️ PR Gating**: All tests must pass before a PR can be merged. The workflow runs tests for:
- `./scoring` - Scoring logic tests (13 tests)
- `./server` - Server integration tests (10 tests)
- `./engine` - Game engine tests (26 tests)
- `./models` - Data model tests (no tests currently)
- `./handlers` - Handler tests (no tests currently)

Note: `./playtest` is excluded from CI due to pre-existing build issues.

The workflow:
- Runs all tests with race detection enabled
- Generates coverage reports
- Uploads coverage artifacts
- Checks coverage percentage
- **Fails the PR if any tests fail**

## Test Principles

1. **Contract-based**: Tests follow the test-frontend as the source of truth for game mechanics
2. **Comprehensive**: Cover all major game mechanics including scoring, state management, and networking
3. **Isolated**: Tests are independent and can run in any order
4. **Fast**: Most tests complete in milliseconds
5. **Maintainable**: Clear test names and well-structured assertions

## Notes

- Some existing property-based tests may fail intermittently due to game state validation issues (these were pre-existing)
- Server tests use ephemeral ports (`:0`) to avoid conflicts
- WebSocket tests include proper connection cleanup
- Race detector is enabled in CI to catch concurrency issues
