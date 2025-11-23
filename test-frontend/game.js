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
                handleRoundEnd(message.payload);
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
    
    const previousHand = gameState?.myHand;
    const newHand = payload.myHand;
    const previousSelected = gameState?.players?.find(p => p.id === myPlayerId)?.hasSelected;
    const currentSelected = payload.players?.find(p => p.id === myPlayerId)?.hasSelected;
    
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

// Handle round end message
function handleRoundEnd(payload) {
    const round = payload.round || gameState?.currentRound || 0;
    
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

// Helper function to show slide-out animation before updating to new cards
function updateHandWithSlideOut(oldHand, callback) {
    const handDiv = document.getElementById('hand');
    
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
    
    // Clear and add temp container
    handDiv.innerHTML = '';
    handDiv.appendChild(tempContainer);
    
    // After animation completes, call callback to show new cards
    setTimeout(() => {
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
    // NOTE: Backend sends has_chopsticks (snake_case), not hasChopsticks
    const hasChopsticksInCollection = myPlayer?.collection?.some(card => card.type === 'chopsticks');
    const canUseChopsticks = myPlayer?.has_chopsticks !== undefined ? myPlayer.has_chopsticks : hasChopsticksInCollection;
    const hasSelected = myPlayer?.hasSelected;
    
    // If player has already selected, show waiting indicator separately
    if (hasSelected && gameState.phase === 'selecting') {
        const waitingDiv = document.createElement('div');
        waitingDiv.style.cssText = 'background: #fff3cd; color: #856404; padding: 12px; border-radius: 8px; margin-bottom: 15px; text-align: center;';
        waitingDiv.innerHTML = `
            <div style="font-size: 16px; font-weight: bold;">⏳ Waiting for other players...</div>
        `;
        
        handDiv.appendChild(waitingDiv);
    }
    
    // Show chopsticks toggle and controls (only if chopsticks are available to use)
    if (canUseChopsticks && gameState.phase === 'selecting' && !hasSelected) {
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
    }
    
    // Show hint for normal mode OUTSIDE the animated container
    if (gameState.phase === 'selecting' && !hasSelected && !canUseChopsticks) {
        const hintDiv = document.createElement('div');
        hintDiv.style.cssText = 'background: #e3f2fd; color: #1565c0; padding: 8px; border-radius: 5px; margin-bottom: 10px; font-size: 14px; text-align: center; position: relative; z-index: 2;';
        
        // Check if any card is selected (for non-chopsticks mode)
        if (selectedCardIndex !== null && !chopsticksMode) {
            hintDiv.textContent = '👆 Click selected card to withdraw';
            hintDiv.style.background = '#fff3cd';
            hintDiv.style.color = '#856404';
        } else {
            hintDiv.textContent = '👆 Click a card to play it';
        }
        
        handDiv.appendChild(hintDiv);
    }
    
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
        
        // Create stat items - 3 per row for consistent layout
        const stats = [
            { emoji: '🍣', label: 'Maki', count: makiCount, color: '#e3f2fd' },
            { emoji: '🍤', label: 'Tempura', count: tempuraCount, color: '#fff3e0' },
            { emoji: '🐟', label: 'Sashimi', count: sashimiCount, color: '#fce4ec' },
            { emoji: '🥟', label: 'Dumpling', count: dumplingCount, color: '#f3e5f5' },
            { emoji: '🍮', label: 'Pudding', count: puddingCount, color: '#ffb74d' },
            { emoji: '🟢', label: 'Wasabi', count: wasabiStatus, color: hasActiveWasabi ? '#4caf50' : '#e0e0e0', isWasabi: true }
        ];
        
        stats.forEach(stat => {
            const statItem = document.createElement('div');
            const bgColor = stat.isWasabi ? stat.color : 'white';
            const textColor = stat.isWasabi && hasActiveWasabi ? 'white' : '#333';
            
            statItem.style.cssText = `
                background: ${bgColor};
                padding: 10px;
                border-radius: 6px;
                text-align: center;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                ${stat.isWasabi && hasActiveWasabi ? 'animation: pulse 2s infinite;' : ''}
            `;
            
            if (stat.isWasabi) {
                statItem.innerHTML = `
                    <div style="font-size: 20px; margin-bottom: 3px;">${hasActiveWasabi ? '🟢' : '⚪'}</div>
                    <div style="font-size: 14px; font-weight: bold; color: ${textColor}; margin-bottom: 2px;">${stat.count}</div>
                    <div style="font-size: 10px; color: ${hasActiveWasabi ? 'white' : '#666'}; text-transform: uppercase;">${stat.label}</div>
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
        
        li.innerHTML = `
            <div class="player-name">${player.name}${isMe ? ' (You)' : ''} ${selectedIndicator}</div>
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

// Initialize
log('Test client loaded. Enter server URL and connect.', 'info');
