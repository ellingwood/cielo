import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useBoards, useCreateBoard, useDeleteBoard } from '../api/hooks';

export default function BoardList() {
  const { data: boards, isLoading } = useBoards();
  const createBoard = useCreateBoard();
  const deleteBoard = useDeleteBoard();
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    createBoard.mutate({ name: name.trim(), description: desc.trim() }, {
      onSuccess: () => { setName(''); setDesc(''); setShowForm(false); },
    });
  };

  if (isLoading) return <div className="p-8 text-sky-600">Loading boards...</div>;

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Boards</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-700 transition text-sm font-medium"
        >
          + New Board
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="mb-6 bg-white rounded-lg p-4 shadow-sm border border-sky-100">
          <input
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Board name"
            className="w-full mb-2 px-3 py-2 border border-gray-200 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-400"
            autoFocus
          />
          <input
            value={desc}
            onChange={e => setDesc(e.target.value)}
            placeholder="Description (optional)"
            className="w-full mb-3 px-3 py-2 border border-gray-200 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-400"
          />
          <div className="flex gap-2">
            <button type="submit" className="px-4 py-1.5 bg-sky-600 text-white rounded-md text-sm hover:bg-sky-700">Create</button>
            <button type="button" onClick={() => setShowForm(false)} className="px-4 py-1.5 text-gray-500 text-sm hover:text-gray-700">Cancel</button>
          </div>
        </form>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {boards?.map(board => (
          <div key={board.id} className="bg-white rounded-lg shadow-sm border border-sky-100 hover:shadow-md transition group relative">
            <Link to={`/boards/${board.id}`} className="block p-5">
              <h2 className="font-semibold text-gray-800 mb-1">{board.name}</h2>
              {board.description && <p className="text-sm text-gray-500">{board.description}</p>}
            </Link>
            <button
              onClick={() => deleteBoard.mutate(board.id)}
              className="absolute top-3 right-3 text-gray-300 hover:text-red-500 transition opacity-0 group-hover:opacity-100 text-sm"
              title="Delete board"
            >
              &times;
            </button>
          </div>
        ))}
        {boards?.length === 0 && (
          <p className="text-gray-400 col-span-3 text-center py-12">No boards yet. Create one to get started.</p>
        )}
      </div>
    </div>
  );
}
