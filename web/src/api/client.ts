import type { Board, Card, Label, ActivityLog, List } from './types';

const BASE = '/api/v1';

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  if (res.status === 204) return null as T;
  return res.json();
}

export const api = {
  boards: {
    list: (): Promise<Board[]> =>
      fetch(`${BASE}/boards`).then(r => json(r)),
    create: (data: { name: string; description?: string }): Promise<Board> =>
      fetch(`${BASE}/boards`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    get: (id: string): Promise<Board & { lists: List[] }> =>
      fetch(`${BASE}/boards/${id}`).then(r => json(r)),
    update: (id: string, data: Partial<Board>): Promise<Board> =>
      fetch(`${BASE}/boards/${id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    delete: (id: string): Promise<void> =>
      fetch(`${BASE}/boards/${id}`, { method: 'DELETE' }).then(r => json(r)),
  },
  lists: {
    create: (boardId: string, data: { name: string; position?: number }): Promise<List> =>
      fetch(`${BASE}/boards/${boardId}/lists`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    update: (id: string, data: { name?: string; position?: number }): Promise<List> =>
      fetch(`${BASE}/lists/${id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    delete: (id: string): Promise<void> =>
      fetch(`${BASE}/lists/${id}`, { method: 'DELETE' }).then(r => json(r)),
  },
  cards: {
    create: (listId: string, data: { title: string; description?: string; assignee?: string; priority?: string; position?: number }): Promise<Card> =>
      fetch(`${BASE}/lists/${listId}/cards`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    get: (id: string): Promise<Card> =>
      fetch(`${BASE}/cards/${id}`).then(r => json(r)),
    update: (id: string, data: Record<string, unknown>): Promise<Card> =>
      fetch(`${BASE}/cards/${id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    delete: (id: string): Promise<void> =>
      fetch(`${BASE}/cards/${id}`, { method: 'DELETE' }).then(r => json(r)),
    move: (id: string, data: { list_id: string; position: number }): Promise<Card> =>
      fetch(`${BASE}/cards/${id}/move`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    assign: (id: string, assignee: string): Promise<Card> =>
      fetch(`${BASE}/cards/${id}/assign`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ assignee }) }).then(r => json(r)),
  },
  dependencies: {
    add: (cardId: string, dependsOnCardId: string): Promise<void> =>
      fetch(`${BASE}/cards/${cardId}/dependencies`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ depends_on_card_id: dependsOnCardId }) }).then(r => json(r)),
    remove: (cardId: string, depId: string): Promise<void> =>
      fetch(`${BASE}/cards/${cardId}/dependencies/${depId}`, { method: 'DELETE' }).then(r => json(r)),
  },
  labels: {
    list: (boardId: string): Promise<Label[]> =>
      fetch(`${BASE}/boards/${boardId}/labels`).then(r => json(r)),
    create: (boardId: string, data: { name: string; color: string }): Promise<Label> =>
      fetch(`${BASE}/boards/${boardId}/labels`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    update: (id: string, data: Partial<Label>): Promise<Label> =>
      fetch(`${BASE}/labels/${id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => json(r)),
    delete: (id: string): Promise<void> =>
      fetch(`${BASE}/labels/${id}`, { method: 'DELETE' }).then(r => json(r)),
    addToCard: (cardId: string, labelId: string): Promise<void> =>
      fetch(`${BASE}/cards/${cardId}/labels`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ label_id: labelId }) }).then(r => json(r)),
    removeFromCard: (cardId: string, labelId: string): Promise<void> =>
      fetch(`${BASE}/cards/${cardId}/labels/${labelId}`, { method: 'DELETE' }).then(r => json(r)),
  },
  activity: {
    byCard: (cardId: string, limit = 50): Promise<ActivityLog[]> =>
      fetch(`${BASE}/cards/${cardId}/activity?limit=${limit}`).then(r => json(r)),
    byBoard: (boardId: string, limit = 50): Promise<ActivityLog[]> =>
      fetch(`${BASE}/boards/${boardId}/activity?limit=${limit}`).then(r => json(r)),
  },
  search: (boardId: string, params: { q?: string; assignee?: string; status?: string; label?: string }): Promise<Card[]> => {
    const qs = new URLSearchParams(params as Record<string, string>).toString();
    return fetch(`${BASE}/boards/${boardId}/search?${qs}`).then(r => json(r));
  },
};
