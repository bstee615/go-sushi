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
    
    // If already selected this card, deselect it (unless already confirmed)
    if (!myPlayer?.hasSelected) {
      if (chopsticksMode) {
        if (index === selectedCardIndex) {
          selectedCardIndex = null;
          return;
        } else if (index === secondCardIndex) {
          secondCardIndex = null;
          return;
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
        // Allow deselection if clicking same card
        if (index === selectedCardIndex) {
          selectedCardIndex = null;
          return;
        }
        selectedCardIndex = index;
        secondCardIndex = null;
        wsStore.sendMessage('select_card', {
          cardIndex: selectedCardIndex,
          useChopsticks: false
        });
      }
    } else {
      // Already confirmed, allow withdrawal
      withdrawCard();
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

  let canvasRef: any;

  $: canStartGame = gameState?.phase === 'waiting' && (gameState?.players?.length || 0) >= 2;
  $: myPlayer = gameState?.players?.find(p => p.id === gameState?.myPlayerId);
  $: canUseChopsticks = (myPlayer?.chopsticksCount || 0) > 0;
  $: myHand = gameState?.myHand || [];
  $: selectedIndices = [selectedCardIndex, secondCardIndex].filter((i): i is number => i !== null);
  $: canSelectCards = gameState?.phase === 'selecting' && !myPlayer?.hasSelected;
  $: allPlayers = gameState?.players || [];
</script>

<!-- Skeumorphic Conveyor Belt Sushi Restaurant - Full Canvas -->
<div class="min-h-screen bg-gradient-to-b from-amber-900 via-amber-800 to-amber-900">
  <!-- Minimal Floating Control Bar -->
  <div class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
    <div class="bg-black/60 backdrop-blur-lg text-white shadow-2xl rounded-2xl px-6 py-3 flex items-center gap-4">
      <span class="text-2xl">🍣</span>
      {#if gameState?.gameId}
        <div class="text-sm font-mono">
          {gameState.gameId}
          <button 
            on:click={copyGameId}
            class="ml-2 px-2 py-1 bg-white/20 rounded text-xs hover:bg-white/30"
          >
            📋
          </button>
        </div>
      {/if}
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
          ⭐ Start Game!
        </button>
      {/if}
      {#if canUseChopsticks && canSelectCards}
        <button
          on:click={toggleChopsticks}
          class="px-3 py-2 rounded-lg font-semibold animate-pulse-rotate {chopsticksMode ? 'bg-yellow-500 text-black' : 'bg-yellow-600/80 text-white'}"
        >
          🥢 {chopsticksMode ? 'Chopsticks Mode' : 'Use Chopsticks'}
          {#if chopsticksMode}
            <span class="text-xs block">(Pick 2 cards)</span>
          {/if}
        </button>
      {/if}
      <button 
        on:click={onLogout}
        class="px-3 py-2 bg-red-500/80 hover:bg-red-600 rounded-lg"
      >
        Exit
      </button>
    </div>
  </div>

  <!-- Main Conveyor Belt Canvas (Full Screen) -->
  <div class="flex items-center justify-center h-screen w-full overflow-hidden">
    {#if gameState?.phase === 'waiting'}
      <div class="text-center text-white">
        <div class="text-9xl mb-6 animate-bounce">🍣</div>
        <div class="text-5xl font-bold mb-4 drop-shadow-lg">Sushi Go! Restaurant</div>
        <div class="text-2xl opacity-90 mb-8">Waiting for players...</div>
        <div class="bg-black/40 backdrop-blur-sm rounded-xl p-6 inline-block">
          {#if gameState.players}
            <div class="space-y-2">
              {#each gameState.players as player}
                <div class="text-xl">
                  {player.id === gameState.myPlayerId ? '👤' : '👥'} {player.name}
                  {#if player.id === gameState.myPlayerId}
                    <span class="text-yellow-300">(You)</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {:else if myHand.length > 0}
      <ConveyorBeltCanvas 
        bind:this={canvasRef}
        cards={myHand}
        onCardSelect={handleCardSelection}
        selectedIndices={selectedIndices}
        canSelect={canSelectCards}
        players={allPlayers}
        myPlayerId={gameState?.myPlayerId || ''}
        currentRound={gameState?.round || 0}
      />
    {:else if gameState?.phase === 'selecting'}
      <div class="text-center text-white">
        <div class="text-7xl mb-4">⏳</div>
        <div class="text-3xl font-bold opacity-90">Waiting for cards...</div>
      </div>
    {:else if gameState?.phase === 'complete'}
      <div class="text-center text-white">
        <div class="text-9xl mb-6">🏆</div>
        <div class="text-5xl font-bold mb-8">Game Complete!</div>
        <div class="bg-black/40 backdrop-blur-sm rounded-xl p-8 inline-block">
          <h3 class="text-3xl mb-4">Final Scores</h3>
          {#if gameState.players}
            <div class="space-y-3">
              {#each gameState.players.sort((a, b) => b.score - a.score) as player, index}
                <div class="text-2xl flex items-center gap-4">
                  <span class="text-3xl">
                    {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : '🎖️'}
                  </span>
                  <span class="font-bold">{player.name}</span>
                  <span class="text-yellow-300">{player.score} points</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>