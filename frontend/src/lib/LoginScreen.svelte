<script lang="ts">
  import { wsStore } from '../lib/websocket';
  import { onMount } from 'svelte';

  export let onJoinGame: () => void;

  let serverUrl = '';
  let playerName = '';
  let gameId = '';
  let games: any[] = [];
  let connected = false;
  let isConnecting = false;

  wsStore.connected.subscribe(value => {
    connected = value;
    if (value) {
      requestGamesList();
    }
  });

  onMount(() => {
    // Set the WebSocket URL based on current location
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Always use port 8080 for backend, even if frontend is on different port
    const hostname = window.location.hostname || 'localhost';
    serverUrl = `${protocol}//${hostname}:8080/ws`;
    
    // Auto-connect
    connectToServer();

    // Listen for games list
    wsStore.onMessage('list_games', (payload) => {
      games = payload.games || [];
    });
  });

  async function connectToServer() {
    if (isConnecting) return;
    isConnecting = true;
    
    try {
      await wsStore.connect(serverUrl);
    } catch (error) {
      console.error('Failed to connect:', error);
    } finally {
      isConnecting = false;
    }
  }

  function requestGamesList() {
    wsStore.sendMessage('list_games', {});
  }

  function createGame() {
    wsStore.sendMessage('join_game', {
      gameId: '',
      playerName: playerName
    });
    onJoinGame();
  }

  function joinGame() {
    if (!gameId) {
      alert('Please enter a game ID');
      return;
    }
    
    wsStore.sendMessage('join_game', {
      gameId: gameId,
      playerName: playerName
    });
    onJoinGame();
  }

  function joinGameById(id: string) {
    wsStore.sendMessage('join_game', {
      gameId: id,
      playerName: playerName
    });
    onJoinGame();
  }

  function deleteGame(id: string) {
    if (confirm(`Are you sure you want to delete game ${id}?`)) {
      wsStore.sendMessage('delete_game', { gameId: id });
    }
  }
</script>

<div class="min-h-screen bg-gradient-to-br from-primary to-purple-600 p-5 animate-fade-in">
  <div class="max-w-4xl mx-auto">
    <!-- Header -->
    <div class="bg-white rounded-lg shadow-lg p-5 mb-5">
      <div class="flex items-center justify-between">
        <h1 class="text-3xl font-bold text-gray-800">🍣 Sushi Go!</h1>
        {#if isConnecting}
          <span class="inline-block px-4 py-2 rounded-full text-sm font-semibold bg-yellow-100 text-yellow-800">
            <span class="inline-block w-3 h-3 border-2 border-yellow-800 border-t-transparent rounded-full animate-spin mr-2"></span>
            Connecting...
          </span>
        {:else if connected}
          <span class="inline-block px-4 py-2 rounded-full text-sm font-semibold bg-green-100 text-green-800 hidden">
            Connected
          </span>
        {/if}
      </div>
    </div>

    <!-- Login Controls -->
    <div class="bg-white rounded-lg shadow-lg p-5 mb-5">
      <div class="space-y-4">
        <div>
          <label for="serverUrl" class="block mb-1 font-semibold text-gray-700">Server URL:</label>
          <input 
            type="text" 
            id="serverUrl" 
            bind:value={serverUrl}
            class="w-full max-w-md px-4 py-2 border-2 border-gray-300 rounded-md text-sm"
            placeholder="ws://localhost:8080/ws"
          />
        </div>

        <div>
          <label for="playerName" class="block mb-1 font-semibold text-gray-700">Player Name:</label>
          <input 
            type="text" 
            id="playerName" 
            bind:value={playerName}
            class="w-full max-w-md px-4 py-2 border-2 border-gray-300 rounded-md text-sm"
            placeholder="Enter your name or leave empty for random name"
          />
          <small class="block mt-1 text-xs text-gray-600">
            💡 Leave empty to get a random name from famous sushi chefs, anime characters, or historical figures
          </small>
        </div>

        <div>
          <label for="gameId" class="block mb-1 font-semibold text-gray-700">Game ID (leave empty to create new game):</label>
          <input 
            type="text" 
            id="gameId" 
            bind:value={gameId}
            class="w-full max-w-md px-4 py-2 border-2 border-gray-300 rounded-md text-sm"
            placeholder="Leave empty to create new game"
          />
        </div>

        <div class="flex gap-3 flex-wrap">
          <button 
            on:click={createGame}
            disabled={!connected}
            class="px-4 py-2 bg-primary text-white border-none rounded-md text-sm font-semibold cursor-pointer transition-colors hover:bg-primary-dark disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Create New Game
          </button>
          <button 
            on:click={joinGame}
            disabled={!connected}
            class="px-4 py-2 bg-primary text-white border-none rounded-md text-sm font-semibold cursor-pointer transition-colors hover:bg-primary-dark disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Join Existing Game
          </button>
        </div>
      </div>
    </div>

    <!-- Existing Games List -->
    <div class="bg-white rounded-lg shadow-lg p-5">
      <h3 class="mb-4 text-xl font-semibold text-gray-800">Existing Games</h3>
      <div class="max-h-80 overflow-y-auto">
        {#if games.length === 0}
          <p class="text-center text-gray-500">No active games</p>
        {:else}
          {#each games as game}
            <div class="bg-gray-100 p-3 mb-2 rounded-md flex justify-between items-center">
              <div>
                <div class="font-bold text-gray-800">{game.id}</div>
                <div class="text-xs text-gray-600">Players: {game.playerCount} | Phase: {game.phase}</div>
              </div>
              <div class="flex gap-2">
                <button 
                  on:click={() => joinGameById(game.id)}
                  class="px-3 py-1 text-xs bg-primary text-white border-none rounded cursor-pointer hover:bg-primary-dark"
                >
                  Join
                </button>
                <button 
                  on:click={() => deleteGame(game.id)}
                  class="px-3 py-1 text-xs bg-red-600 text-white border-none rounded cursor-pointer hover:bg-red-700"
                >
                  ✕
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>
