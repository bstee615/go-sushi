// Game state
let ws = null;
let gameState = null;
let myPlayerId = null;
let selectedCardIndex = null;

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
    }
    
    // Update UI
    updatePhaseIndicator();
    updateHand();
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

function selectCard(index) {
    if (!gameState || gameState.phase !== 'selecting') {
        log('Cannot select card: not in selection phase', 'error');
        return;
    }
    
    selectedCardIndex = index;
    updateHand(); // Update visual selection
    
    // Send selection
    sendMessage('select_card', {
        cardIndex: index,
        useChopsticks: false
    });
    
    log(`Selected card at index ${index}`, 'sent');
}

// UI Updates
function updatePhaseIndicator() {
    if (!gameState) {
        phaseIndicator.textContent = '';
        return;
    }
    
    phaseIndicator.className = 'phase-indicator';
    const round = gameState.currentRound || gameState.current_round || 0;
    phaseIndicator.textContent = `Round ${round} - ${gameState.phase}`;
}

function updateHand() {
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
    
    // Display actual cards
    myHand.forEach((card, index) => {
        const cardEl = document.createElement('div');
        cardEl.className = 'card';
        if (selectedCardIndex === index) {
            cardEl.className += ' selected';
        }
        cardEl.onclick = () => selectCard(index);
        
        cardEl.innerHTML = `
            <div class="card-type">${formatCardType(card.type)}</div>
            ${card.variant ? `<div class="card-variant">${card.variant}</div>` : ''}
            ${card.value ? `<div class="card-variant">Value: ${card.value}</div>` : ''}
        `;
        
        handDiv.appendChild(cardEl);
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
