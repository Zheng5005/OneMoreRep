import { useState, useEffect, useMemo } from 'react';
import { Modal } from '../../ui/Modal';
import { Button } from '../../ui/Button';
import { Input } from '../../ui/Input';
import { useStore } from '../../../stores';
import { useAutoFill } from '../../../hooks/useAutoFill';
import type { SetResponse, Exercise } from '../../../types';
import './SetLogger.css';

interface SetLoggerProps {
  isOpen: boolean;
  onClose: () => void;
  sessionId: string;
  editingSet?: SetResponse | null;
  onSetLogged?: () => void;
}

export function SetLogger({ isOpen, onClose, sessionId, editingSet, onSetLogged }: SetLoggerProps) {
  const { exercises, fetchExercises } = useStore();
  const [selectedExerciseId, setSelectedExerciseId] = useState<string>('');
  const [weight, setWeight] = useState<string>('');
  const [reps, setReps] = useState<string>('');
  const [errors, setErrors] = useState<{ weight?: string; reps?: string; exercise?: string }>({});
  const [loading, setLoading] = useState(false);

  const { lastValues, loading: autofillLoading } = useAutoFill(editingSet?.exercise_id || selectedExerciseId || null);

  useEffect(() => {
    if (isOpen && exercises.length === 0) {
      fetchExercises();
    }
  }, [isOpen, exercises.length, fetchExercises]);

  const computedWeight = useMemo(() => {
    if (editingSet) {
      return editingSet.weight.toString();
    }
    if (lastValues && selectedExerciseId && lastValues.weight !== null) {
      return lastValues.weight.toString();
    }
    return weight;
  }, [editingSet, lastValues, selectedExerciseId, weight]);

  const computedReps = useMemo(() => {
    if (editingSet) {
      return editingSet.reps.toString();
    }
    if (lastValues && selectedExerciseId && lastValues.reps !== null) {
      return lastValues.reps.toString();
    }
    return reps;
  }, [editingSet, lastValues, selectedExerciseId, reps]);

  const validate = (): boolean => {
    const newErrors: typeof errors = {};

    if (!editingSet && !selectedExerciseId) {
      newErrors.exercise = 'Please select an exercise';
    }

    const weightNum = parseFloat(computedWeight);
    if (isNaN(weightNum) || weightNum < 0) {
      newErrors.weight = 'Weight must be 0 or greater';
    }

    const repsNum = parseInt(computedReps, 10);
    if (isNaN(repsNum) || repsNum <= 0) {
      newErrors.reps = 'Reps must be greater than 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setLoading(true);
    try {
      const { addSet, updateSet } = useStore.getState();

      if (editingSet) {
        await updateSet(sessionId, editingSet.id, {
          weight: parseFloat(computedWeight),
          reps: parseInt(computedReps, 10),
        });
      } else {
        await addSet(sessionId, {
          exercise_id: selectedExerciseId,
          weight: parseFloat(computedWeight),
          reps: parseInt(computedReps, 10),
        });
      }

      onClose();
      onSetLogged?.();
    } catch (err) {
      setErrors({ weight: err instanceof Error ? err.message : 'Failed to save set' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={editingSet ? 'Edit Set' : 'Log Set'}>
      <form onSubmit={handleSubmit} className="set-logger-form">
        {!editingSet && (
          <div className="set-logger-field">
            <label className="input-label">Exercise</label>
            <select
              className="input-field"
              value={selectedExerciseId}
              onChange={(e) => setSelectedExerciseId(e.target.value)}
            >
              <option value="">Select an exercise</option>
              {exercises.map((ex: Exercise) => (
                <option key={ex.id} value={ex.id}>
                  {ex.name}
                </option>
              ))}
            </select>
            {errors.exercise && <span className="input-error-msg">{errors.exercise}</span>}
          </div>
        )}

        <div className="set-logger-row">
          <div className="set-logger-field">
            <Input
              label="Weight (lbs)"
              type="number"
              step="0.5"
              min="0"
              value={computedWeight}
              onChange={(e) => setWeight(e.target.value)}
              error={errors.weight}
              placeholder={autofillLoading ? 'Loading...' : '0'}
            />
          </div>
          <div className="set-logger-field">
            <Input
              label="Reps"
              type="number"
              min="1"
              value={computedReps}
              onChange={(e) => setReps(e.target.value)}
              error={errors.reps}
              placeholder="0"
            />
          </div>
        </div>

        <div className="set-logger-actions">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" loading={loading}>
            {editingSet ? 'Update' : 'Log Set'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
