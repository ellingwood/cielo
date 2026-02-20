import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from './client';
import type { Board, List } from './types';

export function useBoards() {
  return useQuery({ queryKey: ['boards'], queryFn: api.boards.list });
}

export function useBoard(id: string) {
  return useQuery<Board & { lists: List[] }>({
    queryKey: ['board', id],
    queryFn: () => api.boards.get(id),
    enabled: !!id,
  });
}

export function useCreateBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) => api.boards.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['boards'] }),
  });
}

export function useDeleteBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.boards.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['boards'] }),
  });
}

export function useCreateList(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; position?: number }) => api.lists.create(boardId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['board', boardId] }),
  });
}

export function useDeleteList(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.lists.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['board', boardId] }),
  });
}

export function useCreateCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { listId: string; title: string; description?: string; priority?: string; position?: number }) => {
      const { listId, ...rest } = data;
      return api.cards.create(listId, rest);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['board', boardId] }),
  });
}

export function useMoveCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { cardId: string; listId: string; position: number }) =>
      api.cards.move(data.cardId, { list_id: data.listId, position: data.position }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['board', boardId] }),
  });
}

export function useUpdateCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; updates: Record<string, unknown> }) =>
      api.cards.update(data.id, data.updates),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
      qc.invalidateQueries({ queryKey: ['card'] });
    },
  });
}

export function useDeleteCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.cards.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['board', boardId] }),
  });
}

export function useCard(id: string) {
  return useQuery({
    queryKey: ['card', id],
    queryFn: () => api.cards.get(id),
    enabled: !!id,
  });
}

export function useLabels(boardId: string) {
  return useQuery({
    queryKey: ['labels', boardId],
    queryFn: () => api.labels.list(boardId),
    enabled: !!boardId,
  });
}

export function useCreateLabel(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => api.labels.create(boardId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels', boardId] }),
  });
}

export function useAddLabelToCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { cardId: string; labelId: string }) => api.labels.addToCard(data.cardId, data.labelId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
      qc.invalidateQueries({ queryKey: ['card'] });
    },
  });
}

export function useRemoveLabelFromCard(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { cardId: string; labelId: string }) => api.labels.removeFromCard(data.cardId, data.labelId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
      qc.invalidateQueries({ queryKey: ['card'] });
    },
  });
}

export function useAddDependency(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { cardId: string; dependsOnCardId: string }) => api.dependencies.add(data.cardId, data.dependsOnCardId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
      qc.invalidateQueries({ queryKey: ['card'] });
    },
  });
}

export function useRemoveDependency(boardId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { cardId: string; depId: string }) => api.dependencies.remove(data.cardId, data.depId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['board', boardId] });
      qc.invalidateQueries({ queryKey: ['card'] });
    },
  });
}
