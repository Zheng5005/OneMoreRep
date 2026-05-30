import { useState, useEffect } from 'react';
import { useStore } from '../../../stores';
import type { Routine, RoutineExercise, Exercise } from '../../../types';
import { Button } from '../../ui/Button';
import { Input } from '../../ui/Input';
import { Modal } from '../../ui/Modal';
import { RoutineExerciseRow } from './RoutineExerciseRow';
import './RoutineBuilder.css';

interface RoutineBuilderProps {
  routine?: Routine;
  onSave: () => void;
  onCancel: () => void;
}

export function RoutineBuilder({ routine, onSave, onCancel }: RoutineBuilderProps) {
  const {
    createRoutine,
    updateRoutine,
    addExerciseToRoutine,
    removeExerciseFromRoutine,
    fetchRoutines,
    exercises,
    fetchExercises,
  } = useStore();

  const [name, setName] = useState(routine?.name ?? '');
  const [localExercises, setLocalExercises] = useState<RoutineExercise[]>(
    routine?.exercises ?? []
  );
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (routine) {
      fetchRoutines();
    }
  }, [routine, fetchRoutines]);

  useEffect(() => {
    fetchExercises();
  }, [fetchExercises]);

  const filteredExercises = exercises.filter((ex) =>
    ex.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleAddExercise = async (exercise: Exercise) => {
    if (!routine) return;

    const isDuplicate = localExercises.some((e) => e.exercise_id === exercise.id);
    if (isDuplicate) {
      setError('This exercise is already in the routine');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const updatedRoutine = await addExerciseToRoutine(routine.id, exercise.id);
      setLocalExercises(updatedRoutine.exercises ?? []);
      setIsAddModalOpen(false);
      setSearchQuery('');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to add exercise');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveExercise = async (reId: string) => {
    if (!routine) return;

    setLoading(true);
    setError(null);
    try {
      await removeExerciseFromRoutine(routine.id, reId);
      setLocalExercises((prev) => prev.filter((e) => e.id !== reId));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to remove exercise');
    } finally {
      setLoading(false);
    }
  };

  const handleMoveUp = async (index: number) => {
    if (!routine || index === 0) return;

    const newExercises = [...localExercises];
    const temp = newExercises[index].order;
    newExercises[index].order = newExercises[index - 1].order;
    newExercises[index - 1].order = temp;

    const [moved] = newExercises.splice(index, 1);
    newExercises.splice(index - 1, 0, moved);

    setLocalExercises(newExercises);

    const currentExercise = localExercises[index];
    const aboveExercise = localExercises[index - 1];

    try {
      await useStore.getState().reorderExercises?.(routine.id, [
        ...localExercises.filter((_, i) => i !== index && i !== index - 1).map((e) => e.id),
        currentExercise.id,
        aboveExercise.id,
      ]);
    } catch {
      setLocalExercises(localExercises);
    }
  };

  const handleMoveDown = async (index: number) => {
    if (!routine || index === localExercises.length - 1) return;

    const newExercises = [...localExercises];
    const temp = newExercises[index].order;
    newExercises[index].order = newExercises[index + 1].order;
    newExercises[index + 1].order = temp;

    const [moved] = newExercises.splice(index, 1);
    newExercises.splice(index + 1, 0, moved);

    setLocalExercises(newExercises);

    const currentExercise = localExercises[index];
    const belowExercise = localExercises[index + 1];

    try {
      await useStore.getState().reorderExercises?.(routine.id, [
        ...localExercises.filter((_, i) => i !== index && i !== index + 1).map((e) => e.id),
        belowExercise.id,
        currentExercise.id,
      ]);
    } catch {
      setLocalExercises(localExercises);
    }
  };

  const handleSave = async () => {
    if (!name.trim()) {
      setError('Routine name is required');
      return;
    }

    setSaving(true);
    setError(null);
    try {
      if (routine) {
        await updateRoutine(routine.id, { name });
      } else {
        await createRoutine({ name });
      }
      onSave();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save routine');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="routine-builder">
      <div className="routine-builder-header">
        <h2 className="routine-builder-title">
          {routine ? 'Edit Routine' : 'Create Routine'}
        </h2>
      </div>

      <div className="routine-builder-form">
        <Input
          label="Routine Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g., Push Day, Upper Body"
          required
        />

        {error && <p className="routine-builder-error">{error}</p>}
      </div>

      <div className="routine-builder-exercises">
        <div className="routine-builder-exercises-header">
          <h3>Exercises</h3>
          {routine && (
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setIsAddModalOpen(true)}
            >
              + Add Exercise
            </Button>
          )}
        </div>

        {localExercises.length === 0 ? (
          <div className="routine-builder-empty">
            <p>No exercises in this routine yet.</p>
            {routine && (
              <Button
                variant="primary"
                onClick={() => setIsAddModalOpen(true)}
              >
                + Add Exercise
              </Button>
            )}
          </div>
        ) : (
          <div className="routine-exercise-list">
            {localExercises.map((exercise, index) => (
              <RoutineExerciseRow
                key={exercise.id}
                exercise={exercise}
                order={index + 1}
                totalExercises={localExercises.length}
                onMoveUp={() => handleMoveUp(index)}
                onMoveDown={() => handleMoveDown(index)}
                onRemove={() => handleRemoveExercise(exercise.id)}
              />
            ))}
          </div>
        )}
      </div>

      <div className="routine-builder-actions">
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button
          variant="primary"
          onClick={handleSave}
          loading={saving}
          disabled={loading}
        >
          {routine ? 'Save Changes' : 'Create Routine'}
        </Button>
      </div>

      <Modal
        isOpen={isAddModalOpen}
        onClose={() => {
          setIsAddModalOpen(false);
          setSearchQuery('');
          setError(null);
        }}
        title="Add Exercise"
      >
        <div className="add-exercise-modal">
          <Input
            placeholder="Search exercises..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            autoFocus
          />
          {error && <p className="add-exercise-error">{error}</p>}
          <div className="add-exercise-list">
            {filteredExercises.length === 0 ? (
              <p className="add-exercise-empty">No exercises found</p>
            ) : (
              filteredExercises.map((ex) => (
                <button
                  key={ex.id}
                  className="add-exercise-item"
                  onClick={() => handleAddExercise(ex)}
                  disabled={loading}
                >
                  <span className="add-exercise-name">{ex.name}</span>
                  <span className="add-exercise-muscle">{ex.target_muscle}</span>
                </button>
              ))
            )}
          </div>
        </div>
      </Modal>
    </div>
  );
}