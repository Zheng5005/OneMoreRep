import { useState, useEffect } from 'react';
import { Modal } from '../../ui/Modal';
import { Button } from '../../ui/Button';
import { useStore } from '../../../stores';
import type { Routine } from '../../../types';
import './StartWorkoutModal.css';

interface StartWorkoutModalProps {
  isOpen: boolean;
  onClose: () => void;
  onStart: () => void;
}

function StartWorkoutModalContent({ onClose, onStart, routines, fetchRoutines, startSession }: {
  onClose: () => void;
  onStart: () => void;
  routines: Routine[];
  fetchRoutines: () => void;
  startSession: (routineId?: string) => Promise<unknown>;
}) {
  const [selectedRoutineId, setSelectedRoutineId] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (routines.length === 0) {
      fetchRoutines();
    }
  }, [routines.length, fetchRoutines]);

  const handleStart = async () => {
    setLoading(true);
    setError(null);
    try {
      await startSession(selectedRoutineId || undefined);
      onStart();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start workout');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="start-workout-content">
      <p className="start-workout-description">Choose how you want to start your workout:</p>

      <div className="start-workout-options">
        <button
          className={`start-workout-option ${!selectedRoutineId ? 'selected' : ''}`}
          onClick={() => setSelectedRoutineId('')}
        >
          <span className="option-icon">🏋️</span>
          <span className="option-title">Free Workout</span>
          <span className="option-desc">Add exercises as you go</span>
        </button>

        <button
          className={`start-workout-option ${selectedRoutineId ? 'selected' : ''}`}
          onClick={() => {}}
        >
          <span className="option-icon">📋</span>
          <span className="option-title">From Routine</span>
          <span className="option-desc">Start with a saved routine</span>
        </button>
      </div>

      {selectedRoutineId !== '' && (
        <div className="start-workout-routine-select">
          <label className="input-label">Select Routine</label>
          <select
            className="input-field"
            value={selectedRoutineId}
            onChange={(e) => setSelectedRoutineId(e.target.value)}
          >
            <option value="">Choose a routine...</option>
            {routines.map((routine: Routine) => (
              <option key={routine.id} value={routine.id}>
                {routine.name}
              </option>
            ))}
          </select>
        </div>
      )}

      {error && <p className="start-workout-error">{error}</p>}

      <div className="start-workout-actions">
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="primary" onClick={handleStart} loading={loading}>
          Start Workout
        </Button>
      </div>
    </div>
  );
}

export function StartWorkoutModal({ isOpen, onClose, onStart }: StartWorkoutModalProps) {
  const { routines, fetchRoutines, startSession } = useStore();

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Start Workout">
      <StartWorkoutModalContent
        onClose={onClose}
        onStart={onStart}
        routines={routines}
        fetchRoutines={fetchRoutines}
        startSession={startSession as (routineId?: string) => Promise<unknown>}
      />
    </Modal>
  );
}
