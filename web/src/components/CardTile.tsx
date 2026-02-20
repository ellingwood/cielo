import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Card } from '../api/types';

const priorityColors: Record<string, string> = {
  critical: 'bg-red-500',
  high: 'bg-orange-400',
  medium: 'bg-yellow-400',
  low: 'bg-green-400',
};

const statusIcons: Record<string, string> = {
  unassigned: '',
  assigned: '👤',
  in_progress: '⚡',
  blocked: '🚫',
  done: '✅',
};

interface Props {
  card: Card;
  onClick: () => void;
}

export default function CardTile({ card, onClick }: Props) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: card.id,
    data: { type: 'card', card },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className="bg-white rounded-lg shadow-sm border border-gray-100 p-3 cursor-pointer hover:shadow-md transition group"
    >
      <div className="flex items-start gap-2">
        <div className={`w-1.5 h-1.5 rounded-full mt-2 shrink-0 ${priorityColors[card.priority] || 'bg-gray-300'}`} />
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-gray-800 leading-snug">{card.title}</p>
          {card.labels && card.labels.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-1.5">
              {card.labels.map(l => (
                <span key={l.id} className="text-[10px] px-1.5 py-0.5 rounded-full text-white font-medium" style={{ backgroundColor: l.color }}>
                  {l.name}
                </span>
              ))}
            </div>
          )}
          <div className="flex items-center gap-2 mt-1.5">
            {card.assignee && (
              <span className="text-[10px] text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded">{card.assignee}</span>
            )}
            {statusIcons[card.status] && <span className="text-xs">{statusIcons[card.status]}</span>}
          </div>
        </div>
      </div>
    </div>
  );
}
