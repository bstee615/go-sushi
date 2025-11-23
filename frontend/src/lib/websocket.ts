import { writable } from 'svelte/store';

export interface GameState {
  gameId?: string;
  myPlayerId?: string;
  phase?: string;
  currentRound?: number;
  current_round?: number;
  players?: Player[];
  myHand?: Card[];
}

export interface Player {
  id: string;
  name: string;
  score: number;
  handSize: number;
  hasSelected: boolean;
  collection?: Card[];
  puddingCards?: Card[];
  roundScores?: number[];
  chopsticksCount?: number;
}

export interface Card {
  type: string;
  variant?: string;
  value?: number;
}

export interface Message {
  type: string;
  payload: any;
}

class WebSocketStore {
  private ws: WebSocket | null = null;
  public connected = writable(false);
  public gameState = writable<GameState | null>(null);
  public error = writable<string | null>(null);
  private messageHandlers: Map<string, (payload: any) => void> = new Map();

  connect(url: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
          console.log('Connected to server');
          this.connected.set(true);
          this.error.set(null);
          resolve();
        };

        this.ws.onclose = () => {
          console.log('Disconnected from server');
          this.connected.set(false);
          
          // Try to reconnect after 2 seconds
          setTimeout(() => {
            if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
              this.connect(url);
            }
          }, 2000);
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.error.set('Connection error');
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };
      } catch (error) {
        console.error('Connection error:', error);
        this.error.set('Failed to connect');
        reject(error);
      }
    });
  }

  private handleMessage(data: string) {
    try {
      const message: Message = JSON.parse(data);
      console.log('Received:', message.type, message.payload);

      // Call specific handlers
      const handler = this.messageHandlers.get(message.type);
      if (handler) {
        handler(message.payload);
      }

      // Update game state for game_state messages
      if (message.type === 'game_state') {
        this.gameState.set(message.payload);
      } else if (message.type === 'error') {
        this.error.set(message.payload.message || JSON.stringify(message.payload));
      }
    } catch (error) {
      console.error('Error parsing message:', error);
    }
  }

  sendMessage(type: string, payload: any) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('Cannot send message: not connected');
      return;
    }

    const message: Message = { type, payload };
    this.ws.send(JSON.stringify(message));
    console.log('Sent:', type, payload);
  }

  onMessage(type: string, handler: (payload: any) => void) {
    this.messageHandlers.set(type, handler);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected.set(false);
    this.gameState.set(null);
  }
}

export const wsStore = new WebSocketStore();
