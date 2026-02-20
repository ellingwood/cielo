export interface Board {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  lists?: List[];
}

export interface List {
  id: string;
  board_id: string;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
  cards: Card[];
}

export interface Card {
  id: string;
  list_id: string;
  title: string;
  description: string;
  position: number;
  assignee: string;
  status: string;
  priority: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
  labels: Label[];
  dependencies: Card[];
  dependents: Card[];
  activity: ActivityLog[];
}

export interface Label {
  id: string;
  board_id: string;
  name: string;
  color: string;
}

export interface ActivityLog {
  id: string;
  card_id: string;
  actor: string;
  action: string;
  detail: string;
  created_at: string;
}

export type CardStatus = 'unassigned' | 'assigned' | 'in_progress' | 'blocked' | 'done';
export type CardPriority = 'low' | 'medium' | 'high' | 'critical';
