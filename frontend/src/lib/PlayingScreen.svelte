<script lang="ts">
  import { wsStore, type GameState } from './websocket';
  import GameCanvas from './GameCanvas.svelte';

  export let onLogout: () => void;

  let gameState: GameState | null = null;
  let selectedCardIndex: number | null = null;
  let secondCardIndex: number | null = null;
  let chopsticksMode = false;

  wsStore.gameState.subscribe(value => {
    gameState = value;
  });

  function startGame() {
    if (!gameState?.gameId) return;
    wsStore.sendMessage('start_game', { gameId: gameState.gameId });
  }

  function copyGameId() {
    if (!gameState?.gameId) return;
    navigator.clipboard.writeText(gameState.gameId).then(() => {
      alert('Game ID copied to clipboard!');
    });
  }

  function selectCard(index: number) {
    if (!gameState || gameState.phase !== 'selecting') return;

    const myPlayer = gameState.players?.find(p => p.id === gameState.myPlayerId);
    if (myPlayer?.hasSelected) {
      console.log('Already selected, withdrawing...');
      withdrawCard();
      return;
    }

    if (chopsticksMode) {
      if (index === selectedCardIndex) {
        selectedCardIndex = null;
      } else if (index === secondCardIndex) {
        secondCardIndex = null;
      } else if (selectedCardIndex === null) {
        selectedCardIndex = index;
      } else if (secondCardIndex === null) {
        secondCardIndex = index;
        // Auto-submit both cards
        wsStore.sendMessage('select_card', {
          cardIndex: selectedCardIndex,
          useChopsticks: true,
          secondCardIndex: secondCardIndex
        });
      }
    } else {
      selectedCardIndex = index;
      secondCardIndex = null;
      // Auto-submit single card
      wsStore.sendMessage('select_card', {
        cardIndex: selectedCardIndex,
        useChopsticks: false
      });
    }
  }

  function withdrawCard() {
    if (!gameState?.gameId) return;
    wsStore.sendMessage('withdraw_card', { gameId: gameState.gameId });
    selectedCardIndex = null;
    secondCardIndex = null;
  }

  function toggleChopsticks() {
    chopsticksMode = !chopsticksMode;
    if (!chopsticksMode) {
      selectedCardIndex = null;
      secondCardIndex = null;
    }
  }

  function kickPlayer(playerId: string) {
    if (!gameState?.gameId) return;
    if (gameState.phase !== 'waiting') {
      alert('Can only kick players before game starts');
      return;
    }
    wsStore.sendMessage('kick_player', { playerId });
  }

  function formatCardType(type: string): string {
    return type.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
  }

  $: canStartGame = gameState?.phase === 'waiting' && (gameState?.players?.length || 0) >= 2;
  $: myPlayer = gameState?.players?.find(p => p.id === gameState?.myPlayerId);
  $: canUseChopsticks = (myPlayer?.chopsticksCount || 0) > 0;
</script>

<div class="min-h-screen bg-gradient-to-br from-primary to-purple-600 p-5 animate-fade-in">
  <div class="max-w-7xl mx-auto">
    <!-- Header -->
    <div class="bg-white rounded-lg shadow-lg p-5 mb-5">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <h1 class="text-3xl font-bold text-gray-800">🍣 Sushi Go! Game</h1>
        <div class="flex items-center gap-4 flex-wrap">
          {#if gameState?.gameId}
            <div class="px-4 py-2 bg-gray-100 rounded-md">
              <strong>Game:</strong> {gameState.gameId}
              <button 
                on:click={copyGameId}
                class="ml-2 px-2 py-1 text-xs bg-primary text-white rounded hover:bg-primary-dark"
              >
                Copy
              </button>
            </div>
          {/if}
          <button 
            on:click={startGame}
            disabled={!canStartGame}
            class="px-4 py-2 bg-primary text-white rounded text-sm font-semibold hover:bg-primary-dark disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Start Game
          </button>
          <button 
            on:click={onLogout}
            class="px-4 py-2 bg-red-600 text-white rounded text-sm font-semibold hover:bg-red-700"
          >
            Logout
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
      <!-- Main Game Area (Left side, larger) -->
      <div class="lg:col-span-2 space-y-5">
        <!-- Hand Panel -->
        <div class="bg-white rounded-lg shadow-lg p-5">
          <h2 class="text-2xl font-bold text-gray-800 mb-4 pb-2 border-b-2 border-gray-200">Your Hand</h2>
          
          {#if gameState?.phase === 'waiting'}
            <div class="text-center py-16">
              <div class="text-5xl mb-5">🍣</div>
              <div class="text-2xl font-bold text-gray-800 mb-2">Ready to Play!</div>
              <div class="text-lg text-gray-600">Click "Start Game" to begin</div>
            </div>
          {:else}
            <!-- Round and Turn Indicators -->
            <div class="flex gap-3 mb-4 flex-wrap">
              {#if (gameState?.currentRound || gameState?.current_round || 0) > 0}
                <div class="bg-gray-100 px-4 py-2 rounded-md flex flex-col items-center min-w-[120px]">
                  <div class="text-xs font-semibold text-gray-600 mb-1">ROUND</div>
                  <div class="flex gap-1">
                    {#each [1, 2, 3] as round}
                      <span class="text-lg font-bold {round === (gameState?.currentRound || gameState?.current_round) ? 'text-blue-500' : round < (gameState?.currentRound || gameState?.current_round || 0) ? 'text-green-500' : 'text-gray-400'}">
                        {round}
                      </span>
                    {/each}
                  </div>
                </div>
              {/if}
              
              {#if myPlayer?.hasSelected && gameState?.phase === 'selecting'}
                <div class="bg-yellow-100 text-yellow-800 px-4 py-2 rounded-md flex items-center flex-1 min-w-[200px] justify-center font-semibold">
                  ⏳ Waiting for other players...
                </div>
              {:else if gameState?.phase === 'selecting' && !myPlayer?.hasSelected}
                <div class="bg-blue-100 text-blue-800 px-4 py-2 rounded-md flex items-center flex-1 min-w-[200px] justify-center font-semibold">
                  👆 Click a card to play it
                </div>
              {/if}
            </div>

            <!-- Game Canvas -->
            <GameCanvas 
              {gameState}
              onSelectCard={selectCard}
              {selectedCardIndex}
              {secondCardIndex}
            />

            <!-- Stats Bar -->
            {#if myPlayer?.collection}
              <div class="mt-5 p-4 bg-gradient-to-br from-primary to-purple-600 rounded-lg grid grid-cols-3 sm:grid-cols-4 md:grid-cols-7 gap-2">
                {#each [
                  { emoji: '🍣', label: 'Maki', count: myPlayer.collection.filter(c => c.type === 'maki_roll').reduce((sum, c) => sum + (c.value || 0), 0) },
                  { emoji: '🍤', label: 'Tempura', count: myPlayer.collection.filter(c => c.type === 'tempura').length },
                  { emoji: '🐟', label: 'Sashimi', count: myPlayer.collection.filter(c => c.type === 'sashimi').length },
                  { emoji: '🥟', label: 'Dumpling', count: myPlayer.collection.filter(c => c.type === 'dumpling').length },
                  { emoji: '🍮', label: 'Pudding', count: myPlayer.puddingCards?.length || 0 },
                  { emoji: '🟢', label: 'Wasabi', count: 'ACTIVE', special: true },
                  { emoji: '🥢', label: 'Chopsticks', count: myPlayer.chopsticksCount || 0, special: true, clickable: canUseChopsticks }
                ] as stat}
                  {#if stat.clickable}
                    <button 
                      class="bg-white p-2 rounded text-center cursor-pointer hover:scale-105 transition-transform"
                      on:click={toggleChopsticks}
                    >
                      <div class="text-xl mb-1">{stat.emoji}</div>
                      <div class="text-sm font-bold text-gray-800">{stat.count}</div>
                      <div class="text-xs text-gray-600 uppercase">{stat.label}</div>
                    </button>
                  {:else}
                    <div class="bg-white p-2 rounded text-center">
                      <div class="text-xl mb-1">{stat.emoji}</div>
                      <div class="text-sm font-bold text-gray-800">{stat.count}</div>
                      <div class="text-xs text-gray-600 uppercase">{stat.label}</div>
                    </div>
                  {/if}
                {/each}
              </div>
            {/if}
          {/if}
        </div>

        <!-- Collection Panel -->
        {#if gameState?.phase !== 'waiting' && myPlayer?.collection}
          <div class="bg-white rounded-lg shadow-lg p-5">
            <h2 class="text-2xl font-bold text-gray-800 mb-4 pb-2 border-b-2 border-gray-200">Your Collection</h2>
            
            <!-- Score Display -->
            <div class="mb-4 p-4 bg-gradient-to-br from-primary to-purple-600 rounded-lg text-white text-center">
              <div class="text-sm opacity-90 mb-2">TOTAL SCORE</div>
              <div class="text-4xl font-bold">{myPlayer.score || 0}</div>
              {#if myPlayer.roundScores && myPlayer.roundScores.length > 0}
                <div class="text-xs opacity-80 mt-2">
                  {myPlayer.roundScores.map((score, i) => `R${i + 1}: ${score}pts`).join(' | ')}
                </div>
              {/if}
            </div>

            <!-- Collection Cards -->
            <div class="flex flex-wrap gap-2">
              {#if myPlayer.collection.length === 0}
                <p class="text-gray-600">No cards collected yet</p>
              {:else}
                {#each myPlayer.collection as card}
                  <span class="bg-gray-200 px-3 py-1 rounded text-sm text-gray-800">
                    {formatCardType(card.type)}
                    {#if card.variant} ({card.variant}){/if}
                    {#if card.type === 'maki_roll'} [{card.value || 0}]{/if}
                  </span>
                {/each}
              {/if}
            </div>

            <!-- Pudding Counter -->
            <div class="mt-4 p-3 bg-amber-500 rounded-lg text-center font-bold text-gray-800">
              🍮 Pudding: {myPlayer.puddingCards?.length || 0} 
              <span class="text-xs opacity-80">(scored at game end)</span>
            </div>
          </div>
        {/if}
      </div>

      <!-- Players Panel (Right side) -->
      <div class="space-y-5">
        <div class="bg-white rounded-lg shadow-lg p-5">
          <h2 class="text-2xl font-bold text-gray-800 mb-4 pb-2 border-b-2 border-gray-200">Players</h2>
          
          {#if gameState?.players}
            <ul class="space-y-3">
              {#each gameState.players as player}
                <li class="bg-gray-100 p-4 rounded-lg border-l-4 border-primary">
                  <div class="flex justify-between items-center mb-2">
                    <div class="font-bold text-gray-800">
                      {player.name}
                      {#if player.id === gameState.myPlayerId} (You){/if}
                      {player.hasSelected ? ' ✓' : ' ○'}
                    </div>
                    {#if gameState.phase === 'waiting' && player.id !== gameState.myPlayerId}
                      <button 
                        on:click={() => kickPlayer(player.id)}
                        class="px-2 py-1 text-xs bg-red-600 text-white rounded hover:bg-red-700"
                      >
                        Kick
                      </button>
                    {/if}
                  </div>
                  <div class="text-sm text-gray-600">
                    Score: {player.score} | Hand: {player.handSize} cards
                  </div>
                  {#if player.collection && player.collection.length > 0}
                    <div class="flex flex-wrap gap-1 mt-2">
                      {#each player.collection as card}
                        <span class="text-xs bg-white px-2 py-1 rounded shadow-sm">
                          {formatCardType(card.type)}
                        </span>
                      {/each}
                    </div>
                  {/if}
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>
