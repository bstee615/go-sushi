<script lang="ts">
  import { wsStore, type GameState, type Card } from './websocket';

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

  $: canStartGame = gameState?.phase === 'waiting' && (gameState?.players?.length || 0) >= 2;
  $: myPlayer = gameState?.players?.find(p => p.id === gameState?.myPlayerId);
  $: canUseChopsticks = (myPlayer?.chopsticksCount || 0) > 0;
  $: myHand = gameState?.myHand || [];
</script>

<div class="min-h-screen p-4 sm:p-6 animate-fade-in">
  <div class="max-w-7xl mx-auto">
    <!-- Header -->
    <div class="japanese-header rounded-2xl p-4 sm:p-6 mb-6">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <span class="text-4xl">🏮</span>
          <div>
            <h1 class="text-2xl sm:text-3xl font-bold text-shadow">Sushi Go! Game</h1>
            {#if gameState?.gameId}
              <div class="text-sm opacity-90 mt-1">
                Table: <span class="font-mono font-bold">{gameState.gameId}</span>
                <button 
                  on:click={copyGameId}
                  class="ml-2 px-2 py-1 bg-white/20 rounded text-xs hover:bg-white/30 transition-all"
                >
                  📋 Copy
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
            class="px-4 py-2 bg-white/20 hover:bg-white/30 text-white rounded-lg transition-all"
          >
            Logout
          </button>
        </div>
      </div>
    </div>

    <!-- Main Game Area -->
    <div class="grid lg:grid-cols-3 gap-6">
      <!-- Left: Hand and Collection (Takes 2 columns) -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Your Hand -->
        <div class="card-base p-6">
          <h2 class="text-2xl font-bold text-amber-900 mb-4 flex items-center">
            <span class="text-3xl mr-2">🎴</span>
            Your Hand
          </h2>

          {#if gameState?.phase === 'waiting'}
            <!-- Waiting State -->
            <div class="text-center py-16">
              <div class="text-7xl mb-4 animate-bounce">🍣</div>
              <div class="text-2xl font-bold text-amber-900 mb-2">Ready to Play!</div>
              <div class="text-gray-600">Click "Start Game" when all players are ready</div>
            </div>
          {:else}
            <!-- Round Indicator -->
            {#if (gameState?.currentRound || 0) > 0}
              <div class="flex flex-wrap gap-3 mb-6">
                <div class="px-4 py-2 bg-amber-50 rounded-lg border-2 border-amber-800/30">
                  <div class="text-xs text-gray-600 font-semibold mb-1">ROUND</div>
                  <div class="flex gap-2">
                    {#each [1, 2, 3] as round}
                      <span class="text-xl font-bold {round === gameState?.currentRound ? 'text-red-600' : round < (gameState?.currentRound || 0) ? 'text-green-500' : 'text-gray-400'}">
                        {round}
                      </span>
                    {/each}
                  </div>
                </div>

                {#if myPlayer?.hasSelected && gameState?.phase === 'selecting'}
                  <div class="px-4 py-2 bg-amber-100 text-amber-800 rounded-lg border-2 border-amber-300 font-semibold flex items-center animate-gentle-pulse">
                    <span class="mr-2">⏳</span>
                    Waiting for others...
                  </div>
                {:else if gameState?.phase === 'selecting'}
                  <div class="px-4 py-2 bg-blue-100 text-blue-800 rounded-lg border-2 border-blue-300 font-semibold flex items-center">
                    <span class="mr-2">👆</span>
                    Click a card to play
                  </div>
                {/if}
              </div>
            {/if}

            <!-- Hand Cards - Fan Layout -->
            {#if myHand.length > 0}
              <div class="relative min-h-[280px] flex items-end justify-center pb-4">
                {#each myHand as card, index}
                  {@const isSelected = selectedCardIndex === index || secondCardIndex === index}
                  {@const rotation = (index - myHand.length / 2) * 4}
                  {@const translateY = Math.abs(index - myHand.length / 2) * 10}
                  
                  <button
                    on:click={() => selectCard(index)}
                    disabled={myPlayer?.hasSelected && !isSelected}
                    class="absolute transform transition-all duration-300 hover:scale-110 hover:-translate-y-8 disabled:opacity-50"
                    style="
                      left: {50 + (index - myHand.length / 2) * 12}%;
                      transform: translateX(-50%) translateY({isSelected ? -40 : translateY}px) rotate({rotation}deg);
                      z-index: {isSelected ? 100 : index};
                    "
                  >
                    <div class="w-32 h-44 bg-gradient-to-br {getCardColor(card.type)} rounded-xl shadow-card border-4 border-white relative overflow-hidden">
                      <!-- Card Header -->
                      <div class="absolute top-0 left-0 right-0 bg-white/90 p-2 text-center">
                        <div class="text-3xl mb-1">{getCardEmoji(card.type)}</div>
                        <div class="text-xs font-bold text-gray-800 uppercase">{formatCardType(card.type)}</div>
                      </div>
                      
                      <!-- Card Center Content -->
                      <div class="absolute inset-0 flex items-center justify-center text-white font-bold text-5xl opacity-20">
                        {getCardEmoji(card.type)}
                      </div>
                      
                      <!-- Card Footer -->
                      {#if card.variant || card.value !== undefined}
                        <div class="absolute bottom-0 left-0 right-0 bg-white/90 p-2 text-center">
                          <div class="text-xs font-bold text-gray-800">
                            {card.variant || `${card.value}`}
                          </div>
                        </div>
                      {/if}

                      <!-- Selection Indicator -->
                      {#if isSelected}
                        <div class="absolute inset-0 border-4 border-yellow-400 rounded-xl animate-pulse"></div>
                        <div class="absolute top-2 right-2 bg-yellow-400 text-yellow-900 rounded-full w-6 h-6 flex items-center justify-center font-bold text-sm">
                          ✓
                        </div>
                      {/if}
                    </div>
                  </button>
                {/each}
              </div>
            {:else}
              <div class="text-center py-8 text-gray-500">
                <div class="text-4xl mb-2">🍽️</div>
                <p>No cards in hand</p>
              </div>
            {/if}
          {/if}
        </div>

        <!-- Collection -->
        {#if gameState?.phase !== 'waiting' && myPlayer?.collection}
          <div class="card-base p-6">
            <h2 class="text-2xl font-bold text-amber-900 mb-4 flex items-center justify-between">
              <span class="flex items-center">
                <span class="text-3xl mr-2">🍱</span>
                Your Collection
              </span>
              <div class="text-4xl font-bold text-red-600">{myPlayer.score || 0}</div>
            </h2>

            <div class="flex flex-wrap gap-2 mb-4">
              {#if myPlayer.collection.length === 0}
                <p class="text-gray-500 w-full text-center py-4">No cards collected yet</p>
              {:else}
                {#each myPlayer.collection as card}
                  <div class="px-3 py-2 bg-gradient-to-br {getCardColor(card.type)} text-white rounded-lg text-sm font-semibold shadow-sm flex items-center gap-1">
                    <span>{getCardEmoji(card.type)}</span>
                    <span>{formatCardType(card.type)}</span>
                    {#if card.variant}
                      <span class="text-xs opacity-75">({card.variant})</span>
                    {/if}
                    {#if card.type === 'maki_roll'}
                      <span class="text-xs opacity-75">[{card.value || 0}]</span>
                    {/if}
                  </div>
                {/each}
              {/if}
            </div>

            <div class="p-3 bg-amber-100 rounded-lg border-2 border-amber-300 text-center">
              <span class="text-2xl mr-2">🍮</span>
              <span class="font-bold text-amber-900">
                Pudding: {myPlayer.puddingCards?.length || 0}
              </span>
              <span class="text-xs text-amber-700 ml-2">(scored at end)</span>
            </div>
          </div>
        {/if}
      </div>

      <!-- Right: Players List -->
      <div class="space-y-6">
        <div class="card-base p-6">
          <h2 class="text-2xl font-bold text-amber-900 mb-4 flex items-center">
            <span class="text-3xl mr-2">👥</span>
            Players
          </h2>

          {#if gameState?.players}
            <div class="space-y-3">
              {#each gameState.players as player}
                <div class="p-4 bg-gradient-to-r from-amber-50 to-white rounded-lg border-2 {player.id === gameState.myPlayerId ? 'border-red-600' : 'border-amber-800/20'}">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-2xl">{player.hasSelected ? '✓' : '○'}</span>
                      <div>
                        <div class="font-bold text-amber-900">
                          {player.name}
                          {#if player.id === gameState.myPlayerId}
                            <span class="text-xs text-red-600">(You)</span>
                          {/if}
                        </div>
                        <div class="text-xs text-gray-600">
                          Score: {player.score} | Hand: {player.handSize}
                        </div>
                      </div>
                    </div>
                    {#if gameState.phase === 'waiting' && player.id !== gameState.myPlayerId}
                      <button 
                        on:click={() => kickPlayer(player.id)}
                        class="px-2 py-1 bg-red-100 text-red-600 rounded text-xs hover:bg-red-200"
                      >
                        Kick
                      </button>
                    {/if}
                  </div>

                  {#if player.collection && player.collection.length > 0}
                    <div class="flex flex-wrap gap-1 mt-2">
                      {#each player.collection.slice(0, 6) as card}
                        <span class="text-lg">{getCardEmoji(card.type)}</span>
                      {/each}
                      {#if player.collection.length > 6}
                        <span class="text-xs text-gray-500">+{player.collection.length - 6}</span>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>