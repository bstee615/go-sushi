<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Application, Graphics, Text, Container, Assets } from 'pixi.js';
  import type { Card } from './websocket';

  export let cards: Card[] = [];
  export let onCardSelect: (index: number) => void;
  export let selectedIndices: number[] = [];
  export let canSelect: boolean = true;
  export let players: any[] = [];
  export let myPlayerId: string = '';
  export let currentRound: number = 0;

  let canvas: HTMLCanvasElement;
  let app: Application;
  let beltContainer: Container;
  let tableContainer: Container;
  let playersContainer: Container;
  let cardsContainer: Container;

  const CANVAS_WIDTH = 1400;
  const CANVAS_HEIGHT = 900;
  const CARD_WIDTH = 90;
  const CARD_HEIGHT = 130;
  const CARD_SPACING = 10;
  const BELT_RADIUS = 320;
  const CENTER_X = CANVAS_WIDTH / 2;
  const CENTER_Y = CANVAS_HEIGHT / 2;
  
  let cardSprites: any[] = [];
  let beltOffset = 0;
  let animationSpeed = 0.05;
  let isAnimating = false;
  let previousHandSize = 0;
  let previousRound = 0;
  let roundAnimationFrame = 0;

  onMount(async () => {
    app = new Application();
    await app.init({
      canvas,
      width: CANVAS_WIDTH,
      height: CANVAS_HEIGHT,
      backgroundColor: 0x8B4513, // Brown wooden table
      antialias: true,
    });

    // Layer containers from back to front
    tableContainer = new Container();
    beltContainer = new Container();
    playersContainer = new Container();
    cardsContainer = new Container();
    
    app.stage.addChild(tableContainer);
    app.stage.addChild(beltContainer);
    app.stage.addChild(playersContainer);
    app.stage.addChild(cardsContainer);

    renderTable();
    renderBelt();
    renderPlayers();
    renderCards();

    // Animation loop for smooth card transitions
    app.ticker.add(() => {
      if (isAnimating) {
        beltOffset += animationSpeed;
        if (beltOffset >= 1) {
          beltOffset = 0;
          isAnimating = false;
        }
        updateCardPositions();
      }
      
      // Round change animation
      if (roundAnimationFrame > 0) {
        roundAnimationFrame--;
        if (roundAnimationFrame % 10 < 5) {
          // Flash effect
          app.stage.alpha = 0.8;
        } else {
          app.stage.alpha = 1;
        }
        if (roundAnimationFrame === 0) {
          app.stage.alpha = 1;
        }
      }
    });
  });

  onDestroy(() => {
    if (app) {
      app.destroy(true, { children: true });
    }
  });

  function renderTable() {
    // Draw wooden table base
    const tableGraphics = new Graphics();
    tableGraphics.ellipse(CENTER_X, CENTER_Y, 380, 320);
    tableGraphics.fill(0x654321); // Darker brown
    tableContainer.addChild(tableGraphics);
  }

  function renderBelt() {
    if (!beltContainer) return;
    beltContainer.removeChildren();

    // Draw conveyor belt track (ellipse around the table)
    const beltTrack = new Graphics();
    beltTrack.ellipse(CENTER_X, CENTER_Y, BELT_RADIUS, BELT_RADIUS * 0.85);
    beltTrack.stroke({ width: 12, color: 0x2a2a2a }); // Dark gray belt outline
    beltContainer.addChild(beltTrack);

    // Belt movement indicators (little arrows)
    for (let i = 0; i < 16; i++) {
      const angle = (i * 22.5) * Math.PI / 180;
      const x = CENTER_X + Math.cos(angle) * BELT_RADIUS;
      const y = CENTER_Y + Math.sin(angle) * BELT_RADIUS * 0.85;
      
      const arrow = new Graphics();
      arrow.moveTo(0, 0);
      arrow.lineTo(8, 4);
      arrow.lineTo(0, 8);
      arrow.fill(0x555555);
      arrow.x = x;
      arrow.y = y;
      arrow.rotation = angle + Math.PI / 2;
      beltContainer.addChild(arrow);
    }
  }

  // Get player position around table - evenly spaced from player perspective
  function getPlayerPosition(playerIndex: number, totalPlayers: number) {
    // Find my index
    const myIndex = players.findIndex(p => p.id === myPlayerId);
    
    // Calculate relative position from my perspective (I'm always at bottom)
    let relativeIndex = (playerIndex - myIndex + totalPlayers) % totalPlayers;
    
    // Calculate angle - I'm at bottom (90 degrees), others evenly spaced
    let angle: number;
    if (totalPlayers === 2) {
      // 2 players: opposite each other (bottom and top)
      angle = relativeIndex === 0 ? Math.PI / 2 : -Math.PI / 2;
    } else if (totalPlayers === 3) {
      // 3 players: equidistant
      const baseAngle = Math.PI / 2; // Start at bottom
      angle = baseAngle + (relativeIndex * 2 * Math.PI / 3);
    } else if (totalPlayers === 4) {
      // 4 players: cardinal directions
      angle = Math.PI / 2 + (relativeIndex * Math.PI / 2);
    } else {
      // 5+ players: evenly distributed
      angle = Math.PI / 2 + (relativeIndex * 2 * Math.PI / totalPlayers);
    }
    
    const distance = 420;
    return {
      x: CENTER_X + Math.cos(angle) * distance,
      y: CENTER_Y + Math.sin(angle) * distance,
      angle: angle
    };
  }
  function renderPlayers() {
    if (!playersContainer) return;
    playersContainer.removeChildren();
    
    const numPlayers = players.length;
    
    players.forEach((player, index) => {
      const pos = getPlayerPosition(index, numPlayers);
      const isMe = player.id === myPlayerId;

      // Player area (rectangular table section)
      const playerArea = new Graphics();
      playerArea.roundRect(-90, -50, 180, 100, 10);
      
      if (isMe) {
        playerArea.fill(0xFFD700); // Gold for current player
      } else if (player.hasSelected) {
        playerArea.fill(0x4CAF50); // Green if selected
      } else {
        playerArea.fill(0x8B6914); // Light wood
      }
      
      playerArea.stroke({ width: 3, color: 0x000000 });
      playerArea.x = pos.x;
      playerArea.y = pos.y;
      playersContainer.addChild(playerArea);

      // Player name
      const nameText = new Text({
        text: isMe ? `${player.name} (You)` : player.name,
        style: {
          fontSize: 16,
          fill: 0xffffff,
          fontWeight: 'bold',
          stroke: { color: 0x000000, width: 3 }
        }
      });
      nameText.anchor.set(0.5);
      nameText.x = pos.x;
      nameText.y = pos.y - 20;
      playersContainer.addChild(nameText);

      // Player score
      const scoreText = new Text({
        text: `Score: ${player.score}`,
        style: {
          fontSize: 14,
          fill: 0xffffff,
        }
      });
      scoreText.anchor.set(0.5);
      scoreText.x = pos.x;
      scoreText.y = pos.y;
      playersContainer.addChild(scoreText);

      // Hand size
      const handText = new Text({
        text: `Hand: ${player.handSize || 0}`,
        style: {
          fontSize: 12,
          fill: 0xffffff,
        }
      });
      handText.anchor.set(0.5);
      handText.x = pos.x;
      handText.y = pos.y + 20;
      playersContainer.addChild(handText);

      // Selection indicator
      if (player.hasSelected) {
        const checkmark = new Text({
          text: '✅',
          style: { fontSize: 32 }
        });
        checkmark.anchor.set(0.5);
        checkmark.x = pos.x + 70;
        checkmark.y = pos.y - 25;
        playersContainer.addChild(checkmark);
      }
    });
  }

  function renderCards() {
    if (!cardsContainer) return;
    
    cardSprites.forEach(sprite => {
      if (sprite.container) {
        sprite.container.parent?.removeChild(sprite.container);
      }
    });
    cardSprites = [];
    cardsContainer.removeChildren();

    // Render my cards (front-facing, clickable, in hand layout)
    renderMyCards();
    
    // Render other players' cards (back-facing, positioned in front of them)
    renderOtherPlayersCards();
  }

  function renderMyCards() {
    const myIndex = players.findIndex(p => p.id === myPlayerId);
    if (myIndex === -1) return;
    
    const pos = getPlayerPosition(myIndex, players.length);
    const numCards = cards.length;
    
    cards.forEach((card, index) => {
      const isSelected = selectedIndices.includes(index);
      
      const cardContainer = new Container();
      
      // Create card graphics
      const cardGraphics = new Graphics();
      cardGraphics.roundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 8);
      cardGraphics.fill(getCardColor(card.type));
      cardGraphics.stroke({ width: 3, color: 0xffffff });
      
      // Card shadow for depth
      const shadow = new Graphics();
      shadow.roundRect(3, 3, CARD_WIDTH, CARD_HEIGHT, 8);
      shadow.fill(0x000000);
      shadow.alpha = 0.3;
      cardContainer.addChild(shadow);
      cardContainer.addChild(cardGraphics);

      // Card emoji
      const emoji = new Text({
        text: getCardEmoji(card.type),
        style: {
          fontSize: 36,
          fill: 0xffffff,
        }
      });
      emoji.x = CARD_WIDTH / 2;
      emoji.y = CARD_HEIGHT / 2 - 10;
      emoji.anchor.set(0.5);
      cardContainer.addChild(emoji);

      // Card type text
      const cardText = new Text({
        text: formatCardType(card.type),
        style: {
          fontSize: 10,
          fill: 0xffffff,
          fontWeight: 'bold',
        }
      });
      cardText.x = CARD_WIDTH / 2;
      cardText.y = CARD_HEIGHT - 15;
      cardText.anchor.set(0.5);
      cardContainer.addChild(cardText);

      cardContainer.pivot.set(CARD_WIDTH / 2, CARD_HEIGHT / 2);
      
      // Card is always clickable - for selection and deselection
      if (canSelect) {
        cardGraphics.eventMode = 'static';
        cardGraphics.cursor = 'pointer';

        cardGraphics.on('pointerdown', () => {
          onCardSelect(index);
        });

        cardGraphics.on('pointerover', () => {
          if (!isSelected) {
            cardContainer.scale.set(1.15);
            cardContainer.zIndex = 1000;
          }
        });

        cardGraphics.on('pointerout', () => {
          if (!isSelected) {
            cardContainer.scale.set(1);
            cardContainer.zIndex = index;
          }
        });
      }

      if (isSelected) {
        // Selected card glows
        const glow = new Graphics();
        glow.roundRect(-5, -5, CARD_WIDTH + 10, CARD_HEIGHT + 10, 10);
        glow.stroke({ width: 4, color: 0xFFD700 });
        glow.alpha = 0.8;
        cardContainer.addChildAt(glow, 0);
      }
      
      // Position cards in a hand/fan layout in front of me on the belt
      const totalWidth = numCards * (CARD_WIDTH + CARD_SPACING);
      const startX = CENTER_X - totalWidth / 2 + CARD_WIDTH / 2;
      
      // Fan effect: slight rotation and curve
      const fanAngle = (index - (numCards - 1) / 2) * 3; // 3 degrees per card from center
      const fanCurve = Math.abs(index - (numCards - 1) / 2) * 5; // Slight curve up for outer cards
      
      cardContainer.x = startX + index * (CARD_WIDTH + CARD_SPACING);
      cardContainer.y = pos.y - 120 - fanCurve; // Position on belt in front of player
      cardContainer.rotation = (fanAngle * Math.PI) / 180;
      
      if (isSelected) {
        cardContainer.y -= 20; // Lift selected cards
      }
      
      cardsContainer.addChild(cardContainer);
      cardSprites.push({ 
        container: cardContainer, 
        index, 
        isSelected,
        isMyCard: true
      });
    });
  }

  function renderOtherPlayersCards() {
    players.forEach((player, playerIndex) => {
      if (player.id === myPlayerId) return; // Skip my cards
      
      const pos = getPlayerPosition(playerIndex, players.length);
      const numCards = player.handSize || 0;
      
      for (let i = 0; i < numCards; i++) {
        const cardContainer = new Container();
        
        // Create card back (not revealing what it is)
        const cardGraphics = new Graphics();
        cardGraphics.roundRect(0, 0, CARD_WIDTH * 0.8, CARD_HEIGHT * 0.8, 6);
        cardGraphics.fill(0x2C3E50); // Dark blue/gray card back
        cardGraphics.stroke({ width: 2, color: 0xffffff });
        
        // Card back pattern
        const backPattern = new Text({
          text: '🎴',
          style: {
            fontSize: 30,
            fill: 0xffffff,
          }
        });
        backPattern.x = (CARD_WIDTH * 0.8) / 2;
        backPattern.y = (CARD_HEIGHT * 0.8) / 2;
        backPattern.anchor.set(0.5);
        cardContainer.addChild(cardGraphics);
        cardContainer.addChild(backPattern);

        cardContainer.pivot.set((CARD_WIDTH * 0.8) / 2, (CARD_HEIGHT * 0.8) / 2);
        
        // Position in a hand layout in front of the other player
        const totalWidth = numCards * ((CARD_WIDTH * 0.8) + CARD_SPACING);
        const startX = pos.x - totalWidth / 2 + (CARD_WIDTH * 0.8) / 2;
        
        const fanAngle = (i - (numCards - 1) / 2) * 2;
        const fanCurve = Math.abs(i - (numCards - 1) / 2) * 3;
        
        // Calculate position relative to player's angle
        const angle = pos.angle;
        const distanceFromPlayer = 120;
        
        cardContainer.x = pos.x + Math.cos(angle - Math.PI / 2) * distanceFromPlayer + (i - (numCards - 1) / 2) * (CARD_WIDTH * 0.7);
        cardContainer.y = pos.y + Math.sin(angle - Math.PI / 2) * distanceFromPlayer - fanCurve;
        cardContainer.rotation = angle + (fanAngle * Math.PI) / 180;
        cardContainer.zIndex = -100; // Put other players' cards behind player names
        
        cardsContainer.addChild(cardContainer);
        cardSprites.push({ 
          container: cardContainer, 
          index: i, 
          isSelected: false,
          isMyCard: false,
          playerId: player.id
        });
      }
    });
  }
  function updateCardPositions() {
    // Animation when hands are passed - cards move along belt
    if (!isAnimating) return;
    
    cardSprites.forEach((sprite) => {
      if (!sprite.isMyCard) {
        // Animate other players' cards fading and repositioning
        const fadeAmount = Math.sin(beltOffset * Math.PI);
        sprite.container.alpha = 1 - (fadeAmount * 0.3); // Slight fade
        // Rotate cards slightly as they "move"
        const rotationOffset = beltOffset * 0.2;
        sprite.container.rotation += rotationOffset;
      }
    });
  }

  // Trigger animation when hand size changes (cards passed to next player)
  $: if (cards.length !== previousHandSize) {
    if (previousHandSize > 0 && cards.length < previousHandSize) {
      // Hand changed (cards passed), animate conveyor belt
      isAnimating = true;
      beltOffset = 0;
    }
    previousHandSize = cards.length;
  }
  
  // Trigger round change animation
  $: if (currentRound !== previousRound && previousRound > 0) {
    // Round changed, trigger flash animation
    roundAnimationFrame = 30; // 30 frames of animation (~0.5 seconds at 60fps)
    previousRound = currentRound;
  } else if (previousRound === 0 && currentRound > 0) {
    previousRound = currentRound;
  }

  function formatCardType(type: string): string {
    return type.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
  }

  function getCardColor(type: string): number {
    const colors: Record<string, number> = {
      maki_roll: 0xff6b6b,
      tempura: 0xffa726,
      sashimi: 0xf48fb1,
      dumpling: 0xba68c8,
      nigiri: 0x4fc3f7,
      wasabi: 0x66bb6a,
      chopsticks: 0xffd54f,
      pudding: 0xffb74d
    };
    return colors[type] || 0x9e9e9e;
  }

  function getCardEmoji(type: string): string {
    const emojis: Record<string, string> = {
      maki_roll: '🍣',
      tempura: '🍤',
      sashimi: '🐟',
      dumpling: '🥟',
      nigiri: '🍱',
      wasabi: '🟢',
      chopsticks: '🥢',
      pudding: '🍮'
    };
    return emojis[type] || '🎴';
  }

  // Start belt animation
  export function startBeltAnimation() {
    isAnimating = true;
    beltOffset = 0;
  }

  export function stopBeltAnimation() {
    isAnimating = false;
  }

  $: if (cards || selectedIndices || players) {
    if (app) {
      renderPlayers();
      renderCards();
    }
  }
</script>

<canvas bind:this={canvas}></canvas>

<style>
  canvas {
    display: block;
    width: 100%;
    height: 100vh;
    max-height: 900px;
    margin: 0 auto;
    border-radius: 12px;
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.5);
    border: 4px solid #3E2723;
  }
</style>
