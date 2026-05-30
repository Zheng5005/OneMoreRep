import type { Routine } from '../../../types';
import { Button } from '../../ui/Button';
import './RoutineCard.css';

interface RoutineCardProps {
  routine: Routine;
  onEdit: (routine: Routine) => void;
  onDelete: (routine: Routine) => void;
}

export function RoutineCard({ routine, onEdit, onDelete }: RoutineCardProps) {
  const exerciseCount = routine.exercises?.length ?? 0;

  return (
    <div className="routine-card">
      <div className="routine-card-main">
        <h4 className="routine-name">{routine.name}</h4>
        <span className="routine-exercises">{exerciseCount} exercise{exerciseCount !== 1 ? 's' : ''}</span>
        <span className="routine-date">{new Date(routine.created_at).toLocaleDateString()}</span>
      </div>
      <div className="routine-card-actions">
        <Button variant="ghost" size="sm" onClick={() => onEdit(routine)}>
          Edit
        </Button>
        <Button variant="danger" size="sm" onClick={() => onDelete(routine)}>
          Delete
        </Button>
      </div>
    </div>
  );
}