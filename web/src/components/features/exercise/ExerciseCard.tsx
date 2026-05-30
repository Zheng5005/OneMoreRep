import type { Exercise } from '../../../types';
import { Button } from '../../ui/Button';
import './ExerciseCard.css';

interface ExerciseCardProps {
  exercise: Exercise;
  onEdit: (exercise: Exercise) => void;
  onDelete: (exercise: Exercise) => void;
}

export function ExerciseCard({ exercise, onEdit, onDelete }: ExerciseCardProps) {
  return (
    <div className="exercise-card">
      <div className="exercise-card-main">
        <h4 className="exercise-name">{exercise.name}</h4>
        <span className="exercise-muscle">{exercise.target_muscle}</span>
        {exercise.notes && (
          <p className="exercise-notes">{exercise.notes}</p>
        )}
      </div>
      <div className="exercise-card-actions">
        <Button variant="ghost" size="sm" onClick={() => onEdit(exercise)}>
          Edit
        </Button>
        <Button variant="danger" size="sm" onClick={() => onDelete(exercise)}>
          Delete
        </Button>
      </div>
    </div>
  );
}
