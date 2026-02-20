import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';

export function useSSE(boardId: string) {
  const qc = useQueryClient();

  useEffect(() => {
    if (!boardId) return;

    const es = new EventSource(`/api/v1/boards/${boardId}/events`);

    const handler = () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
    };

    es.addEventListener('card.created', handler);
    es.addEventListener('card.updated', handler);
    es.addEventListener('card.moved', handler);
    es.addEventListener('card.deleted', handler);
    es.addEventListener('list.created', handler);
    es.addEventListener('list.updated', handler);
    es.addEventListener('list.deleted', handler);
    es.addEventListener('label.created', handler);
    es.addEventListener('label.updated', handler);
    es.addEventListener('label.deleted', handler);
    es.addEventListener('activity.new', handler);

    return () => es.close();
  }, [boardId, qc]);
}
