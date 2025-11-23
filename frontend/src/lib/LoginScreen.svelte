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
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const hostname = window.location.hostname || 'localhost';
    serverUrl = `${protocol}//${hostname}:8080/ws`;
    connectToServer();
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

<div class="min-h-screen p-6 sm:p-8 animate-fade-in">
  <div class="max-w-5xl mx-auto">
    <!-- Header with Japanese aesthetic -->
    <div class="japanese-header rounded-2xl p-8 mb-8 text-center relative overflow-hidden">
      <div class="absolute inset-0 opacity-10 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGRlZnM+PHBhdHRlcm4gaWQ9ImdyaWQiIHdpZHRoPSI2MCIgaGVpZ2h0PSI2MCIgcGF0dGVyblVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHBhdGggZD0iTSAxMCAwIEwgMCAwIDAgMTAiIGZpbGw9Im5vbmUiIHN0cm9rZT0id2hpdGUiIHN0cm9rZS13aWR0aD0iMSIvPjwvcGF0dGVybj48L2RlZnM+PHJlY3Qgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSIgZmlsbD0idXJsKCNncmlkKSIvPjwvc3ZnPg==')]"></div>
      <div class="relative z-10">
        <h1 class="text-5xl sm:text-6xl font-bold mb-3 text-shadow">
          <span class="text-6xl sm:text-7xl">🍣</span> Sushi Go!
        </h1>
        <p class="text-lg opacity-90">Welcome to the Sushi House</p>
        {#if connected}
          <div class="mt-4 inline-block px-4 py-2 bg-white/20 rounded-full text-sm backdrop-blur-sm">
            <span class="inline-block w-2 h-2 bg-green-400 rounded-full animate-pulse mr-2"></span>
            Connected to Kitchen
          </div>
        {/if}
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="grid lg:grid-cols-2 gap-8">
      <!-- Left Column: Game Controls -->
      <div class="space-y-6">
        <!-- Connection Card -->
        <div class="card-base p-6">
          <h2 class="text-xl font-bold text-amber-900 mb-4 flex items-center">
            <span class="text-2xl mr-2">🏮</span>
            Server Connection
          </h2>
          <div class="space-y-4">
            <div>
              <label for="serverUrl" class="block text-sm font-medium text-gray-700 mb-2">
                Kitchen URL
              </label>
              <input 
                type="text" 
                id="serverUrl" 
                bind:value={serverUrl}
                class="w-full px-4 py-3 border-2 border-amber-800/30 rounded-lg focus:border-red-600 focus:ring-2 focus:ring-red-600/20 outline-none transition-all"
                placeholder="ws://localhost:8080/ws"
                disabled={isConnecting}
              />
            </div>
          </div>
        </div>

        <!-- Player Info Card -->
        <div class="card-base p-6">
          <h2 class="text-xl font-bold text-amber-900 mb-4 flex items-center">
            <span class="text-2xl mr-2">👤</span>
            Player Info
          </h2>
          <div class="space-y-4">
            <div>
              <label for="playerName" class="block text-sm font-medium text-gray-700 mb-2">
                Your Name
              </label>
              <input 
                type="text" 
                id="playerName" 
                bind:value={playerName}
                class="w-full px-4 py-3 border-2 border-amber-800/30 rounded-lg focus:border-red-600 focus:ring-2 focus:ring-red-600/20 outline-none transition-all"
                placeholder="Enter your name"
              />
              <p class="mt-2 text-xs text-gray-600 flex items-start">
                <span class="mr-1">💡</span>
                <span>Leave empty for a random name from famous sushi chefs or anime characters</span>
              </p>
            </div>

            <div>
              <label for="gameId" class="block text-sm font-medium text-gray-700 mb-2">
                Game ID <span class="text-xs text-gray-500">(optional)</span>
              </label>
              <input 
                type="text" 
                id="gameId" 
                bind:value={gameId}
                class="w-full px-4 py-3 border-2 border-amber-800/30 rounded-lg focus:border-red-600 focus:ring-2 focus:ring-red-600/20 outline-none transition-all"
                placeholder="Leave empty to create new"
              />
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-col sm:flex-row gap-4">
          <button 
            on:click={createGame}
            disabled={!connected}
            class="btn-primary flex-1 animate-gentle-pulse"
          >
            <span class="text-xl mr-2">🎴</span>
            Create New Game
          </button>
          <button 
            on:click={joinGame}
            disabled={!connected}
            class="btn-secondary flex-1"
          >
            <span class="text-xl mr-2">🚪</span>
            Join Game
          </button>
        </div>
      </div>

      <!-- Right Column: Active Tables (Games List) -->
      <div class="card-base p-6">
        <h2 class="text-xl font-bold text-amber-900 mb-4 flex items-center justify-between">
          <span class="flex items-center">
            <span class="text-2xl mr-2">🍱</span>
            Active Tables
          </span>
          <span class="text-sm font-normal text-gray-500">{games.length} {games.length === 1 ? 'table' : 'tables'}</span>
        </h2>
        
        <div class="space-y-3 max-h-[500px] overflow-y-auto pr-2" style="scrollbar-width: thin; scrollbar-color: #D32F2F #FFF8E1;">
          {#if games.length === 0}
            <div class="text-center py-12 text-gray-400">
              <div class="text-5xl mb-3">🍶</div>
              <p class="text-sm">No active tables</p>
              <p class="text-xs mt-1">Create a new game to start!</p>
            </div>
          {:else}
            {#each games as game}
              <div class="bg-gradient-to-r from-amber-50 to-white p-4 rounded-lg border-2 border-amber-800/20 hover:border-red-600/40 transition-all duration-200 hover:shadow-md">
                <div class="flex items-center justify-between">
                  <div class="flex-1">
                    <div class="font-bold text-amber-900 flex items-center">
                      <span class="text-lg mr-2">🏯</span>
                      {game.id}
                    </div>
                    <div class="flex gap-3 mt-2 text-xs">
                      <span class="px-2 py-1 bg-red-600/10 text-red-600 rounded-full font-medium">
                        👥 {game.playerCount} {game.playerCount === 1 ? 'player' : 'players'}
                      </span>
                      <span class="px-2 py-1 bg-amber-800/10 text-amber-800 rounded-full font-medium">
                        {game.phase}
                      </span>
                    </div>
                  </div>
                  <div class="flex gap-2 ml-4">
                    <button 
                      on:click={() => joinGameById(game.id)}
                      class="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-all shadow-sm hover:shadow-md"
                    >
                      Join
                    </button>
                    <button 
                      on:click={() => deleteGame(game.id)}
                      class="px-3 py-2 bg-red-100 text-red-600 rounded-lg text-sm hover:bg-red-200 transition-all"
                    >
                      ✕
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          {/if}
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="mt-8 text-center text-sm text-gray-600">
      <p>🎋 Traditional Japanese Sushi Card Game 🎋</p>
    </div>
  </div>
</div>

<style>
  /* Custom scrollbar for games list */
  :global(.overflow-y-auto::-webkit-scrollbar) {
    width: 6px;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-track) {
    background: #FFF8E1;
    border-radius: 3px;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-thumb) {
    background: #D32F2F;
    border-radius: 3px;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-thumb:hover) {
    background: #B71C1C;
  }
</style>
