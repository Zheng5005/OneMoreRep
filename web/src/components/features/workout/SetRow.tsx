import { Button } from '../../ui/Button';
import type { SetResponse } from '../../../types';
import './SetRow.css';

interface SetRowProps {
  set: SetResponse;
  onEdit: (set: SetResponse) => void;
  onDelete: (setId: string) => void;
}

export function SetRow({ set, onEdit, onDelete }: SetRowProps) {
  const handleDelete = () => {
    if (window.confirm('Delete this set?')) {
      onDelete(set.id);
    }
  };

  return (
    <div className="set-row">
      <div className="set-row-number">Set {set.set_number}</div>
      <div className="set-row-values">
        <span className="set-row-weight">{set.weight} lbs</span>
        <span className="set-row-separator">×</span>
        <span className="set-row-reps">{set.reps} reps</span>
      </div>
      <div className="set-row-actions">
        <Button variant="ghost" size="sm" onClick={() => onEdit(set)}>
          Edit
        </Button>
        <Button variant="danger" size="sm" onClick={handleDelete}>
          Delete
        </Button>
      </div>
    </div>
  );
}
