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

  let canvas: HTMLCanvasElement;
  let app: Application;
  let beltContainer: Container;
  let tableContainer: Container;
  let playersContainer: Container;
  let myTableContainer: Container;

  const CANVAS_WIDTH = 1400;
  const CANVAS_HEIGHT = 900;
  const CARD_WIDTH = 100;
  const CARD_HEIGHT = 140;
  const CARD_SPACING = 15;
  const BELT_RADIUS = 280;
  const CENTER_X = CANVAS_WIDTH / 2;
  const CENTER_Y = CANVAS_HEIGHT / 2;
  
  let cardSprites: any[] = [];
  let beltOffset = 0;
  let animationSpeed = 0.3;
  let isAnimating = false;

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
    beltContainer = new Container();
    playersContainer = new Container();
    tableContainer = new Container();
    myTableContainer = new Container();
    
    app.stage.addChild(beltContainer);
    app.stage.addChild(playersContainer);
    app.stage.addChild(tableContainer);
    app.stage.addChild(myTableContainer);

    renderTable();
    renderPlayers();
    renderBelt();
    renderCards();

    // Animation loop for belt movement
    app.ticker.add(() => {
      if (isAnimating) {
        beltOffset += animationSpeed;
        if (beltOffset >= 360) beltOffset -= 360;
        updateCardPositions();
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
    tableGraphics.ellipse(CENTER_X, CENTER_Y, 350, 300);
    tableGraphics.fill(0x654321); // Darker brown
    tableContainer.addChild(tableGraphics);

    // Draw conveyor belt track
    const beltTrack = new Graphics();
    beltTrack.ellipse(CENTER_X, CENTER_Y, BELT_RADIUS, BELT_RADIUS * 0.85);
    beltTrack.fill(0x2a2a2a); // Dark gray belt
    beltTrack.stroke({ width: 8, color: 0x1a1a1a });
    beltContainer.addChild(beltTrack);

    // Belt movement indicators (little arrows)
    for (let i = 0; i < 12; i++) {
      const angle = (i * 30) * Math.PI / 180;
      const x = CENTER_X + Math.cos(angle) * BELT_RADIUS;
      const y = CENTER_Y + Math.sin(angle) * BELT_RADIUS * 0.85;
      
      const arrow = new Graphics();
      arrow.moveTo(0, 0);
      arrow.lineTo(10, 5);
      arrow.lineTo(0, 10);
      arrow.fill(0x444444);
      arrow.x = x;
      arrow.y = y;
      arrow.rotation = angle + Math.PI / 2;
      beltContainer.addChild(arrow);
    }
  }

  function renderPlayers() {
    playersContainer.removeChildren();
    
    const numPlayers = players.length;
    const angleStep = (2 * Math.PI) / numPlayers;
    
    players.forEach((player, index) => {
      const angle = index * angleStep - Math.PI / 2; // Start at top
      const distance = 450;
      const x = CENTER_X + Math.cos(angle) * distance;
      const y = CENTER_Y + Math.sin(angle) * distance;

      // Player area (rectangular table section)
      const playerArea = new Graphics();
      playerArea.roundRect(-80, -40, 160, 80, 10);
      
      const isMe = player.id === myPlayerId;
      if (isMe) {
        playerArea.fill(0xFFD700); // Gold for current player
      } else if (player.hasSelected) {
        playerArea.fill(0x4CAF50); // Green if selected
      } else {
        playerArea.fill(0x8B6914); // Light wood
      }
      
      playerArea.stroke({ width: 3, color: 0x000000 });
      playerArea.x = x;
      playerArea.y = y;
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
      nameText.x = x;
      nameText.y = y - 15;
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
      scoreText.x = x;
      scoreText.y = y + 5;
      playersContainer.addChild(scoreText);

      // Selection indicator
      if (player.hasSelected) {
        const checkmark = new Text({
          text: '✅',
          style: { fontSize: 32 }
        });
        checkmark.anchor.set(0.5);
        checkmark.x = x;
        checkmark.y = y + 25;
        playersContainer.addChild(checkmark);
      }
    });
  }

  function renderBelt() {
    // Belt is already rendered in renderTable
  }

  function renderCards() {
    cardSprites.forEach(sprite => {
      if (sprite.container) {
        sprite.container.parent?.removeChild(sprite.container);
      }
    });
    cardSprites = [];

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
          fontSize: 40,
          fill: 0xffffff,
        }
      });
      emoji.x = CARD_WIDTH / 2;
      emoji.y = CARD_HEIGHT / 2;
      emoji.anchor.set(0.5);
      cardContainer.addChild(emoji);

      // Card type text
      const cardText = new Text({
        text: formatCardType(card.type),
        style: {
          fontSize: 11,
          fill: 0xffffff,
          fontWeight: 'bold',
        }
      });
      cardText.x = CARD_WIDTH / 2;
      cardText.y = CARD_HEIGHT - 15;
      cardText.anchor.set(0.5);
      cardContainer.addChild(cardText);

      cardContainer.pivot.set(CARD_WIDTH / 2, CARD_HEIGHT / 2);
      
      if (canSelect && !isSelected) {
        cardGraphics.eventMode = 'static';
        cardGraphics.cursor = 'pointer';

        cardGraphics.on('pointerdown', () => {
          onCardSelect(index);
        });

        cardGraphics.on('pointerover', () => {
          cardContainer.scale.set(1.15);
          cardContainer.zIndex = 1000;
        });

        cardGraphics.on('pointerout', () => {
          cardContainer.scale.set(1);
          cardContainer.zIndex = index;
        });
      }

      if (isSelected) {
        // Move to my table area (bottom of screen)
        cardContainer.x = CENTER_X - (cards.length * (CARD_WIDTH + 10)) / 2 + index * (CARD_WIDTH + 10) + CARD_WIDTH / 2;
        cardContainer.y = CANVAS_HEIGHT - 150;
        cardContainer.scale.set(1.1);
        
        // Add glow effect for selected cards
        const glow = new Graphics();
        glow.roundRect(-5, -5, CARD_WIDTH + 10, CARD_HEIGHT + 10, 10);
        glow.stroke({ width: 4, color: 0xFFD700 });
        glow.alpha = 0.8;
        cardContainer.addChildAt(glow, 0);
        
        myTableContainer.addChild(cardContainer);
      } else {
        beltContainer.addChild(cardContainer);
      }

      cardSprites.push({ 
        container: cardContainer, 
        index, 
        originalAngle: (index / cards.length) * 2 * Math.PI,
        isSelected 
      });
    });

    updateCardPositions();
  }

  function updateCardPositions() {
    cardSprites.forEach((sprite) => {
      if (sprite.isSelected) return; // Selected cards stay on table
      
      const angleOffset = (beltOffset * Math.PI) / 180;
      const angle = sprite.originalAngle + angleOffset;
      
      const x = CENTER_X + Math.cos(angle) * BELT_RADIUS;
      const y = CENTER_Y + Math.sin(angle) * BELT_RADIUS * 0.85;
      
      sprite.container.x = x;
      sprite.container.y = y;
      sprite.container.rotation = angle + Math.PI / 2;
    });
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
    max-width: 1400px;
    margin: 0 auto;
    border-radius: 12px;
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.5);
    border: 4px solid #3E2723;
  }
</style>
