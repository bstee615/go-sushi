export type MessageType = 
  | 'join_game'
  | 'start_game'
  | 'select_card'
  | 'game_state'
  | 'card_revealed'
  | 'round_end'
  | 'game_end'
  | 'error';

export interface Message {
  type: MessageType;
  payload: unknown;
}

export interface SelectCardPayload {
  cardIndex: number;
  useChopsticks: boolean;
  secondCardIndex?: number;
}
