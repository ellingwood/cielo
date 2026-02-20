import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { DndContext, type DragEndEvent, PointerSensor, useSensor, useSensors, closestCorners, DragOverlay, type DragStartEvent } from '@dnd-kit/core';
import { useBoard, useCreateList, useCreateCard, useDeleteList, useMoveCard, useCreateLabel } from '../api/hooks';
import { useSSE } from '../components/SSEProvider';
import ListColumn from '../components/ListColumn';
import CardDetail from '../components/CardDetail';
import type { Card } from '../api/types';

export default function BoardView() {
  const { boardId } = useParams<{ boardId: string }>();
  const { data: board, isLoading } = useBoard(boardId!);
  const createList = useCreateList(boardId!);
  const createCard = useCreateCard(boardId!);
  const deleteList = useDeleteList(boardId!);
  const moveCard = useMoveCard(boardId!);
  const createLabel = useCreateLabel(boardId!);

  useSSE(boardId!);

  const [selectedCardId, setSelectedCardId] = useState<string | null>(null);
  const [showAddList, setShowAddList] = useState(false);
  const [listName, setListName] = useState('');
  const [showLabelForm, setShowLabelForm] = useState(false);
  const [labelName, setLabelName] = useState('');
  const [labelColor, setLabelColor] = useState('#3b82f6');
  const [activeCard, setActiveCard] = useState<Card | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  );

  if (isLoading || !board) return <div className="p-8 text-sky-600">Loading board...</div>;

  const handleAddList = (e: React.FormEvent) => {
    e.preventDefault();
    if (!listName.trim()) return;
    createList.mutate({ name: listName.trim(), position: (board.lists?.length || 0) });
    setListName('');
    setShowAddList(false);
  };

  const handleCreateCard = (listId: string, title: string) => {
    const list = board.lists?.find(l => l.id === listId);
    createCard.mutate({ listId, title, position: list?.cards?.length || 0 });
  };

  const handleDragStart = (event: DragStartEvent) => {
    const { active } = event;
    if (active.data.current?.type === 'card') {
      setActiveCard(active.data.current.card);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveCard(null);
    const { active, over } = event;
    if (!over) return;

    const cardId = active.id as string;
    const activeData = active.data.current;
    if (activeData?.type !== 'card') return;

    let targetListId: string;
    let position: number;

    const overData = over.data.current;
    if (overData?.type === 'list') {
      targetListId = over.id as string;
      const targetList = board.lists?.find(l => l.id === targetListId);
      position = targetList?.cards?.length || 0;
    } else if (overData?.type === 'card') {
      const overCard = overData.card as Card;
      targetListId = overCard.list_id;
      const targetList = board.lists?.find(l => l.id === targetListId);
      const idx = targetList?.cards?.findIndex(c => c.id === over.id) ?? 0;
      position = idx;
    } else {
      return;
    }

    moveCard.mutate({ cardId, listId: targetListId, position });
  };

  const handleAddLabel = (e: React.FormEvent) => {
    e.preventDefault();
    if (!labelName.trim()) return;
    createLabel.mutate({ name: labelName.trim(), color: labelColor });
    setLabelName('');
    setShowLabelForm(false);
  };

  return (
    <div className="flex flex-col h-[calc(100vh-53px)]">
      <div className="flex items-center gap-4 px-6 py-3 border-b border-gray-200 bg-white/50">
        <Link to="/" className="text-sm text-sky-600 hover:text-sky-800">&larr; Boards</Link>
        <h2 className="text-lg font-bold text-gray-800">{board.name}</h2>
        {board.description && <span className="text-sm text-gray-400">{board.description}</span>}
        <button
          onClick={() => setShowLabelForm(!showLabelForm)}
          className="ml-auto text-xs text-sky-600 hover:text-sky-800"
        >
          + Label
        </button>
      </div>

      {showLabelForm && (
        <form onSubmit={handleAddLabel} className="px-6 py-2 bg-white border-b flex items-center gap-2">
          <input value={labelName} onChange={e => setLabelName(e.target.value)} placeholder="Label name" className="text-sm border rounded px-2 py-1" autoFocus />
          <input type="color" value={labelColor} onChange={e => setLabelColor(e.target.value)} className="w-8 h-8 rounded cursor-pointer" />
          <button type="submit" className="text-sm px-3 py-1 bg-sky-600 text-white rounded hover:bg-sky-700">Create</button>
          <button type="button" onClick={() => setShowLabelForm(false)} className="text-sm text-gray-400">Cancel</button>
        </form>
      )}

      <DndContext sensors={sensors} collisionDetection={closestCorners} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div className="flex-1 overflow-x-auto p-4">
          <div className="flex gap-4 h-full items-start">
            {board.lists?.map(list => (
              <ListColumn
                key={list.id}
                list={list}
                onCardClick={setSelectedCardId}
                onCreateCard={handleCreateCard}
                onDeleteList={(id) => deleteList.mutate(id)}
              />
            ))}
            <div className="w-72 shrink-0">
              {showAddList ? (
                <form onSubmit={handleAddList} className="bg-gray-50 rounded-xl border border-gray-200 p-3">
                  <input
                    value={listName}
                    onChange={e => setListName(e.target.value)}
                    placeholder="List name"
                    className="w-full text-sm border border-gray-200 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-sky-400 mb-2"
                    autoFocus
                  />
                  <div className="flex gap-1">
                    <button type="submit" className="px-3 py-1 bg-sky-600 text-white rounded text-xs hover:bg-sky-700">Add</button>
                    <button type="button" onClick={() => setShowAddList(false)} className="px-3 py-1 text-gray-400 text-xs">Cancel</button>
                  </div>
                </form>
              ) : (
                <button
                  onClick={() => setShowAddList(true)}
                  className="w-full py-3 text-sm text-gray-400 hover:text-sky-600 bg-gray-50/50 rounded-xl border border-dashed border-gray-200 hover:border-sky-300 transition"
                >
                  + Add List
                </button>
              )}
            </div>
          </div>
        </div>
        <DragOverlay>
          {activeCard && (
            <div className="bg-white rounded-lg shadow-lg border border-sky-200 p-3 w-64 opacity-90">
              <p className="text-sm font-medium text-gray-800">{activeCard.title}</p>
            </div>
          )}
        </DragOverlay>
      </DndContext>

      {selectedCardId && (
        <CardDetail cardId={selectedCardId} boardId={boardId!} onClose={() => setSelectedCardId(null)} />
      )}
    </div>
  );
}
