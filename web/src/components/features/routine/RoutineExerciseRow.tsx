import { Button } from '../../ui/Button';
import './RoutineExerciseRow.css';

interface RoutineExerciseRowProps {
  exercise: {
    id: string;
    exercise_name?: string;
    exercise_id: string;
    target_muscle?: string;
  };
  order: number;
  totalExercises: number;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onRemove: () => void;
}

export function RoutineExerciseRow({
  exercise,
  order,
  totalExercises,
  onMoveUp,
  onMoveDown,
  onRemove,
}: RoutineExerciseRowProps) {
  return (
    <div className="routine-exercise-row">
      <span className="routine-exercise-order">{order}</span>
      <div className="routine-exercise-info">
        <span className="routine-exercise-name">{exercise.exercise_name}</span>
        {exercise.target_muscle && (
          <span className="routine-exercise-muscle">{exercise.target_muscle}</span>
        )}
      </div>
      <div className="routine-exercise-actions">
        <Button
          variant="ghost"
          size="sm"
          onClick={onMoveUp}
          disabled={order === 1}
          aria-label="Move up"
        >
          ↑
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={onMoveDown}
          disabled={order === totalExercises}
          aria-label="Move down"
        >
          ↓
        </Button>
        <Button
          variant="danger"
          size="sm"
          onClick={onRemove}
          aria-label="Remove exercise"
        >
          ✕
        </Button>
      </div>
    </div>
  );
}