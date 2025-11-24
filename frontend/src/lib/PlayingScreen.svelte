<script lang="ts">
  import { wsStore, type GameState, type Card } from './websocket';

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

  $: canStartGame = gameState?.phase === 'waiting' && (gameState?.players?.length || 0) >= 2;
  $: myPlayer = gameState?.players?.find(p => p.id === gameState?.myPlayerId);
  $: canUseChopsticks = (myPlayer?.chopsticksCount || 0) > 0;
  $: myHand = gameState?.myHand || [];
  $: sets = myPlayer?.collection ? countSets(myPlayer.collection) : null;
</script>

<!-- Sushi Restaurant Layout -->
<div class="min-h-screen bg-gradient-to-b from-amber-50 to-orange-50">
  <!-- Wooden Bar at Top (like sushi counter) -->
  <div class="bg-gradient-to-r from-amber-800 via-amber-900 to-amber-800 text-white shadow-2xl">
    <div class="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-4xl">🍣</span>
        <div>
          <h1 class="text-2xl font-bold text-shadow">Sushi Go! Restaurant</h1>
          {#if gameState?.gameId}
            <div class="text-sm opacity-90">
              Table: <span class="font-mono font-semibold">{gameState.gameId}</span>
              <button 
                on:click={copyGameId}
                class="ml-2 px-2 py-1 bg-white/20 rounded text-xs hover:bg-white/30 transition-all"
              >
                📋
              </button>
            </div>
          {/if}
        </div>
      </div>
      <div class="flex gap-3">
        {#if canStartGame}
          <button 
            on:click={startGame}
            class="btn-primary animate-pulse-rotate"
          >
            <span class="text-xl mr-1">⭐</span>
            Start Game!
          </button>
        {/if}
        <button 
          on:click={onLogout}
          class="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-all"
        >
          Logout
        </button>
      </div>
    </div>
  </div>

  <div class="max-w-7xl mx-auto p-4">
    <!-- Main Restaurant Layout -->
    <div class="grid grid-cols-12 gap-4">
      
      <!-- Left Side: Players Sitting at Table -->
      <div class="col-span-3 space-y-3">
        <div class="bg-white rounded-xl shadow-lg p-4 border-4 border-amber-800">
          <h2 class="text-xl font-bold text-amber-900 mb-3 flex items-center">
            <span class="text-2xl mr-2">👥</span>
            Diners
          </h2>

          {#if gameState?.players}
            <div class="space-y-2">
              {#each gameState.players as player}
                <div class="p-3 bg-gradient-to-r from-amber-50 to-white rounded-lg border-2 {player.id === gameState.myPlayerId ? 'border-red-600 shadow-md' : 'border-amber-200'} {player.hasSelected && gameState.phase === 'selecting' ? 'ring-2 ring-green-400' : ''}">
                  <div class="flex items-start gap-2">
                    {#if player.hasSelected && gameState.phase === 'selecting'}
                      <span class="text-xl animate-bounce">✅</span>
                    {:else}
                      <span class="text-xl">🧑‍🍳</span>
                    {/if}
                    <div class="flex-1 min-w-0">
                      <div class="font-bold text-sm text-amber-900 truncate">
                        {player.name}
                        {#if player.id === gameState.myPlayerId}
                          <span class="text-xs text-red-600">(You)</span>
                        {/if}
                      </div>
                      <div class="text-xs text-gray-600 flex justify-between mt-1">
                        <span>Score: {player.score}</span>
                        <span>Hand: {player.handSize}</span>
                      </div>
                      {#if player.collection && player.collection.length > 0}
                        <div class="flex flex-wrap gap-1 mt-1">
                          {#each player.collection.slice(0, 6) as card}
                            <span class="text-sm">{getCardEmoji(card.type)}</span>
                          {/each}
                          {#if player.collection.length > 6}
                            <span class="text-xs text-gray-500">+{player.collection.length - 6}</span>
                          {/if}
                        </div>
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Center: Conveyor Belt with Sushi Cards -->
      <div class="col-span-9 space-y-4">
        
        <!-- Conveyor Belt Background -->
        <div class="bg-gradient-to-b from-gray-700 to-gray-800 rounded-2xl shadow-2xl p-6 border-8 border-gray-900" style="min-height: 300px;">
          <div class="bg-gray-600/30 rounded-xl p-4 backdrop-blur-sm">
            
            {#if gameState?.phase === 'waiting'}
              <!-- Waiting State -->
              <div class="text-center py-16">
                <div class="text-7xl mb-4 animate-bounce">🍣</div>
                <div class="text-3xl font-bold text-white mb-2 text-shadow">Ready to Play!</div>
                <div class="text-xl text-gray-300">Click "Start Game" to begin</div>
              </div>
            {:else}
              <!-- Round and Action Indicators -->
              <div class="flex flex-wrap gap-3 mb-6">
                <!-- Round Indicator -->
                {#if (gameState?.currentRound || 0) > 0}
                  <div class="px-4 py-2 bg-white rounded-lg border-2 border-amber-800 shadow-md">
                    <div class="text-xs text-gray-600 font-semibold mb-1">ROUND</div>
                    <div class="flex gap-2">
                      {#each [1, 2, 3] as round}
                        <span class="text-2xl font-bold {round === gameState?.currentRound ? 'text-red-600' : round < (gameState?.currentRound || 0) ? 'text-green-500' : 'text-gray-400'}">
                          {round}
                        </span>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Chopsticks Button -->
                {#if canUseChopsticks && gameState?.phase === 'selecting' && !myPlayer?.hasSelected}
                  <button
                    on:click={toggleChopsticks}
                    class="px-4 py-2 rounded-lg border-2 font-semibold transition-all shadow-md {chopsticksMode ? 'bg-yellow-500 text-white border-yellow-600 animate-pulse-rotate' : 'bg-white text-yellow-800 border-yellow-400 hover:bg-yellow-50'}"
                  >
                    <span class="text-xl mr-2">🥢</span>
                    Use Chopsticks!
                    {#if chopsticksMode}
                      <span class="ml-2 text-sm">(Pick 2)</span>
                    {/if}
                  </button>
                {/if}

                <!-- Status Message -->
                {#if myPlayer?.hasSelected && gameState?.phase === 'selecting'}
                  <div class="px-4 py-2 bg-amber-100 text-amber-900 rounded-lg border-2 border-amber-400 font-semibold animate-gentle-pulse shadow-md">
                    <span class="mr-2">⏳</span>
                    Waiting for others...
                  </div>
                {:else if gameState?.phase === 'selecting'}
                  <div class="px-4 py-2 bg-white text-blue-900 rounded-lg border-2 border-blue-400 font-semibold shadow-md">
                    <span class="mr-2">👆</span>
                    Click a card to play
                  </div>
                {/if}
              </div>

              <!-- Horizontal Conveyor Belt with Cards -->
              {#if myHand.length > 0}
                <div class="relative overflow-x-auto">
                  <div class="flex gap-4 pb-4" style="min-width: min-content;">
                    {#each myHand as card, index}
                      {@const isSelected = selectedCardIndex === index || secondCardIndex === index}
                      
                      <button
                        on:click={() => selectCard(index)}
                        disabled={myPlayer?.hasSelected && !isSelected}
                        class="flex-shrink-0 transform transition-all duration-300 hover:scale-110 hover:-translate-y-4 disabled:opacity-50 disabled:cursor-not-allowed"
                        class:scale-110={isSelected}
                        class:-translate-y-4={isSelected}
                      >
                        <div class="w-40 h-56 bg-gradient-to-br {getCardColor(card.type)} rounded-xl shadow-lg border-4 border-white relative overflow-hidden">
                          <!-- Card Header -->
                          <div class="absolute top-0 left-0 right-0 bg-white/95 p-3 text-center">
                            <div class="text-4xl mb-1">{getCardEmoji(card.type)}</div>
                            <div class="text-xs font-bold text-gray-800 uppercase tracking-wide">{formatCardType(card.type)}</div>
                          </div>
                          
                          <!-- Card Center -->
                          <div class="absolute inset-0 flex items-center justify-center text-white font-bold text-6xl opacity-20">
                            {getCardEmoji(card.type)}
                          </div>
                          
                          <!-- Card Footer -->
                          {#if card.variant || card.value !== undefined}
                            <div class="absolute bottom-0 left-0 right-0 bg-white/95 p-2 text-center">
                              <div class="text-sm font-bold text-gray-800">
                                {card.variant || card.value}
                              </div>
                            </div>
                          {/if}

                          <!-- Selection Glow -->
                          {#if isSelected}
                            <div class="absolute inset-0 border-4 border-yellow-400 rounded-xl animate-pulse shadow-xl"></div>
                            <div class="absolute top-2 right-2 bg-yellow-400 text-yellow-900 rounded-full w-8 h-8 flex items-center justify-center font-bold shadow-lg">
                              ✓
                            </div>
                          {/if}
                        </div>
                      </button>
                    {/each}
                  </div>
                </div>
              {:else}
                <div class="text-center py-12 text-white">
                  <div class="text-5xl mb-3">🍽️</div>
                  <p class="text-xl">No cards in hand</p>
                </div>
              {/if}
            {/if}
          </div>
        </div>

        <!-- Your Collection Display (Below Conveyor Belt) -->
        {#if gameState?.phase !== 'waiting' && myPlayer?.collection}
          <div class="bg-white rounded-xl shadow-lg p-6 border-4 border-amber-800">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-2xl font-bold text-amber-900 flex items-center">
                <span class="text-3xl mr-2">🍱</span>
                Your Collection
              </h2>
              <div class="text-4xl font-bold text-red-600">{myPlayer.score || 0}</div>
            </div>

            <!-- Collection Cards -->
            <div class="flex flex-wrap gap-2 mb-4">
              {#if myPlayer.collection.length === 0}
                <p class="text-gray-500 w-full text-center py-4">No cards collected yet</p>
              {:else}
                {#each myPlayer.collection as card}
                  <div class="px-3 py-2 bg-gradient-to-br {getCardColor(card.type)} text-white rounded-lg text-sm font-semibold shadow-sm">
                    <span class="text-lg mr-1">{getCardEmoji(card.type)}</span>
                    <span>{formatCardType(card.type)}</span>
                    {#if card.variant}
                      <span class="text-xs opacity-75 ml-1">({card.variant})</span>
                    {/if}
                    {#if card.type === 'maki_roll'}
                      <span class="text-xs opacity-75 ml-1">[{card.value || 0}]</span>
                    {/if}
                  </div>
                {/each}
              {/if}
            </div>

            <!-- Set Multipliers -->
            {#if sets}
              <div class="grid grid-cols-2 md:grid-cols-4 gap-2">
                <!-- Tempura -->
                {#if sets.tempuraSets > 0 || sets.tempuraRemainder > 0}
                  <div class="px-3 py-2 bg-orange-100 border-2 border-orange-300 rounded-lg text-center">
                    <div class="text-2xl">🍤</div>
                    <div class="text-xs font-bold text-orange-900">
                      {#if sets.tempuraSets > 0}
                        <span class="text-lg animate-pulse-rotate">✨{sets.tempuraSets}x2!</span>
                      {:else}
                        {sets.tempuraRemainder}/2
                      {/if}
                    </div>
                  </div>
                {/if}

                <!-- Sashimi -->
                {#if sets.sashimiSets > 0 || sets.sashimiRemainder > 0}
                  <div class="px-3 py-2 bg-pink-100 border-2 border-pink-300 rounded-lg text-center">
                    <div class="text-2xl">🐟</div>
                    <div class="text-xs font-bold text-pink-900">
                      {#if sets.sashimiSets > 0}
                        <span class="text-lg animate-pulse-rotate">✨{sets.sashimiSets}x3!</span>
                      {:else}
                        {sets.sashimiRemainder}/3
                      {/if}
                    </div>
                  </div>
                {/if}

                <!-- Wasabi -->
                {#if sets.wasabiActive}
                  <div class="px-3 py-2 bg-green-100 border-2 border-green-300 rounded-lg text-center">
                    <div class="text-2xl">🟢</div>
                    <div class="text-xs font-bold text-green-900">
                      <span class="text-lg animate-pulse-rotate">✨3x!</span>
                      <div class="text-xs opacity-75">Active</div>
                    </div>
                  </div>
                {/if}

                <!-- Chopsticks -->
                {#if canUseChopsticks}
                  <div class="px-3 py-2 bg-yellow-100 border-2 border-yellow-300 rounded-lg text-center">
                    <div class="text-2xl">🥢</div>
                    <div class="text-xs font-bold text-yellow-900">
                      {myPlayer.chopsticksCount} Available
                    </div>
                  </div>
                {/if}
              </div>
            {/if}

            <!-- Pudding -->
            <div class="mt-4 p-3 bg-amber-100 rounded-lg border-2 border-amber-300 text-center">
              <span class="text-2xl mr-2">🍮</span>
              <span class="font-bold text-amber-900">
                Pudding: {myPlayer.puddingCards?.length || 0}
              </span>
              <span class="text-xs text-amber-700 ml-2">(scored at end)</span>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>