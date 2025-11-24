<script lang="ts">
  import { wsStore, type GameState, type Card } from './websocket';
  import ConveyorBeltCanvas from './ConveyorBeltCanvas.svelte';

  export let onLogout: () => void;

  let gameState: GameState | null = null;
  let selectedCardIndex: number | null = null;
  let secondCardIndex: number | null = null;
  let chopsticksMode = false;
  let previousHandSize = 0;

  wsStore.gameState.subscribe(value => {
    gameState = value;
    
    // Reset selection when hand changes (new round or cards passed)
    const currentHandSize = gameState?.myHand?.length || 0;
    if (currentHandSize !== previousHandSize) {
      selectedCardIndex = null;
      secondCardIndex = null;
      chopsticksMode = false;
      previousHandSize = currentHandSize;
    }
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
        wsStore.sendMessage('select_card', {
          cardIndex: selectedCardIndex,
          useChopsticks: true,
          secondCardIndex: secondCardIndex
        });
      }
    } else {
      selectedCardIndex = index;
      secondCardIndex = null;
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

  function getCardColor(type: string): string {
    const colors: Record<string, string> = {
      'maki_roll': 'from-red-400 to-red-600',
      'tempura': 'from-orange-400 to-orange-600',
      'sashimi': 'from-pink-400 to-pink-600',
      'dumpling': 'from-purple-400 to-purple-600',
      'nigiri': 'from-cyan-400 to-cyan-600',
      'wasabi': 'from-green-400 to-green-600',
      'chopsticks': 'from-yellow-400 to-yellow-600',
      'pudding': 'from-amber-400 to-amber-600'
    };
    return colors[type] || 'from-gray-400 to-gray-600';
  }

  function getCardEmoji(type: string): string {
    const emojis: Record<string, string> = {
      'maki_roll': '🍣',
      'tempura': '🍤',
      'sashimi': '🐟',
      'dumpling': '🥟',
      'nigiri': '🍱',
      'wasabi': '🟢',
      'chopsticks': '🥢',
      'pudding': '🍮'
    };
    return emojis[type] || '🎴';
  }

  function formatCardType(type: string): string {
    return type.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
  }

  // Count sets for multiplier indicators
  function countSets(collection: Card[]) {
    const tempuraCount = collection.filter(c => c.type === 'tempura').length;
    const sashimiCount = collection.filter(c => c.type === 'sashimi').length;
    const wasabiCount = collection.filter(c => c.type === 'wasabi').length;
    const dumplingCount = collection.filter(c => c.type === 'dumpling').length;
    
    return {
      tempuraSets: Math.floor(tempuraCount / 2),
      tempuraRemainder: tempuraCount % 2,
      sashimiSets: Math.floor(sashimiCount / 3),
      sashimiRemainder: sashimiCount % 3,
      wasabiActive: wasabiCount > 0,
      dumplingCount
    };
  }

  function handleCardSelection(index: number) {
    selectCard(index);
  }

  $: canStartGame = gameState?.phase === 'waiting' && (gameState?.players?.length || 0) >= 2;
  $: myPlayer = gameState?.players?.find(p => p.id === gameState?.myPlayerId);
  $: canUseChopsticks = (myPlayer?.chopsticksCount || 0) > 0;
  $: myHand = gameState?.myHand || [];
  $: selectedIndices = [selectedCardIndex, secondCardIndex].filter((i): i is number => i !== null);
  $: canSelectCards = gameState?.phase === 'selecting' && !myPlayer?.hasSelected;
</script>

<!-- Skeumorphic Conveyor Belt Sushi Restaurant -->
<div class="min-h-screen bg-gradient-to-b from-amber-900 via-amber-800 to-amber-900">
  <!-- Minimal Top Bar -->
  <div class="bg-black/30 text-white shadow-2xl">
    <div class="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-3xl">🍣</span>
        <div class="text-sm">
          {#if gameState?.gameId}
            Table: <span class="font-mono font-semibold">{gameState.gameId}</span>
            <button 
              on:click={copyGameId}
              class="ml-2 px-2 py-1 bg-white/20 rounded text-xs hover:bg-white/30"
            >
              📋
            </button>
          {/if}
        </div>
      </div>
      <div class="flex gap-3 items-center text-sm">
        {#if gameState?.currentRound}
          <span class="px-3 py-1 bg-white/20 rounded">
            Round {gameState.currentRound}/3
          </span>
        {/if}
        {#if canStartGame}
          <button 
            on:click={startGame}
            class="px-4 py-2 bg-green-500 hover:bg-green-600 rounded-lg font-semibold animate-pulse-rotate"
          >
            ⭐ Start Game
          </button>
        {/if}
        {#if canUseChopsticks && canSelectCards}
          <button
            on:click={toggleChopsticks}
            class="px-3 py-2 rounded-lg font-semibold {chopsticksMode ? 'bg-yellow-500 text-black' : 'bg-white/20'}"
          >
            🥢 {chopsticksMode ? 'Chopsticks ON' : 'Use Chopsticks'}
          </button>
        {/if}
        <button 
          on:click={onLogout}
          class="px-3 py-2 bg-red-500/80 hover:bg-red-600 rounded-lg"
        >
          Logout
        </button>
      </div>
    </div>
  </div>

  <div class="max-w-7xl mx-auto p-4">
    <!-- Main Conveyor Belt Canvas -->
    <div class="my-6">
      {#if gameState?.phase === 'waiting'}
        <div class="text-center py-32 text-white">
          <div class="text-9xl mb-6 animate-bounce">🍣</div>
          <div class="text-4xl font-bold mb-4 drop-shadow-lg">Ready to Play!</div>
          <div class="text-xl opacity-90">Click "Start Game" when all players are ready</div>
        </div>
      {:else if myHand.length > 0}
        <ConveyorBeltCanvas 
          cards={myHand}
          onCardSelect={handleCardSelection}
          selectedIndices={selectedIndices}
          canSelect={canSelectCards}
        />
      {:else}
        <div class="text-center py-32 text-white">
          <div class="text-7xl mb-4">🍽️</div>
          <div class="text-2xl opacity-80">Waiting for cards...</div>
        </div>
      {/if}
    </div>

    <!-- Bottom Info Panel -->
    <div class="grid grid-cols-3 gap-4 mt-6">
      <!-- Players -->
      <div class="bg-black/40 backdrop-blur-sm rounded-xl p-4 text-white">
        <h3 class="text-lg font-bold mb-3 flex items-center">
          <span class="text-2xl mr-2">👥</span>
          Players
        </h3>
        {#if gameState?.players}
          <div class="space-y-2">
            {#each gameState.players as player}
              <div class="p-2 bg-white/10 rounded {player.id === gameState.myPlayerId ? 'ring-2 ring-yellow-400' : ''} {player.hasSelected ? 'ring-2 ring-green-400' : ''}">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    {#if player.hasSelected && gameState.phase === 'selecting'}
                      <span class="text-lg animate-bounce">✅</span>
                    {:else}
                      <span class="text-lg">👤</span>
                    {/if}
                    <div class="text-sm">
                      <div class="font-semibold">
                        {player.name}
                        {#if player.id === gameState.myPlayerId}
                          <span class="text-xs text-yellow-300">(You)</span>
                        {/if}
                      </div>
                      <div class="text-xs opacity-75">
                        Score: {player.score} | Hand: {player.handSize}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Your Collection -->
      <div class="bg-black/40 backdrop-blur-sm rounded-xl p-4 text-white">
        <h3 class="text-lg font-bold mb-3 flex items-center justify-between">
          <span class="flex items-center">
            <span class="text-2xl mr-2">🍱</span>
            Collection
          </span>
          <span class="text-2xl text-yellow-400">{myPlayer?.score || 0}</span>
        </h3>
        {#if myPlayer?.collection && myPlayer.collection.length > 0}
          <div class="flex flex-wrap gap-1">
            {#each myPlayer.collection as card}
              <span class="text-2xl">{getCardEmoji(card.type)}</span>
            {/each}
          </div>
        {:else}
          <div class="text-sm opacity-60 text-center py-4">No cards yet</div>
        {/if}
      </div>

      <!-- Set Indicators -->
      <div class="bg-black/40 backdrop-blur-sm rounded-xl p-4 text-white">
        <h3 class="text-lg font-bold mb-3">🎯 Sets</h3>
        {#if myPlayer?.collection}
          {@const tempura = myPlayer.collection.filter(c => c.type === 'tempura').length}
          {@const sashimi = myPlayer.collection.filter(c => c.type === 'sashimi').length}
          {@const wasabi = myPlayer.collection.filter(c => c.type === 'wasabi').length}
          <div class="space-y-2 text-sm">
            {#if tempura > 0}
              <div class="flex justify-between items-center">
                <span>🍤 Tempura:</span>
                <span class="font-bold">
                  {#if tempura >= 2}
                    <span class="text-green-400 animate-pulse">✨{Math.floor(tempura/2)}x2!</span>
                  {:else}
                    {tempura}/2
                  {/if}
                </span>
              </div>
            {/if}
            {#if sashimi > 0}
              <div class="flex justify-between items-center">
                <span>🐟 Sashimi:</span>
                <span class="font-bold">
                  {#if sashimi >= 3}
                    <span class="text-green-400 animate-pulse">✨{Math.floor(sashimi/3)}x3!</span>
                  {:else}
                    {sashimi}/3
                  {/if}
                </span>
              </div>
            {/if}
            {#if wasabi > 0}
              <div class="flex justify-between items-center">
                <span>🟢 Wasabi:</span>
                <span class="font-bold text-green-400 animate-pulse">✨3x Active!</span>
              </div>
            {/if}
            {#if myPlayer.puddingCards && myPlayer.puddingCards.length > 0}
              <div class="flex justify-between items-center">
                <span>🍮 Pudding:</span>
                <span class="font-bold">{myPlayer.puddingCards.length}</span>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>