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

// UI Elements
const connectionStatus = document.getElementById('connectionStatus');
const connectBtn = document.getElementById('connectBtn');
const disconnectBtn = document.getElementById('disconnectBtn');
const createBtn = document.getElementById('createBtn');
const joinBtn = document.getElementById('joinBtn');
const startBtn = document.getElementById('startBtn');
const handDiv = document.getElementById('hand');
const playersListDiv = document.getElementById('playersList');
const collectionDiv = document.getElementById('collection');
const logDiv = document.getElementById('log');
const phaseIndicator = document.getElementById('phaseIndicator');
const currentGameIdDiv = document.getElementById('currentGameId');
const gameIdDisplay = document.getElementById('gameIdDisplay');

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
    
    try {
        ws = new WebSocket(serverUrl);
        
        ws.onopen = () => {
            log('Connected to server', 'sent');
            connectionStatus.textContent = 'Connected';
            connectionStatus.className = 'connection-status status-connected';
            connectBtn.disabled = true;
            disconnectBtn.disabled = false;
            createBtn.disabled = false;
            joinBtn.disabled = false;
        };
        
        ws.onclose = () => {
            log('Disconnected from server', 'error');
            connectionStatus.textContent = 'Disconnected';
            connectionStatus.className = 'connection-status status-disconnected';
            connectBtn.disabled = false;
            disconnectBtn.disabled = true;
            createBtn.disabled = true;
            joinBtn.disabled = true;
            startBtn.disabled = true;
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

function disconnect() {
    if (ws) {
        ws.close();
        ws = null;
    }
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
                log(`Round ended: ${JSON.stringify(message.payload)}`, 'received');
                break;
            case 'game_end':
                log(`Game ended: ${JSON.stringify(message.payload)}`, 'received');
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
    const previousHand = gameState?.myHand;
    const newHand = payload.myHand;
    const previousSelected = gameState?.players?.find(p => p.id === myPlayerId)?.hasSelected;
    const currentSelected = payload.players?.find(p => p.id === myPlayerId)?.hasSelected;
    
    // Detect animation type
    const currentRound = payload.currentRound || payload.current_round || 0;
    const currentHandSize = newHand?.length || 0;
    
    // Determine if this is a new deal (start of round) or turn transition (hand passed)
    let animationType = null;
    if (currentRound !== previousRound && currentRound > 0) {
        // New round - deal animation
        animationType = 'deal';
        isFirstDeal = false;
    } else if (previousHandSize > 0 && currentHandSize > 0 && currentHandSize !== previousHandSize) {
        // Hand passed - turn animation
        animationType = 'turn';
    } else if (currentHandSize > 0 && previousHandSize === 0 && !isFirstDeal) {
        // First hand of a new round
        animationType = 'deal';
    }
    
    // Update tracking variables
    previousHandSize = currentHandSize;
    previousRound = currentRound;
    
    gameState = payload;
    
    // Get my player ID from the payload
    if (payload.myPlayerId) {
        myPlayerId = payload.myPlayerId;
    }
    
    // Update game ID display
    if (payload.gameId) {
        document.getElementById('gameId').value = payload.gameId;
        gameIdDisplay.textContent = payload.gameId;
        currentGameIdDiv.style.display = 'block';
    }
    
    // Reset selected card index if hand has changed
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

// Game actions
function createGame() {
    const playerName = document.getElementById('playerName').value;
    
    if (!playerName) {
        log('Please enter a player name', 'error');
        return;
    }
    
    log('Creating new game...', 'info');
    
    // Send join_game with empty gameId to create a new game
    sendMessage('join_game', {
        gameId: '',
        playerName: playerName
    });
}

function joinGame() {
    const playerName = document.getElementById('playerName').value;
    const gameId = document.getElementById('gameId').value;
    
    if (!playerName) {
        log('Please enter a player name', 'error');
        return;
    }
    
    if (!gameId) {
        log('Please enter a game ID to join, or click "Create New Game"', 'error');
        return;
    }
    
    log(`Joining game ${gameId}...`, 'info');
    
    sendMessage('join_game', {
        gameId: gameId,
        playerName: playerName
    });
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
        phaseIndicator.textContent = '';
        return;
    }
    
    phaseIndicator.className = 'phase-indicator';
    const round = gameState.currentRound || gameState.current_round || 0;
    
    // Calculate turn number (each round has multiple turns as hands are passed)
    const myPlayer = gameState.players?.find(p => p.id === myPlayerId);
    const handSize = myPlayer?.handSize || 0;
    const initialHandSize = gameState.phase === 'selecting' ? handSize : 0;
    
    // Estimate turn based on cards remaining (rough approximation)
    let turnInfo = '';
    if (gameState.phase === 'selecting' && handSize > 0) {
        turnInfo = ` - Turn ${11 - handSize}`;
    }
    
    phaseIndicator.textContent = `Round ${round}${turnInfo} - ${gameState.phase}`;
}

function updateHand(animationType = null) {
    handDiv.innerHTML = '';
    
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
    const hasChopsticks = myPlayer?.collection?.some(card => card.type === 'chopsticks');
    const hasSelected = myPlayer?.hasSelected;
    
    // If player has already selected, show waiting state
    if (hasSelected && gameState.phase === 'selecting') {
        const waitingDiv = document.createElement('div');
        waitingDiv.style.cssText = 'background: #fff3cd; color: #856404; padding: 15px; border-radius: 8px; margin-bottom: 15px; text-align: center;';
        waitingDiv.innerHTML = `
            <div style="font-size: 18px; font-weight: bold; margin-bottom: 8px;">⏳ Waiting for other players...</div>
            <div style="font-size: 14px; margin-bottom: 10px;">You have played your card(s) for this turn</div>
        `;
        
        const withdrawBtn = document.createElement('button');
        withdrawBtn.textContent = '↩️ Withdraw Selection';
        withdrawBtn.onclick = withdrawCard;
        withdrawBtn.style.cssText = `
            padding: 8px 16px;
            background: #dc3545;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-weight: 600;
        `;
        waitingDiv.appendChild(withdrawBtn);
        
        handDiv.appendChild(waitingDiv);
        
        // Show remaining cards greyed out
        const greyedHandDiv = document.createElement('div');
        greyedHandDiv.style.cssText = 'opacity: 0.4; pointer-events: none;';
        myHand.forEach((card, index) => {
            const cardEl = document.createElement('div');
            cardEl.className = 'card';
            cardEl.style.cssText = 'display: inline-block; margin: 5px;';
            
            cardEl.innerHTML = `
                <div class="card-type">${formatCardType(card.type)}</div>
                ${card.variant ? `<div class="card-variant">${card.variant}</div>` : ''}
                ${card.value ? `<div class="card-variant">Value: ${card.value}</div>` : ''}
            `;
            
            greyedHandDiv.appendChild(cardEl);
        });
        handDiv.appendChild(greyedHandDiv);
        return;
    }
    
    // Show chopsticks toggle and controls
    if (hasChopsticks && gameState.phase === 'selecting') {
        const controlsDiv = document.createElement('div');
        controlsDiv.style.cssText = 'background: #f5f5f5; padding: 12px; border-radius: 8px; margin-bottom: 15px;';
        
        const toggleBtn = document.createElement('button');
        toggleBtn.textContent = chopsticksMode ? '🥢 Put Away Chopsticks' : '🥢 Use Chopsticks!';
        toggleBtn.onclick = toggleChopsticks;
        toggleBtn.style.cssText = `
            padding: 8px 16px;
            background: ${chopsticksMode ? '#4caf50' : '#667eea'};
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-weight: 600;
            margin-right: 10px;
        `;
        
        controlsDiv.appendChild(toggleBtn);
        
        // Show status message
        const statusDiv = document.createElement('div');
        statusDiv.style.cssText = 'margin-top: 8px; font-size: 14px; color: #666;';
        if (chopsticksMode) {
            if (selectedCardIndex === null) {
                statusDiv.textContent = '📌 Select your first card (click to deselect)';
            } else if (secondCardIndex === null) {
                statusDiv.textContent = '📌 Select your second card (auto-submits when both selected)';
            } else {
                statusDiv.textContent = '✓ Both cards selected - submitting...';
            }
        } else {
            statusDiv.textContent = '📌 Click a card to play it immediately';
        }
        controlsDiv.appendChild(statusDiv);
        
        handDiv.appendChild(controlsDiv);
    } else if (gameState.phase === 'selecting') {
        // Show hint for normal mode
        const hintDiv = document.createElement('div');
        hintDiv.style.cssText = 'background: #e3f2fd; color: #1565c0; padding: 8px; border-radius: 5px; margin-bottom: 10px; font-size: 14px; text-align: center;';
        hintDiv.textContent = '👆 Click a card to play it';
        handDiv.appendChild(hintDiv);
    }
    
    // Display actual cards with animation
    myHand.forEach((card, index) => {
        const cardEl = document.createElement('div');
        cardEl.className = 'card';
        
        // Add animation class based on type
        if (animationType === 'deal') {
            cardEl.classList.add('deal-animation');
        } else if (animationType === 'turn') {
            cardEl.classList.add('turn-animation');
        }
        
        if (selectedCardIndex === index || secondCardIndex === index) {
            cardEl.className += ' selected';
        }
        cardEl.onclick = () => selectCard(index);
        
        cardEl.innerHTML = `
            <div class="card-type">${formatCardType(card.type)}</div>
            ${card.variant ? `<div class="card-variant">${card.variant}</div>` : ''}
            ${card.value ? `<div class="card-variant">Value: ${card.value}</div>` : ''}
        `;
        
        handDiv.appendChild(cardEl);
        
        // Remove animation class after animation completes to allow re-triggering
        if (animationType) {
            setTimeout(() => {
                cardEl.classList.remove('deal-animation', 'turn-animation');
            }, 500 + (index * 50)); // Match animation duration + delay
        }
    });
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
        
        li.innerHTML = `
            <div class="player-name">${player.name}${isMe ? ' (You)' : ''} ${selectedIndicator}</div>
            <div class="player-stats">
                Score: ${player.score} | Hand: ${player.handSize} cards
            </div>
            <div class="collection">
                ${player.collection && player.collection.length > 0 ? player.collection.map(card => 
                    `<span class="collection-card">${formatCardType(card.type)}${card.variant ? ` (${card.variant})` : ''}</span>`
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
    
    myPlayer.collection.forEach(card => {
        const cardEl = document.createElement('span');
        cardEl.className = 'collection-card';
        cardEl.textContent = `${formatCardType(card.type)}${card.variant ? ` (${card.variant})` : ''}`;
        collectionDiv.appendChild(cardEl);
    });
}

function formatCardType(type) {
    return type.split('_').map(word => 
        word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
}

// Initialize
log('Test client loaded. Enter server URL and connect.', 'info');
