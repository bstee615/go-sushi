<script lang="ts">
  import { wsStore } from './lib/websocket';
  import LoginScreen from './lib/LoginScreen.svelte';
  import PlayingScreen from './lib/PlayingScreen.svelte';
  import { onMount } from 'svelte';

  let currentScreen: 'login' | 'playing' = 'login';

  function switchToPlaying() {
    currentScreen = 'playing';
  }

  function switchToLogin() {
    currentScreen = 'login';
    wsStore.gameState.set(null);
  }

  // Handle game deletion and kick events
  onMount(() => {
    wsStore.onMessage('game_deleted', (payload) => {
      alert(payload.message || 'This game has been deleted');
      switchToLogin();
    });

    wsStore.onMessage('player_kicked', (payload) => {
      alert(payload.message || 'You have been kicked from the game');
      switchToLogin();
    });

    wsStore.onMessage('round_end', (payload) => {
      console.log('Round ended:', payload);
      // Could add overlay animation here
    });

    wsStore.onMessage('game_end', (payload) => {
      console.log('Game ended:', payload);
      // Game end is handled in the PlayingScreen component
    });
  });
</script>

{#if currentScreen === 'login'}
  <LoginScreen onJoinGame={switchToPlaying} />
{:else}
  <PlayingScreen onLogout={switchToLogin} />
{/if}
