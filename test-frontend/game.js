// Game state
let ws = null;
let gameState = null;
let myPlayerId = null;
let selectedCardIndex = null;
let secondCardIndex = null;
let chopsticksMode = false;
let previousHandSize = 0;
let previousRound = 0;
let isFirstDeal = true;
let isAnimating = false;

// UI Elements
const connectionStatus = document.getElementById('connectionStatus');
const createBtn = document.getElementById('createBtn');
const joinBtn = document.getElementById('joinBtn');
const startBtn = document.getElementById('startBtn');
const handDiv = document.getElementById('hand');
const playersListDiv = document.getElementById('playersList');
const collectionDiv = document.getElementById('collection');
// Log div removed from UI but keeping function for debugging
const logDiv = { appendChild: () => {}, scrollTop: 0, scrollHeight: 0 };
const roundIndicator = document.getElementById('roundIndicator');
const roundDots = document.getElementById('roundDots');
const turnIndicator = document.getElementById('turnIndicator');
const turnDots = document.getElementById('turnDots');
const handStatusMessage = document.getElementById('handStatusMessage');
const currentGameIdDiv = document.getElementById('currentGameId');
const gameIdDisplay = document.getElementById('gameIdDisplay');
const loginScreen = document.getElementById('loginScreen');
const playingScreen = document.getElementById('playingScreen');
const gamesListDiv = document.getElementById('gamesList');

// Logging
function log(message, type = 'info') {
    const entry = document.createElement('div');
    entry.className = `log-entry log-${type}`;
    const timestamp = new Date().toLocaleTimeString();
    entry.textContent = `[${timestamp}] ${message}`;
    logDiv.appendChild(entry);
    logDiv.scrollTop = logDiv.scrollHeight;
}

// Connection management
function connect() {
    const serverUrl = document.getElementById('serverUrl').value;
    
    // Show connecting spinner
    connectionStatus.innerHTML = '<span class="spinner"></span> Connecting...';
    connectionStatus.className = 'connection-status status-connecting';
    
    try {
        ws = new WebSocket(serverUrl);
        
        ws.onopen = () => {
            log('Connected to server', 'sent');
            // Hide connection status once connected
            connectionStatus.style.display = 'none';
            createBtn.disabled = false;
            joinBtn.disabled = false;
            
            // Request list of games
            requestGamesList();
        };
        
        ws.onclose = () => {
            log('Disconnected from server', 'error');
            // Show connecting spinner again
            connectionStatus.style.display = 'inline-block';
            connectionStatus.innerHTML = '<span class="spinner"></span> Connecting...';
            connectionStatus.className = 'connection-status status-connecting';
            createBtn.disabled = true;
            joinBtn.disabled = true;
            startBtn.disabled = true;
            
            // Try to reconnect after 2 seconds
            setTimeout(() => {
                if (!ws || ws.readyState === WebSocket.CLOSED) {
                    connect();
                }
            }, 2000);
        };
        
        ws.onerror = (error) => {
            log(`WebSocket error: ${error}`, 'error');
        };
        
        ws.onmessage = (event) => {
            handleMessage(event.data);
        };
        
    } catch (error) {
        log(`Connection error: ${error.message}`, 'error');
    }
}

function logout() {
    // Don't close WebSocket connection - keep it alive
    
    // Reset game state
    gameState = null;
    myPlayerId = null;
    selectedCardIndex = null;
    secondCardIndex = null;
    chopsticksMode = false;
    previousHandSize = 0;
    previousRound = 0;
    isFirstDeal = true;
    
    // Animate out playing screen
    playingScreen.classList.add('fade-out');
    
    setTimeout(() => {
        // Switch to login screen
        playingScreen.style.display = 'none';
        playingScreen.classList.remove('fade-out', 'screen');
        loginScreen.style.display = 'block';
        // Trigger fade-in animation for login screen
        loginScreen.classList.add('screen');
        
        // Clear UI
        handDiv.innerHTML = '';
        playersListDiv.innerHTML = '';
        collectionDiv.innerHTML = '';
    }, 300);
    
    log('Logged out', 'info');
}

function switchToPlayingScreen() {
    // Animate out login screen
    loginScreen.classList.add('fade-out');
    
    setTimeout(() => {
        loginScreen.style.display = 'none';
        loginScreen.classList.remove('fade-out');
        playingScreen.style.display = 'block';
        // Trigger fade-in animation for playing screen
        playingScreen.classList.add('screen');
    }, 300);
}

// Message handling
function sendMessage(type, payload) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        log('Cannot send message: not connected', 'error');
        return;
    }
    
    const message = {
        type: type,
        payload: payload
    };
    
    ws.send(JSON.stringify(message));
    log(`Sent: ${type} ${JSON.stringify(payload)}`, 'sent');
}

function handleMessage(data) {
    try {
        const message = JSON.parse(data);
        log(`Received: ${message.type}`, 'received');
        
        switch (message.type) {
            case 'game_state':
                handleGameState(message.payload);
                break;
            case 'card_revealed':
                log(`Cards revealed: ${JSON.stringify(message.payload)}`, 'received');
                break;
            case 'round_end':
                handleRoundEnd(message.payload);
                log(`Round ended: ${JSON.stringify(message.payload)}`, 'received');
                break;
            case 'game_end':
                handleGameEnd(message.payload);
                log(`Game ended: ${JSON.stringify(message.payload)}`, 'received');
                break;
            case 'list_games':
                handleGamesList(message.payload);
                break;
            case 'game_deleted':
                handleGameDeleted(message.payload);
                break;
            case 'player_kicked':
                handlePlayerKicked(message.payload);
                break;
            case 'error':
                log(`Error: ${message.payload.message || JSON.stringify(message.payload)}`, 'error');
                break;
            default:
                log(`Unknown message type: ${message.type}`, 'received');
        }
    } catch (error) {
        log(`Error parsing message: ${error.message}`, 'error');
    }
}

function handleGameState(payload) {
    // NOTE: Backend sends snake_case field names (pudding_cards, has_chopsticks, etc.)
    // Frontend must use snake_case when accessing these fields from the payload
    
    const previousHand = gameState?.myHand;
    const newHand = payload.myHand;
    const previousSelected = gameState?.players?.find(p => p.id === myPlayerId)?.hasSelected;
    const currentSelected = payload.players?.find(p => p.id === myPlayerId)?.hasSelected;
    
    // Detect animation type
    const currentRound = payload.currentRound || payload.current_round || 0;
    const currentHandSize = newHand?.length || 0;
    
    // Determine if this is a new deal (start of round) or turn transition (hand passed)
    let animationType = null;
    let needsSlideOut = false;
    
    if (currentRound !== previousRound && currentRound > 0) {
        // New round - deal animation
        animationType = 'deal';
        isFirstDeal = false;
    } else if (previousHandSize > 0 && currentHandSize > 0 && currentHandSize !== previousHandSize) {
        // Hand passed - turn animation with slide out first
        animationType = 'turn';
        needsSlideOut = true;
    } else if (currentHandSize > 0 && previousHandSize === 0 && !isFirstDeal) {
        // First hand of a new round
        animationType = 'deal';
    }
    
    // Clear selections when hand changes (before any animations)
    if (previousHand && newHand && JSON.stringify(previousHand) !== JSON.stringify(newHand)) {
        selectedCardIndex = null;
        secondCardIndex = null;
        chopsticksMode = false;
    }
    
    // If player withdrew their card (was selected, now not selected), clear UI selection
    if (previousSelected && !currentSelected) {
        selectedCardIndex = null;
        secondCardIndex = null;
    }
    
    // If we need to slide out old cards first, do that before updating state
    if (needsSlideOut && previousHand && previousHand.length > 0) {
        // Show slide-out animation with old hand
        updateHandWithSlideOut(previousHand, () => {
            // After slide-out completes, update state and show new cards
            previousHandSize = currentHandSize;
            previousRound = currentRound;
            gameState = payload;
            
            // Continue with normal state update
            continueStateUpdate(payload, animationType);
        });
        return; // Exit early, will continue after animation
    }
    
    // Update tracking variables
    previousHandSize = currentHandSize;
    previousRound = currentRound;
    
    gameState = payload;
    
    // Continue with normal state update
    continueStateUpdate(payload, animationType);
}

function continueStateUpdate(payload, animationType) {
    // Get my player ID from the payload
    if (payload.myPlayerId) {
        myPlayerId = payload.myPlayerId;
        
        // Update player name in the UI if it was randomly generated
        const myPlayer = payload.players?.find(p => p.id === myPlayerId);
        if (myPlayer && myPlayer.name) {
            const playerNameInput = document.getElementById('playerName');
            // Only update if the input is empty (meaning it was randomly generated)
            if (!playerNameInput.value || playerNameInput.value === '') {
                playerNameInput.value = myPlayer.name;
                log(`Your name: ${myPlayer.name}`, 'info');
            }
        }
    }
    
    // Update game ID display
    if (payload.gameId) {
        document.getElementById('gameId').value = payload.gameId;
        gameIdDisplay.textContent = payload.gameId;
        currentGameIdDiv.style.display = 'block';
    }
    
    // Show/hide panels based on game phase
    const waitingMessage = document.getElementById('waitingMessage');
    const handContent = document.getElementById('handContent');
    const collectionPanel = document.getElementById('collectionPanel');
    
    if (gameState.phase === 'waiting') {
        // Show waiting message, hide game content
        waitingMessage.style.display = 'block';
        handContent.style.display = 'none';
        collectionPanel.style.display = 'none';
    } else {
        // Hide waiting message, show game content
        waitingMessage.style.display = 'none';
        handContent.style.display = 'block';
        collectionPanel.style.display = 'block';
    }
    
    // Update UI with animation type
    updatePhaseIndicator();
    updateHand(animationType);
    updatePlayersList();
    updateCollection();
    
    // Enable start button if we're waiting for players and have at least 2 players
    if (gameState.phase === 'waiting' && gameState.players && gameState.players.length >= 2) {
        startBtn.disabled = false;
    } else if (gameState.phase === 'waiting') {
        startBtn.disabled = true;
        log(`Waiting for more players (${gameState.players ? gameState.players.length : 0}/2)`, 'info');
    } else {
        startBtn.disabled = true;
    }
}

// Handle round end message
function handleRoundEnd(payload) {
    // payload.round is the round that just completed
    const round = payload.round || 0;
    
    // Create round end overlay
    const overlay = document.createElement('div');
    overlay.className = 'round-end-overlay';
    overlay.innerHTML = `
        <div class="round-end-content">
            <div class="round-end-title">🍣 Round ${round} Completed! 🍣</div>
            <div class="round-end-subtitle">Cooking up the next round...</div>
        </div>
    `;
    
    document.body.appendChild(overlay);
    
    // Fade in
    setTimeout(() => {
        overlay.classList.add('visible');
    }, 10);
    
    // Remove after 3 seconds
    setTimeout(() => {
        overlay.classList.remove('visible');
        setTimeout(() => {
            overlay.remove();
        }, 500);
    }, 3000);
}

// Handle game end message
function handleGameEnd(payload) {
    // Hide hand and collection panels
    const handContent = document.getElementById('handContent');
    const collectionPanel = document.getElementById('collectionPanel');
    if (handContent) handContent.style.display = 'none';
    if (collectionPanel) collectionPanel.style.display = 'none';
    
    // Show final scores in the hand panel
    const handPanel = document.getElementById('handPanel');
    if (!handPanel) return;
    
    // Sort players by score (descending)
    const sortedPlayers = [...(gameState?.players || [])].sort((a, b) => b.score - a.score);
    
    const finalScoresHTML = `
        <div style="text-align: center; padding: 20px;">
            <div style="font-size: 48px; margin-bottom: 20px;">🎉</div>
            <div style="font-size: 32px; font-weight: bold; color: #333; margin-bottom: 30px;">Game Over!</div>
            
            <div style="background: white; border-radius: 10px; padding: 20px; margin-bottom: 20px;">
                <h3 style="color: #333; margin-bottom: 20px;">Final Scores</h3>
                ${sortedPlayers.map((player, index) => {
                    const isWinner = index === 0;
                    const isMe = player.id === myPlayerId;
                    return `
                        <div style="
                            background: ${isWinner ? 'linear-gradient(135deg, #ffd700 0%, #ffed4e 100%)' : '#f8f9fa'};
                            padding: 15px;
                            margin-bottom: 10px;
                            border-radius: 8px;
                            display: flex;
                            justify-content: space-between;
                            align-items: center;
                            ${isWinner ? 'border: 3px solid #ffa500;' : ''}
                        ">
                            <div style="display: flex; align-items: center; gap: 10px;">
                                <span style="font-size: 24px; font-weight: bold; color: #666;">#${index + 1}</span>
                                <span style="font-size: 18px; font-weight: ${isWinner ? 'bold' : '600'}; color: #333;">
                                    ${player.name}${isMe ? ' (You)' : ''}
                                    ${isWinner ? ' 👑' : ''}
                                </span>
                            </div>
                            <span style="font-size: 24px; font-weight: bold; color: #333;">${player.score}</span>
                        </div>
                    `;
                }).join('')}
            </div>
            
            <button onclick="location.reload()" style="
                padding: 15px 40px;
                font-size: 18px;
                font-weight: bold;
                background: #667eea;
                color: white;
                border: none;
                border-radius: 8px;
                cursor: pointer;
                transition: background 0.2s;
            " onmouseover="this.style.background='#5568d3'" onmouseout="this.style.background='#667eea'">
                🔄 Play Again
            </button>
        </div>
    `;
    
    // Replace hand panel content with final scores
    const waitingMessage = document.getElementById('waitingMessage');
    if (waitingMessage) {
        waitingMessage.innerHTML = finalScoresHTML;
        waitingMessage.style.display = 'block';
    }
}

// Helper function to show slide-out animation before updating to new cards
function updateHandWithSlideOut(oldHand, callback) {
    const handDiv = document.getElementById('hand');
    
    // Find the existing cards wrapper
    const existingWrapper = Array.from(handDiv.children).find(
        child => child.style.cssText.includes('position: relative; overflow: hidden')
    );
    
    if (!existingWrapper) {
        // No wrapper found, just call callback
        callback();
        return;
    }
    
    // Capture wrapper height to prevent jumping
    const wrapperHeight = existingWrapper.offsetHeight;
    existingWrapper.style.minHeight = `${wrapperHeight}px`;
    
    // Create a temporary container with old cards
    const tempContainer = document.createElement('div');
    tempContainer.className = 'hand turn-out-animation';
    
    // Render old cards
    oldHand.forEach((card) => {
        const cardEl = document.createElement('div');
        cardEl.className = 'card';
        cardEl.innerHTML = `
            <div class="card-type">${formatCardType(card.type)}</div>
            ${card.variant ? `<div class="card-variant">${card.variant}</div>` : ''}
            ${card.value ? `<div class="card-variant">Value: ${card.value}</div>` : ''}
        `;
        tempContainer.appendChild(cardEl);
    });
    
    // Replace wrapper contents with animating cards
    existingWrapper.innerHTML = '';
    existingWrapper.appendChild(tempContainer);
    
    // After animation completes, call callback to show new cards
    setTimeout(() => {
        existingWrapper.style.minHeight = '';
        callback();
    }, 400); // Match the slide-out animation duration
}

// Game actions
function createGame() {
    const playerName = document.getElementById('playerName').value;
    
    // Player name is now optional - backend will generate one if empty
    if (playerName) {
        log(`Creating new game as ${playerName}...`, 'info');
    } else {
        log('Creating new game with random name...', 'info');
    }
    
    // Send join_game with empty gameId to create a new game
    sendMessage('join_game', {
        gameId: '',
        playerName: playerName
    });
    
    // Switch to playing screen
    switchToPlayingScreen();
}

function joinGame() {
    const playerName = document.getElementById('playerName').value;
    const gameId = document.getElementById('gameId').value;
    
    if (!gameId) {
        log('Please enter a game ID to join, or click "Create New Game"', 'error');
        return;
    }
    
    // Player name is now optional - backend will generate one if empty
    if (playerName) {
        log(`Joining game ${gameId} as ${playerName}...`, 'info');
    } else {
        log(`Joining game ${gameId} with random name...`, 'info');
    }
    
    sendMessage('join_game', {
        gameId: gameId,
        playerName: playerName
    });
    
    // Switch to playing screen
    switchToPlayingScreen();
}

function joinGameById(gameId) {
    const playerName = document.getElementById('playerName').value;
    
    // Player name is now optional - backend will generate one if empty
    if (playerName) {
        log(`Joining game ${gameId} as ${playerName}...`, 'info');
    } else {
        log(`Joining game ${gameId} with random name...`, 'info');
    }
    
    sendMessage('join_game', {
        gameId: gameId,
        playerName: playerName
    });
    
    // Switch to playing screen
    switchToPlayingScreen();
}

function startGame() {
    if (!gameState || !gameState.gameId) {
        log('No active game', 'error');
        return;
    }
    sendMessage('start_game', { gameId: gameState.gameId });
}

function copyGameId() {
    const gameId = gameIdDisplay.textContent;
    navigator.clipboard.writeText(gameId).then(() => {
        log('Game ID copied to clipboard!', 'info');
    }).catch(err => {
        log('Failed to copy game ID', 'error');
    });
}

function kickPlayer(playerId) {
    if (!gameState || !gameState.gameId) {
        log('No active game', 'error');
        return;
    }
    
    if (gameState.phase !== 'waiting') {
        log('Can only kick players before game starts', 'error');
        return;
    }
    
    sendMessage('kick_player', { playerId: playerId });
    log(`Kicking player ${playerId}...`, 'info');
}

function requestGamesList() {
    sendMessage('list_games', {});
}

function deleteGame(gameId) {
    if (confirm(`Are you sure you want to delete game ${gameId}?`)) {
        sendMessage('delete_game', { gameId: gameId });
        log(`Deleting game ${gameId}...`, 'info');
    }
}

function handleGamesList(payload) {
    const games = payload.games || [];
    
    if (games.length === 0) {
        gamesListDiv.innerHTML = '<p style="color: #999; text-align: center;">No active games</p>';
        return;
    }
    
    gamesListDiv.innerHTML = '';
    games.forEach(game => {
        const gameItem = document.createElement('div');
        gameItem.style.cssText = 'background: #f8f9fa; padding: 12px; margin-bottom: 8px; border-radius: 6px; display: flex; justify-content: space-between; align-items: center;';
        
        gameItem.innerHTML = `
            <div>
                <div style="font-weight: bold; color: #333;">${game.id}</div>
                <div style="font-size: 12px; color: #666;">Players: ${game.playerCount} | Phase: ${game.phase}</div>
            </div>
            <div style="display: flex; gap: 8px;">
                <button onclick="joinGameById('${game.id}')" style="padding: 6px 12px; font-size: 12px; background: #667eea; color: white; border: none; border-radius: 4px; cursor: pointer;">Join</button>
                <button onclick="deleteGame('${game.id}')" style="padding: 6px 10px; font-size: 12px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer;">✕</button>
            </div>
        `;
        
        gamesListDiv.appendChild(gameItem);
    });
}

function handleGameDeleted(payload) {
    const message = payload.message || 'This game has been deleted';
    
    // Show alert
    alert(message);
    
    // Return to login screen
    returnToLogin();
}

function handlePlayerKicked(payload) {
    const message = payload.message || 'You have been kicked from the game';
    
    // Show alert
    alert(message);
    
    // Return to login screen
    returnToLogin();
}

function returnToLogin() {
    // Reset game state
    gameState = null;
    myPlayerId = null;
    selectedCardIndex = null;
    secondCardIndex = null;
    chopsticksMode = false;
    previousHandSize = 0;
    previousRound = 0;
    isFirstDeal = true;
    
    // Animate out playing screen
    playingScreen.classList.add('fade-out');
    
    setTimeout(() => {
        // Switch to login screen
        playingScreen.style.display = 'none';
        playingScreen.classList.remove('fade-out', 'screen');
        loginScreen.style.display = 'block';
        loginScreen.classList.add('screen');
        
        // Clear UI
        handDiv.innerHTML = '';
        playersListDiv.innerHTML = '';
        collectionDiv.innerHTML = '';
        
        // Clear inputs
        document.getElementById('playerName').value = '';
        document.getElementById('gameId').value = '';
        
        // Refresh games list
        requestGamesList();
    }, 300);
    
    log('Returned to login screen', 'info');
}

function toggleChopsticks() {
    chopsticksMode = !chopsticksMode;
    
    // Reset selections when toggling
    if (!chopsticksMode) {
        selectedCardIndex = null;
        secondCardIndex = null;
    }
    
    updateHand();
    log(chopsticksMode ? 'Chopsticks mode enabled - select 2 cards' : 'Chopsticks mode disabled', 'info');
}

function selectCard(index) {
    if (!gameState || gameState.phase !== 'selecting') {
        log('Cannot select card: not in selection phase', 'error');
        return;
    }
    
    // Prevent selection during animations
    if (isAnimating) {
        return;
    }
    
    // Check if player has already selected
    const myPlayer = gameState.players?.find(p => p.id === myPlayerId);
    if (myPlayer?.hasSelected) {
        log('You have already selected a card. Wait for other players.', 'info');
        return;
    }
    
    if (chopsticksMode) {
        // In chopsticks mode, allow selecting/deselecting two cards
        if (index === selectedCardIndex) {
            // Deselect first card
            selectedCardIndex = null;
            updateHand();
            log('First card deselected', 'info');
        } else if (index === secondCardIndex) {
            // Deselect second card
            secondCardIndex = null;
            updateHand();
            log('Second card deselected', 'info');
        } else if (selectedCardIndex === null) {
            // Select first card
            selectedCardIndex = index;
            updateHand();
            log('First card selected. Select a second card.', 'info');
        } else if (secondCardIndex === null) {
            // Select second card and auto-submit
            secondCardIndex = index;
            
            // Immediately send both selections
            sendMessage('select_card', {
                cardIndex: selectedCardIndex,
                useChopsticks: true,
                secondCardIndex: secondCardIndex
            });
            
            log(`Played cards at indices ${selectedCardIndex} and ${secondCardIndex} using chopsticks`, 'sent');
            updateHand();
        } else {
            // Both already selected, replace one
            log('Both cards already selected. Click a selected card to deselect it.', 'info');
        }
    } else {
        // Normal mode - auto-submit single card
        selectedCardIndex = index;
        secondCardIndex = null;
        
        // Immediately send the selection
        sendMessage('select_card', {
            cardIndex: selectedCardIndex,
            useChopsticks: false
        });
        
        log(`Played card at index ${selectedCardIndex}`, 'sent');
        updateHand();
    }
}

function withdrawCard() {
    if (!gameState || !gameState.gameId) {
        log('No active game', 'error');
        return;
    }
    
    sendMessage('withdraw_card', {
        gameId: gameState.gameId
    });
    
    log('Withdrawing card selection...', 'info');
}

function playCards() {
    if (!gameState || gameState.phase !== 'selecting') {
        log('Cannot play cards: not in selection phase', 'error');
        return;
    }
    
    if (selectedCardIndex === null) {
        log('Please select a card first', 'error');
        return;
    }
    
    if (chopsticksMode) {
        // Playing with chopsticks
        if (secondCardIndex === null) {
            log('Please select a second card for chopsticks', 'error');
            return;
        }
        
        sendMessage('select_card', {
            cardIndex: selectedCardIndex,
            useChopsticks: true,
            secondCardIndex: secondCardIndex
        });
        
        log(`Played cards at indices ${selectedCardIndex} and ${secondCardIndex} using chopsticks`, 'sent');
    } else {
        // Playing single card
        sendMessage('select_card', {
            cardIndex: selectedCardIndex,
            useChopsticks: false
        });
        
        log(`Played card at index ${selectedCardIndex}`, 'sent');
    }
    
    // Reset selections
    selectedCardIndex = null;
    secondCardIndex = null;
    chopsticksMode = false;
    updateHand();
}

function clearSelection() {
    selectedCardIndex = null;
    secondCardIndex = null;
    updateHand();
    log('Selection cleared', 'info');
}

// UI Updates
function updatePhaseIndicator() {
    if (!gameState) {
        roundDots.innerHTML = '';
        turnDots.innerHTML = '';
        roundIndicator.style.display = 'none';
        turnIndicator.style.display = 'none';
        return;
    }
    
    const round = gameState.currentRound || gameState.current_round || 0;
    
    // Only show indicators if game has started (round > 0)
    if (round === 0) {
        roundIndicator.style.display = 'none';
        turnIndicator.style.display = 'none';
        return;
    }
    
    // Show round indicator
    roundIndicator.style.display = 'flex';
    
    // Calculate turn number (each round has multiple turns as hands are passed)
    const myPlayer = gameState.players?.find(p => p.id === myPlayerId);
    const handSize = myPlayer?.handSize || 0;
    
    // Create round numbers (3 rounds total)
    const roundNumbersArray = [];
    for (let i = 1; i <= 3; i++) {
        let color, fontWeight;
        if (i < round) {
            color = '#4caf50'; // Green - Completed round
            fontWeight = 'bold';
        } else if (i === round) {
            color = '#2196F3'; // Blue - Current round
            fontWeight = 'bold';
        } else {
            color = '#999'; // Gray - Future round
            fontWeight = 'normal';
        }
        roundNumbersArray.push(`<span style="color: ${color}; font-weight: ${fontWeight}; font-size: 18px;">${i}</span>`);
    }
    roundDots.innerHTML = roundNumbersArray.join(' ');
    
    // Create turn numbers (10 turns per round, based on cards remaining)
    if (round > 0 && round <= 3) {
        const currentTurn = handSize > 0 ? 11 - handSize : 10;
        const turnNumbersArray = [];
        for (let i = 1; i <= 10; i++) {
            let color, fontWeight;
            if (i < currentTurn) {
                color = '#4caf50'; // Green - Completed turn
                fontWeight = 'bold';
            } else if (i === currentTurn) {
                color = '#2196F3'; // Blue - Current turn
                fontWeight = 'bold';
            } else {
                color = '#999'; // Gray - Future turn
                fontWeight = 'normal';
            }
            turnNumbersArray.push(`<span style="color: ${color}; font-weight: ${fontWeight}; font-size: 16px;">${i}</span>`);
        }
        turnDots.innerHTML = turnNumbersArray.join(' ');
        turnIndicator.style.display = 'flex';
    } else {
        turnDots.innerHTML = '';
        turnIndicator.style.display = 'none';
    }
}

function updateStatusMessage(text, bgColor = '#e3f2fd', textColor = '#1565c0') {
    if (!handStatusMessage) return;
    
    if (text) {
        handStatusMessage.textContent = text;
        handStatusMessage.style.background = bgColor;
        handStatusMessage.style.color = textColor;
        handStatusMessage.style.visibility = 'visible';
    } else {
        handStatusMessage.textContent = '';
        handStatusMessage.style.visibility = 'hidden';
    }
}

function updateHand(animationType = null) {
    handDiv.innerHTML = '';
    
    // Remove any existing animation classes
    handDiv.className = '';
    
    if (!gameState || !myPlayerId) {
        handDiv.innerHTML = '<p style="color: #666;">Join a game to see your hand</p>';
        return;
    }
    
    // Get my hand from the game state (server sends it in myHand field)
    const myHand = gameState.myHand || [];
    
    if (myHand.length === 0) {
        handDiv.innerHTML = '<p style="color: #666;">No cards in hand</p>';
        return;
    }
    
    // Check if player has chopsticks and if they've already selected
    const myPlayer = gameState.players?.find(p => p.id === myPlayerId);
    // NOTE: Backend sends hasChopsticks (camelCase) which indicates if chopsticks are available to use (not just in collection)
    const canUseChopsticks = !!(myPlayer?.hasChopsticks);
    const hasSelected = myPlayer?.hasSelected;
    
    // Update status message based on game state
    if (hasSelected && gameState.phase === 'selecting') {
        updateStatusMessage('⏳ Waiting for other players...', '#fff3cd', '#856404');
    } else if (gameState.phase === 'selecting' && !hasSelected) {
        if (chopsticksMode) {
            if (selectedCardIndex === null) {
                updateStatusMessage('📌 Select your first card (click to deselect)', '#e3f2fd', '#1565c0');
            } else if (secondCardIndex === null) {
                updateStatusMessage('📌 Select your second card (auto-submits when both selected)', '#e3f2fd', '#1565c0');
            } else {
                updateStatusMessage('✓ Both cards selected - submitting...', '#4caf50', 'white');
            }
        } else if (selectedCardIndex !== null) {
            updateStatusMessage('👆 Click selected card to withdraw', '#fff3cd', '#856404');
        } else {
            updateStatusMessage('👆 Click a card to play it', '#e3f2fd', '#1565c0');
        }
    } else {
        // Show transition message when not in selecting phase
        updateStatusMessage('🔄 Moving to the next turn...', '#e0e0e0', '#666');
    }
    
    // Chopsticks toggle is now in the stats bar below
    // Create a wrapper for the animated cards container
    const cardsWrapper = document.createElement('div');
    cardsWrapper.style.cssText = 'position: relative; overflow: hidden;';
    
    // Create a container for cards
    const cardsContainer = document.createElement('div');
    cardsContainer.className = 'hand';
    
    // Add animation class to container based on type
    if (animationType === 'deal') {
        cardsContainer.classList.add('deal-animation');
    } else if (animationType === 'turn') {
        cardsContainer.classList.add('turn-animation');
    }
    
    // Count cards in collection for set completion indicators
    const collectionCounts = {};
    let hasUnusedWasabi = false;
    
    if (myPlayer && myPlayer.collection) {
        let wasabiCount = 0;
        myPlayer.collection.forEach(collCard => {
            collectionCounts[collCard.type] = (collectionCounts[collCard.type] || 0) + 1;
            if (collCard.type === 'wasabi') wasabiCount++;
            if (collCard.type === 'nigiri') wasabiCount--;
        });
        hasUnusedWasabi = wasabiCount > 0;
    }
    
    // Display actual cards
    myHand.forEach((card, index) => {
        const cardEl = document.createElement('div');
        cardEl.className = 'card';
        cardEl.style.position = 'relative';
        
        // Determine if this card will complete a set or get a bonus
        let badge = null;
        if (card.type === 'tempura' && (collectionCounts['tempura'] || 0) % 2 === 1) {
            badge = { text: '✓5pts!', color: '#ffd700' };
        } else if (card.type === 'sashimi' && (collectionCounts['sashimi'] || 0) % 3 === 2) {
            badge = { text: '✓10pts!', color: '#ffd700' };
        } else if (card.type === 'nigiri' && hasUnusedWasabi) {
            badge = { text: '3x!', color: '#4caf50' };
        }
        
        // Grey out non-selected cards if player has already selected
        if (hasSelected) {
            cardEl.style.opacity = '0.3';
            cardEl.style.pointerEvents = 'none';
        }
        
        // Highlight selected cards
        if (selectedCardIndex === index || secondCardIndex === index) {
            cardEl.className += ' selected';
            if (hasSelected) {
                cardEl.style.opacity = '1';
                cardEl.style.pointerEvents = 'auto';
                cardEl.style.cursor = 'pointer';
                
                // Add withdraw tooltip on hover
                const withdrawTooltip = document.createElement('div');
                withdrawTooltip.style.cssText = `
                    position: absolute;
                    top: -30px;
                    left: 50%;
                    transform: translateX(-50%);
                    background: #dc3545;
                    color: white;
                    padding: 5px 10px;
                    border-radius: 4px;
                    font-size: 12px;
                    font-weight: bold;
                    white-space: nowrap;
                    opacity: 0;
                    transition: opacity 0.2s;
                    pointer-events: none;
                `;
                withdrawTooltip.textContent = 'Withdraw?';
                cardEl.style.position = 'relative';
                cardEl.appendChild(withdrawTooltip);
                
                cardEl.onmouseover = () => {
                    withdrawTooltip.style.opacity = '1';
                };
                cardEl.onmouseout = () => {
                    withdrawTooltip.style.opacity = '0';
                };
                
                cardEl.onclick = () => {
                    withdrawCard();
                };
            }
        } else {
            cardEl.onclick = () => selectCard(index);
        }
        
        cardEl.innerHTML = `
            <div class="card-type">${formatCardType(card.type)}</div>
            ${card.variant ? `<div class="card-variant">${card.variant}</div>` : ''}
            ${card.value ? `<div class="card-variant">Value: ${card.value}</div>` : ''}
        ` + (cardEl.innerHTML || '');
        
        // Add badge if applicable
        if (badge) {
            const badgeEl = document.createElement('div');
            badgeEl.style.cssText = `
                position: absolute;
                top: -8px;
                right: -8px;
                background: ${badge.color};
                color: #333;
                padding: 4px 8px;
                border-radius: 12px;
                font-size: 11px;
                font-weight: bold;
                box-shadow: 0 2px 4px rgba(0,0,0,0.3);
                z-index: 10;
            `;
            badgeEl.textContent = badge.text;
            cardEl.appendChild(badgeEl);
        }
        
        cardsContainer.appendChild(cardEl);
    });
    
    cardsWrapper.appendChild(cardsContainer);
    handDiv.appendChild(cardsWrapper);
    
    // Add stats bar showing card counts
    if (myPlayer && myPlayer.collection) {
        const statsBar = document.createElement('div');
        statsBar.style.cssText = 'margin-top: 20px; padding: 15px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border-radius: 8px; display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px;';
        
        // Count cards and check for active wasabi
        let makiCount = 0;
        let tempuraCount = 0;
        let sashimiCount = 0;
        let dumplingCount = 0;
        let wasabiCount = 0;
        let nigiriCount = 0;
        
        myPlayer.collection.forEach(card => {
            if (card.type === 'maki_roll') makiCount += card.value || 0;
            if (card.type === 'tempura') tempuraCount++;
            if (card.type === 'sashimi') sashimiCount++;
            if (card.type === 'dumpling') dumplingCount++;
            if (card.type === 'wasabi') wasabiCount++;
            if (card.type === 'nigiri') nigiriCount++;
        });
        
        const puddingCount = myPlayer.puddingCards ? myPlayer.puddingCards.length : 0;
        
        // Check if wasabi is active (more wasabi than nigiri means unused wasabi)
        const hasActiveWasabi = wasabiCount > nigiriCount;
        const wasabiStatus = hasActiveWasabi ? 'ACTIVE' : (wasabiCount > 0 ? 'USED' : 'NONE');
        
        // Check chopsticks status
        const chopsticksStatus = canUseChopsticks ? (chopsticksMode ? 'ACTIVE' : 'READY') : 'NONE';
        const hasChopsticksAvailable = canUseChopsticks && gameState.phase === 'selecting' && !hasSelected;
        
        // Create stat items - 3 per row for consistent layout
        const stats = [
            { emoji: '�', llabel: 'Maki', count: makiCount, color: '#e3f2fd' },
            { emoji: '🍤', label: 'Tempura', count: tempuraCount, color: '#fff3e0' },
            { emoji: '🐟', label: 'Sashimi', count: sashimiCount, color: '#fce4ec' },
            { emoji: '🥟', label: 'Dumpling', count: dumplingCount, color: '#f3e5f5' },
            { emoji: '🍮', label: 'Pudding', count: puddingCount, color: '#ffb74d' },
            { emoji: '🟢', label: 'Wasabi', count: wasabiStatus, color: hasActiveWasabi ? '#4caf50' : '#e0e0e0', isWasabi: true },
            { emoji: '🥢', label: 'Chopsticks', count: chopsticksStatus, color: chopsticksMode ? '#4caf50' : (canUseChopsticks ? '#8B4513' : '#e0e0e0'), isChopsticks: true, clickable: hasChopsticksAvailable }
        ];
        
        stats.forEach(stat => {
            const statItem = document.createElement('div');
            const isSpecial = stat.isWasabi || stat.isChopsticks;
            const bgColor = isSpecial ? stat.color : 'white';
            const textColor = (stat.isWasabi && hasActiveWasabi) || (stat.isChopsticks && chopsticksMode) ? 'white' : '#333';
            const shouldPulse = (stat.isWasabi && hasActiveWasabi) || (stat.isChopsticks && chopsticksMode);
            
            statItem.style.cssText = `
                background: ${bgColor};
                padding: 10px;
                border-radius: 6px;
                text-align: center;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                ${shouldPulse ? 'animation: pulse 2s infinite;' : ''}
                ${stat.clickable ? 'cursor: pointer; transition: transform 0.2s;' : ''}
            `;
            
            if (stat.clickable) {
                statItem.onclick = toggleChopsticks;
                statItem.onmouseover = () => {
                    statItem.style.transform = 'scale(1.05)';
                };
                statItem.onmouseout = () => {
                    statItem.style.transform = 'scale(1)';
                };
            }
            
            if (stat.isWasabi) {
                statItem.innerHTML = `
                    <div style="font-size: 20px; margin-bottom: 3px;">${hasActiveWasabi ? '🟢' : '⚪'}</div>
                    <div style="font-size: 14px; font-weight: bold; color: ${textColor}; margin-bottom: 2px;">${stat.count}</div>
                    <div style="font-size: 10px; color: ${hasActiveWasabi ? 'white' : '#666'}; text-transform: uppercase;">${stat.label}</div>
                `;
            } else if (stat.isChopsticks) {
                statItem.innerHTML = `
                    <div style="font-size: 20px; margin-bottom: 3px;">${stat.emoji}</div>
                    <div style="font-size: 14px; font-weight: bold; color: ${textColor}; margin-bottom: 2px;">${stat.count}</div>
                    <div style="font-size: 10px; color: ${textColor}; text-transform: uppercase;">${stat.label}</div>
                `;
            } else {
                statItem.innerHTML = `
                    <div style="font-size: 20px; margin-bottom: 3px;">${stat.emoji}</div>
                    <div style="font-size: 16px; font-weight: bold; color: ${textColor}; margin-bottom: 2px;">${stat.count}</div>
                    <div style="font-size: 10px; color: #666; text-transform: uppercase;">${stat.label}</div>
                `;
            }
            statsBar.appendChild(statItem);
        });
        
        handDiv.appendChild(statsBar);
    }
    
    // Remove animation class after animation completes to allow re-triggering
    if (animationType) {
        isAnimating = true;
        setTimeout(() => {
            cardsContainer.classList.remove('deal-animation', 'turn-animation');
            isAnimating = false;
        }, 600); // Match animation duration
    }
}

function updatePlayersList() {
    playersListDiv.innerHTML = '';
    
    if (!gameState || !gameState.players) {
        return;
    }
    
    gameState.players.forEach(player => {
        const li = document.createElement('li');
        li.className = 'player-item';
        
        const isMe = player.id === myPlayerId;
        const selectedIndicator = player.hasSelected ? '✓' : '○';
        
        // Count card types for this player
        let makiCount = 0;
        let dumplingCount = 0;
        let tempuraCount = 0;
        let sashimiCount = 0;
        let wasabiCount = 0;
        let nigiriCount = 0;
        
        if (player.collection) {
            player.collection.forEach(card => {
                if (card.type === 'maki_roll') makiCount += card.value || 0;
                if (card.type === 'dumpling') dumplingCount++;
                if (card.type === 'tempura') tempuraCount++;
                if (card.type === 'sashimi') sashimiCount++;
                if (card.type === 'wasabi') wasabiCount++;
                if (card.type === 'nigiri') nigiriCount++;
            });
        }
        
        // Count pudding cards
        const puddingCount = player.puddingCards ? player.puddingCards.length : 0;
        
        // Check if wasabi is active (more wasabi than nigiri means unused wasabi)
        const hasActiveWasabi = wasabiCount > nigiriCount;
        
        // Show kick button only in waiting phase and for other players
        const canKick = gameState.phase === 'waiting' && !isMe;
        
        li.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <div class="player-name">${player.name}${isMe ? ' (You)' : ''} ${selectedIndicator}</div>
                ${canKick ? `<button onclick="kickPlayer('${player.id}')" style="padding: 4px 8px; font-size: 12px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer;">Kick</button>` : ''}
            </div>
            <div class="player-stats">
                Score: ${player.score} | Hand: ${player.handSize} cards
            </div>
            <div style="display: flex; gap: 5px; margin-top: 8px; flex-wrap: wrap;">
                ${makiCount > 0 ? `<span class="mini-stat">🍣 ${makiCount}</span>` : ''}
                ${tempuraCount > 0 ? `<span class="mini-stat">🍤 ${tempuraCount}</span>` : ''}
                ${sashimiCount > 0 ? `<span class="mini-stat">🐟 ${sashimiCount}</span>` : ''}
                ${dumplingCount > 0 ? `<span class="mini-stat">🥟 ${dumplingCount}</span>` : ''}
                ${puddingCount > 0 ? `<span class="mini-stat">🍮 ${puddingCount}</span>` : ''}
                ${hasActiveWasabi ? `<span class="mini-stat mini-stat-active">🟢</span>` : ''}
            </div>
            <div class="collection">
                ${player.collection && player.collection.length > 0 ? player.collection.map(card => 
                    `<span class="collection-card">${formatCardType(card.type)}${card.variant ? ` (${card.variant})` : ''}${card.type === 'maki_roll' ? ` [${card.value || 0}]` : ''}</span>`
                ).join('') : '<span style="color: #999;">No cards yet</span>'}
            </div>
        `;
        
        playersListDiv.appendChild(li);
    });
}

function updateCollection() {
    collectionDiv.innerHTML = '';
    
    if (!gameState || !myPlayerId) {
        return;
    }
    
    const myPlayer = gameState.players.find(p => p.id === myPlayerId);
    if (!myPlayer || !myPlayer.collection || myPlayer.collection.length === 0) {
        collectionDiv.innerHTML = '<p style="color: #666;">No cards collected yet</p>';
        return;
    }
    
    // Count card types for set completion
    const cardCounts = {};
    let totalMakiValue = 0;
    
    myPlayer.collection.forEach(card => {
        cardCounts[card.type] = (cardCounts[card.type] || 0) + 1;
        if (card.type === 'maki_roll') {
            totalMakiValue += card.value || 0;
        }
    });
    
    // Track wasabi usage - count wasabi that haven't been paired with nigiri yet
    let availableWasabiCount = 0;
    
    // Track set completion indices
    let tempuraCount = 0;
    let sashimiCount = 0;
    let dumplingCount = 0;
    
    // Display cards with indicators
    myPlayer.collection.forEach((card, index) => {
        const cardEl = document.createElement('span');
        cardEl.className = 'collection-card';
        
        let cardText = `${formatCardType(card.type)}${card.variant ? ` (${card.variant})` : ''}`;
        let isSetComplete = false;
        let wasabiUsed = false;
        
        // Check for set completion
        if (card.type === 'tempura') {
            tempuraCount++;
            if (tempuraCount % 2 === 0) {
                cardText += ' ✓5pts';
                isSetComplete = true;
            }
        } else if (card.type === 'sashimi') {
            sashimiCount++;
            if (sashimiCount % 3 === 0) {
                cardText += ' ✓10pts';
                isSetComplete = true;
            }
        } else if (card.type === 'dumpling') {
            dumplingCount++;
            // Show points for dumplings: 1=1, 2=3, 3=6, 4=10, 5+=15
            const points = dumplingCount >= 5 ? 15 : [0, 1, 3, 6, 10][dumplingCount];
            cardText += ` (${points}pts)`;
        } else if (card.type === 'maki_roll') {
            cardText += ` [${card.value || 0}]`;
        }
        
        // Track wasabi cards
        if (card.type === 'wasabi') {
            availableWasabiCount++;
            
            // Check if this wasabi will be used by a nigiri that comes after it
            let hasNigiriAfter = false;
            for (let i = index + 1; i < myPlayer.collection.length; i++) {
                if (myPlayer.collection[i].type === 'nigiri') {
                    hasNigiriAfter = true;
                    break;
                }
            }
            
            if (hasNigiriAfter) {
                // Wasabi has been used - show as green
                wasabiUsed = true;
            }
        }
        
        // Check if this is a nigiri that can use wasabi
        if (card.type === 'nigiri' && availableWasabiCount > 0) {
            cardText += ' 3x';
            availableWasabiCount--; // Mark this wasabi as used
        }
        
        // Set text content first
        cardEl.textContent = cardText;
        
        // Apply styling after text (priority: wasabi > completed sets)
        if (wasabiUsed) {
            cardEl.style.background = '#4caf50';
            cardEl.style.color = 'white';
            cardEl.style.fontWeight = 'bold';
        } else if (isSetComplete) {
            cardEl.style.background = '#ffd700';
            cardEl.style.color = '#333';
            cardEl.style.fontWeight = 'bold';
        }
        
        collectionDiv.appendChild(cardEl);
    });
    
    // Show total maki count if any
    if (totalMakiValue > 0) {
        const makiTotal = document.createElement('div');
        makiTotal.style.cssText = 'margin-top: 10px; padding: 8px; background: #e3f2fd; border-radius: 4px; font-weight: bold; color: #1565c0;';
        makiTotal.textContent = `🍣 Total Maki: ${totalMakiValue}`;
        collectionDiv.appendChild(makiTotal);
    }
    
    // Show pudding cards (they persist across rounds)
    // NOTE: Backend sends puddingCards (camelCase)
    const puddingCards = myPlayer.puddingCards || [];
    
    if (puddingCards.length > 0) {
        const puddingDiv = document.createElement('div');
        puddingDiv.style.cssText = 'margin-top: 10px; padding: 8px; background: #ffb74d; border-radius: 4px; font-weight: bold; color: #333;';
        puddingDiv.textContent = `🍮 Pudding Cards: ${puddingCards.length} (scored at game end)`;
        collectionDiv.appendChild(puddingDiv);
    }
}

function formatCardType(type) {
    return type.split('_').map(word => 
        word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
}

function getCardColor(cardType) {
    const colors = {
        'maki_roll': 'linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%)', // Red
        'tempura': 'linear-gradient(135deg, #ffa726 0%, #fb8c00 100%)', // Orange
        'sashimi': 'linear-gradient(135deg, #ec407a 0%, #d81b60 100%)', // Pink
        'dumpling': 'linear-gradient(135deg, #ab47bc 0%, #8e24aa 100%)', // Purple
        'nigiri': 'linear-gradient(135deg, #26c6da 0%, #00acc1 100%)', // Cyan
        'wasabi': 'linear-gradient(135deg, #66bb6a 0%, #43a047 100%)', // Green
        'chopsticks': 'linear-gradient(135deg, #ffd54f 0%, #ffb300 100%)', // Yellow
        'pudding': 'linear-gradient(135deg, #ffb74d 0%, #f57c00 100%)' // Amber
    };
    return colors[cardType] || 'linear-gradient(135deg, #78909c 0%, #546e7a 100%)'; // Default gray
}

// Initialize - Auto-connect on page load
log('Connecting to server...', 'info');
connect();
