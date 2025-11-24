<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Application, Graphics, Text, Container } from 'pixi.js';
  import type { Card } from './websocket';

  export let cards: Card[] = [];
  export let onCardSelect: (index: number) => void;
  export let selectedIndices: number[] = [];
  export let canSelect: boolean = true;

  let canvas: HTMLCanvasElement;
  let app: Application;
  let cardsContainer: Container;

  const CARD_WIDTH = 120;
  const CARD_HEIGHT = 160;
  const CARD_SPACING = 20;
  const BELT_Y = 250;
  
  let cardSprites: any[] = [];

  onMount(async () => {
    app = new Application();
    await app.init({
      canvas,
      width: 1200,
      height: 700,
      backgroundColor: 0xFFF8E1,
      antialias: true,
    });

    cardsContainer = new Container();
    app.stage.addChild(cardsContainer);

    renderCards();
  });

  onDestroy(() => {
    if (app) {
      app.destroy(true, { children: true });
    }
  });

  function renderCards() {
    cardSprites.forEach(sprite => {
      cardsContainer.removeChild(sprite.graphics);
      cardsContainer.removeChild(sprite.text);
    });
    cardSprites = [];

    cards.forEach((card, index) => {
      const isSelected = selectedIndices.includes(index);
      const x = 100 + index * (CARD_WIDTH + CARD_SPACING);
      const y = isSelected ? 500 : BELT_Y;

      const cardGraphics = new Graphics();
      cardGraphics.roundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 10);
      cardGraphics.fill(getCardColor(card.type));
      cardGraphics.stroke({ width: 4, color: 0xffffff });
      
      cardGraphics.x = x;
      cardGraphics.y = y;
      cardGraphics.eventMode = 'static';
      cardGraphics.cursor = 'pointer';

      if (canSelect) {
        cardGraphics.on('pointerdown', () => {
          onCardSelect(index);
        });

        cardGraphics.on('pointerover', () => {
          cardGraphics.scale.set(1.1);
          cardGraphics.y = y - 20;
        });

        cardGraphics.on('pointerout', () => {
          cardGraphics.scale.set(1);
          cardGraphics.y = y;
        });
      }

      const text = new Text({
        text: getCardEmoji(card.type),
        style: {
          fontSize: 48,
          fill: 0xffffff,
        }
      });
      text.x = x + CARD_WIDTH / 2;
      text.y = y + CARD_HEIGHT / 2;
      text.anchor.set(0.5);

      cardsContainer.addChild(cardGraphics);
      cardsContainer.addChild(text);

      cardSprites.push({ graphics: cardGraphics, text });
    });
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

  $: if (cards) {
    renderCards();
  }
</script>

<canvas bind:this={canvas}></canvas>

<style>
  canvas {
    display: block;
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }
</style>
