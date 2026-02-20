import { useState, useEffect } from 'react';
import { useCard, useUpdateCard, useDeleteCard, useLabels, useAddLabelToCard, useRemoveLabelFromCard, useAddDependency, useRemoveDependency } from '../api/hooks';
import type { Card, Label } from '../api/types';

interface Props {
  cardId: string;
  boardId: string;
  onClose: () => void;
}

const statusOptions = ['unassigned', 'assigned', 'in_progress', 'blocked', 'done'];
const priorityOptions = ['low', 'medium', 'high', 'critical'];

export default function CardDetail({ cardId, boardId, onClose }: Props) {
  const { data: card, isLoading } = useCard(cardId);
  const { data: boardLabels } = useLabels(boardId);
  const updateCard = useUpdateCard(boardId);
  const deleteCard = useDeleteCard(boardId);
  const addLabel = useAddLabelToCard(boardId);
  const removeLabel = useRemoveLabelFromCard(boardId);
  const addDep = useAddDependency(boardId);
  const removeDep = useRemoveDependency(boardId);

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [assignee, setAssignee] = useState('');
  const [depInput, setDepInput] = useState('');

  useEffect(() => {
    if (card) {
      setTitle(card.title);
      setDescription(card.description);
      setAssignee(card.assignee);
    }
  }, [card]);

  if (isLoading || !card) return null;

  const handleSave = (updates: Record<string, unknown>) => {
    updateCard.mutate({ id: cardId, updates });
  };

  const handleDelete = () => {
    deleteCard.mutate(cardId);
    onClose();
  };

  const availableLabels = (boardLabels || []).filter(
    (l: Label) => !card.labels?.some((cl: Label) => cl.id === l.id)
  );

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-start justify-center pt-16 px-4" onClick={onClose}>
      <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex items-start justify-between mb-4">
            <input
              value={title}
              onChange={e => setTitle(e.target.value)}
              onBlur={() => title !== card.title && handleSave({ title })}
              className="text-lg font-bold text-gray-800 w-full border-b border-transparent hover:border-gray-200 focus:border-sky-400 focus:outline-none pb-1"
            />
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl ml-4">&times;</button>
          </div>

          {/* Status & Priority */}
          <div className="grid grid-cols-2 gap-3 mb-4">
            <div>
              <label className="text-xs text-gray-500 font-medium">Status</label>
              <select
                value={card.status}
                onChange={e => handleSave({ status: e.target.value })}
                className="w-full mt-1 text-sm border border-gray-200 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-sky-400"
              >
                {statusOptions.map(s => <option key={s} value={s}>{s.replace('_', ' ')}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs text-gray-500 font-medium">Priority</label>
              <select
                value={card.priority}
                onChange={e => handleSave({ priority: e.target.value })}
                className="w-full mt-1 text-sm border border-gray-200 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-sky-400"
              >
                {priorityOptions.map(p => <option key={p} value={p}>{p}</option>)}
              </select>
            </div>
          </div>

          {/* Assignee */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 font-medium">Assignee</label>
            <input
              value={assignee}
              onChange={e => setAssignee(e.target.value)}
              onBlur={() => assignee !== card.assignee && handleSave({ assignee })}
              placeholder="Agent or user name"
              className="w-full mt-1 text-sm border border-gray-200 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-sky-400"
            />
          </div>

          {/* Description */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 font-medium">Description</label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              onBlur={() => description !== card.description && handleSave({ description })}
              placeholder="Add a description..."
              rows={4}
              className="w-full mt-1 text-sm border border-gray-200 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-sky-400 resize-none"
            />
          </div>

          {/* Labels */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 font-medium">Labels</label>
            <div className="flex flex-wrap gap-1.5 mt-1">
              {card.labels?.map((l: Label) => (
                <span
                  key={l.id}
                  className="text-xs px-2 py-0.5 rounded-full text-white cursor-pointer hover:opacity-80"
                  style={{ backgroundColor: l.color }}
                  onClick={() => removeLabel.mutate({ cardId, labelId: l.id })}
                  title="Click to remove"
                >
                  {l.name} &times;
                </span>
              ))}
              {availableLabels.map((l: Label) => (
                <span
                  key={l.id}
                  className="text-xs px-2 py-0.5 rounded-full border cursor-pointer hover:opacity-80 text-gray-500"
                  style={{ borderColor: l.color }}
                  onClick={() => addLabel.mutate({ cardId, labelId: l.id })}
                  title="Click to add"
                >
                  + {l.name}
                </span>
              ))}
            </div>
          </div>

          {/* Dependencies */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 font-medium">Blocked By</label>
            <div className="space-y-1 mt-1">
              {card.dependencies?.map((dep: Card) => (
                <div key={dep.id} className="flex items-center gap-2 text-sm">
                  <span className="text-gray-700">{dep.title}</span>
                  <button
                    onClick={() => removeDep.mutate({ cardId, depId: dep.id })}
                    className="text-red-400 hover:text-red-600 text-xs"
                  >
                    remove
                  </button>
                </div>
              ))}
              <div className="flex gap-1 mt-1">
                <input
                  value={depInput}
                  onChange={e => setDepInput(e.target.value)}
                  placeholder="Card ID to depend on"
                  className="flex-1 text-xs border border-gray-200 rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sky-400"
                />
                <button
                  onClick={() => { if (depInput) { addDep.mutate({ cardId, dependsOnCardId: depInput }); setDepInput(''); } }}
                  className="text-xs px-2 py-1 bg-sky-100 text-sky-700 rounded hover:bg-sky-200"
                >
                  Add
                </button>
              </div>
            </div>
          </div>

          {/* Dependents */}
          {card.dependents && card.dependents.length > 0 && (
            <div className="mb-4">
              <label className="text-xs text-gray-500 font-medium">Blocking</label>
              <div className="space-y-1 mt-1">
                {card.dependents.map((dep: Card) => (
                  <div key={dep.id} className="text-sm text-gray-600">{dep.title}</div>
                ))}
              </div>
            </div>
          )}

          {/* Activity */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 font-medium">Activity</label>
            <div className="mt-1 space-y-2 max-h-48 overflow-y-auto">
              {card.activity?.length === 0 && <p className="text-xs text-gray-400">No activity yet</p>}
              {card.activity?.map(a => (
                <div key={a.id} className="text-xs text-gray-600 flex gap-2">
                  <span className="font-medium text-gray-700">{a.actor}</span>
                  <span>{a.action.replace('_', ' ')}</span>
                  <span className="text-gray-400 ml-auto">{new Date(a.created_at).toLocaleString()}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end border-t border-gray-100 pt-3">
            <button
              onClick={handleDelete}
              className="text-sm text-red-500 hover:text-red-700 transition"
            >
              Delete Card
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
