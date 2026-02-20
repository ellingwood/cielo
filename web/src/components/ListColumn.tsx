import { useState } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import type { List } from '../api/types';
import CardTile from './CardTile';

interface Props {
  list: List;
  onCardClick: (cardId: string) => void;
  onCreateCard: (listId: string, title: string) => void;
  onDeleteList: (listId: string) => void;
}

export default function ListColumn({ list, onCardClick, onCreateCard, onDeleteList }: Props) {
  const [showAdd, setShowAdd] = useState(false);
  const [title, setTitle] = useState('');

  const { setNodeRef, isOver } = useDroppable({ id: list.id, data: { type: 'list', list } });

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    onCreateCard(list.id, title.trim());
    setTitle('');
    setShowAdd(false);
  };

  const cards = list.cards || [];

  return (
    <div className={`w-72 shrink-0 flex flex-col max-h-full bg-gray-50 rounded-xl border ${isOver ? 'border-sky-400 bg-sky-50' : 'border-gray-200'}`}>
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-gray-200">
        <h3 className="font-semibold text-sm text-gray-700">{list.name}</h3>
        <div className="flex items-center gap-1">
          <span className="text-xs text-gray-400">{cards.length}</span>
          <button
            onClick={() => onDeleteList(list.id)}
            className="text-gray-300 hover:text-red-500 text-sm ml-1"
            title="Delete list"
          >
            &times;
          </button>
        </div>
      </div>
      <div ref={setNodeRef} className="flex-1 overflow-y-auto p-2 space-y-2 min-h-[60px]">
        <SortableContext items={cards.map(c => c.id)} strategy={verticalListSortingStrategy}>
          {cards.map(card => (
            <CardTile key={card.id} card={card} onClick={() => onCardClick(card.id)} />
          ))}
        </SortableContext>
      </div>
      <div className="p-2 border-t border-gray-200">
        {showAdd ? (
          <form onSubmit={handleAdd}>
            <input
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="Card title"
              className="w-full px-2 py-1.5 text-sm border border-gray-200 rounded focus:outline-none focus:ring-2 focus:ring-sky-400"
              autoFocus
            />
            <div className="flex gap-1 mt-1">
              <button type="submit" className="px-2 py-1 bg-sky-600 text-white rounded text-xs hover:bg-sky-700">Add</button>
              <button type="button" onClick={() => setShowAdd(false)} className="px-2 py-1 text-gray-400 text-xs">Cancel</button>
            </div>
          </form>
        ) : (
          <button
            onClick={() => setShowAdd(true)}
            className="w-full text-sm text-gray-400 hover:text-sky-600 py-1 transition"
          >
            + Add card
          </button>
        )}
      </div>
    </div>
  );
}
