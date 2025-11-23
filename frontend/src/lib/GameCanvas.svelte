<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Application, Graphics, Text, Container } from 'pixi.js';
  import type { GameState, Card } from './websocket';

  export let gameState: GameState | null;
  export let onSelectCard: (index: number) => void;
  export let selectedCardIndex: number | null = null;
  export let secondCardIndex: number | null = null;

  let canvasContainer: HTMLDivElement;
  let app: Application | null = null;
  let cardsContainer: Container | null = null;
  let isAnimating = false;
  let isInitialized = false;

  const CARD_WIDTH = 120;
  const CARD_HEIGHT = 80;
  const CARD_SPACING = 15;
  const CARD_COLORS = {
    'maki_roll': 0xff6b6b,
    'tempura': 0xffa726,
    'sashimi': 0xec407a,
    'dumpling': 0xab47bc,
    'nigiri': 0x26c6da,
    'wasabi': 0x66bb6a,
    'chopsticks': 0xffd54f,
    'pudding': 0xffb74d,
    'default': 0x78909c
  };

  onMount(async () => {
    if (isInitialized) return;
    
    try {
      // Create PixiJS application
      app = new Application();
      await app.init({
        width: canvasContainer?.clientWidth || 800,
        height: 400,
        backgroundColor: 0xffffff,
        antialias: true,
      });

      if (canvasContainer && app.canvas) {
        canvasContainer.appendChild(app.canvas);
      }

      // Create container for cards
      cardsContainer = new Container();
      if (app.stage) {
        app.stage.addChild(cardsContainer);
      }

      isInitialized = true;

      // Handle window resize
      const handleResize = () => {
        if (app && canvasContainer && app.renderer) {
          app.renderer.resize(canvasContainer.clientWidth, 400);
          updateCards();
        }
      };
      window.addEventListener('resize', handleResize);

      return () => {
        window.removeEventListener('resize', handleResize);
      };
    } catch (error) {
      console.error('Error initializing PixiJS:', error);
    }
  });

  onDestroy(() => {
    if (app && isInitialized) {
      try {
        app.destroy(true, { children: true, texture: true, textureSource: true });
      } catch (error) {
        console.error('Error destroying PixiJS app:', error);
      }
      app = null;
      cardsContainer = null;
      isInitialized = false;
    }
  });

  $: if (gameState && app && cardsContainer && isInitialized) {
    updateCards();
  }

  function updateCards() {
    if (!gameState || !cardsContainer || isAnimating || !app) return;

    // Clear existing cards
    cardsContainer.removeChildren();

    const myHand = gameState.myHand || [];
    if (myHand.length === 0) {
      const text = new Text({ text: 'No cards in hand', style: { fill: 0x666666, fontSize: 18 } });
      text.x = app.screen.width / 2;
      text.y = app.screen.height / 2;
      text.anchor.set(0.5);
      cardsContainer.addChild(text);
      return;
    }

    // Calculate starting position to center cards
    const totalWidth = myHand.length * (CARD_WIDTH + CARD_SPACING) - CARD_SPACING;
    const startX = (app.screen.width - totalWidth) / 2;
    const startY = (app.screen.height - CARD_HEIGHT) / 2;

    // Draw cards
    myHand.forEach((card, index) => {
      const cardGraphic = createCard(card, index);
      cardGraphic.x = startX + index * (CARD_WIDTH + CARD_SPACING);
      cardGraphic.y = startY;
      cardsContainer!.addChild(cardGraphic);
    });
  }

  function createCard(card: Card, index: number): Container {
    const cardContainer = new Container();
    cardContainer.eventMode = 'static';
    cardContainer.cursor = 'pointer';

    // Check if this card is selected
    const isSelected = selectedCardIndex === index || secondCardIndex === index;

    // Card background
    const bg = new Graphics();
    const color = CARD_COLORS[card.type as keyof typeof CARD_COLORS] || CARD_COLORS.default;
    
    // Use modern Graphics API
    bg.rect(0, 0, CARD_WIDTH, CARD_HEIGHT);
    bg.fill(color);

    // Add selection border
    if (isSelected) {
      bg.rect(0, 0, CARD_WIDTH, CARD_HEIGHT);
      bg.stroke({ width: 3, color: 0xffd700 });
    }

    cardContainer.addChild(bg);

    // Card type text
    const cardType = formatCardType(card.type);
    const typeText = new Text({ 
      text: cardType, 
      style: { 
        fill: 0xffffff, 
        fontSize: 14,
        fontWeight: 'bold',
        wordWrap: true,
        wordWrapWidth: CARD_WIDTH - 10,
      } 
    });
    typeText.x = CARD_WIDTH / 2;
    typeText.y = CARD_HEIGHT / 2 - 10;
    typeText.anchor.set(0.5);
    cardContainer.addChild(typeText);

    // Card variant/value text
    if (card.variant || card.value !== undefined) {
      const variantText = new Text({ 
        text: card.variant || `Value: ${card.value}`,
        style: { 
          fill: 0xffffff, 
          fontSize: 10,
        } 
      });
      variantText.x = CARD_WIDTH / 2;
      variantText.y = CARD_HEIGHT / 2 + 10;
      variantText.anchor.set(0.5);
      cardContainer.addChild(variantText);
    }

    // Add emoji based on card type
    const emoji = getCardEmoji(card.type);
    if (emoji) {
      const emojiText = new Text({ 
        text: emoji,
        style: { fontSize: 20 }
      });
      emojiText.x = 10;
      emojiText.y = 10;
      cardContainer.addChild(emojiText);
    }

    // Handle click
    cardContainer.on('pointerdown', () => {
      if (!isAnimating) {
        onSelectCard(index);
      }
    });

    // Hover effect
    cardContainer.on('pointerover', () => {
      cardContainer.y -= 5;
    });
    cardContainer.on('pointerout', () => {
      cardContainer.y += 5;
    });

    return cardContainer;
  }

  function formatCardType(type: string): string {
    return type.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
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
</script>

<div bind:this={canvasContainer} class="w-full min-h-[400px] rounded-lg overflow-hidden bg-white"></div>
