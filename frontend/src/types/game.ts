export type CardType = 
  | 'maki_roll'
  | 'tempura'
  | 'sashimi'
  | 'dumpling'
  | 'nigiri'
  | 'wasabi'
  | 'chopsticks'
  | 'pudding';

export type RoundPhase = 
  | 'waiting'
  | 'selecting'
  | 'revealing'
  | 'passing'
  | 'scoring'
  | 'round_end'
  | 'game_end';

export interface Card {
  id: string;
  type: CardType;
  variant?: string;
  value?: number;
}

export interface PlayerState {
  id: string;
  name: string;
  handSize: number;
  collection: Card[];
  score: number;
  hasSelected: boolean;
}

export interface GameState {
  gameId: string;
  players: PlayerState[];
  currentRound: number;
  phase: RoundPhase;
  myPlayerId: string;
}
