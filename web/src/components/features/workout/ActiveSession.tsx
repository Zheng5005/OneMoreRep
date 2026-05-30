import { useState } from 'react';
import { useStore } from '../../../stores';
import { Button } from '../../ui/Button';
import { Card } from '../../ui/Card';
import { SetRow } from './SetRow';
import { SetLogger } from './SetLogger';
import { RestTimer } from './RestTimer';
import type { SetResponse, Exercise } from '../../../types';
import './ActiveSession.css';

interface ActiveSessionProps {
  onEnd: () => void;
}

interface ExerciseWithSets {
  exercise: Exercise;
  sets: SetResponse[];
}

export function ActiveSession({ onEnd }: ActiveSessionProps) {
  const { currentSession, endSession, deleteSet, fetchSessionSummary } = useStore();
  const [setLoggerOpen, setSetLoggerOpen] = useState(false);
  const [editingSet, setEditingSet] = useState<SetResponse | null>(null);
  const [showRestTimer, setShowRestTimer] = useState(false);

  const exercisesInSession: ExerciseWithSets[] = [];
  if (currentSession?.sets) {
    const exerciseMap = new Map<string, ExerciseWithSets>();
    for (const set of currentSession.sets) {
      if (!exerciseMap.has(set.exercise_id)) {
        exerciseMap.set(set.exercise_id, {
          exercise: { id: set.exercise_id, name: set.exercise_name || 'Unknown', target_muscle: '', notes: null, created_at: '' } as Exercise,
          sets: [],
        });
      }
      exerciseMap.get(set.exercise_id)!.sets.push(set);
    }
    for (const value of exerciseMap.values()) {
      exercisesInSession.push(value);
    }
  }

  const handleEditSet = (set: SetResponse) => {
    setEditingSet(set);
    setSetLoggerOpen(true);
  };

  const handleDeleteSet = async (setId: string) => {
    if (!currentSession) return;
    await deleteSet(currentSession.id, setId);
  };

  const handleSetLogged = () => {
    setShowRestTimer(true);
  };

  const handleAddExercise = () => {
    setEditingSet(null);
    setSetLoggerOpen(true);
  };

  const handleEndWorkout = async () => {
    if (!currentSession) return;
    if (!window.confirm('End this workout?')) return;

    try {
      await endSession(currentSession.id);
      await fetchSessionSummary(currentSession.id);
      onEnd();
    } catch (err) {
      console.error('Failed to end session:', err);
    }
  };

  const formatDuration = (startedAt: string) => {
    const start = new Date(startedAt);
    const now = new Date();
    const diffMs = now.getTime() - start.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const hours = Math.floor(diffMins / 60);
    const mins = diffMins % 60;
    if (hours > 0) {
      return `${hours}h ${mins}m`;
    }
    return `${mins}m`;
  };

  if (!currentSession) {
    return (
      <div className="active-session-empty">
        <p>No active session</p>
      </div>
    );
  }

  return (
    <div className="active-session">
      <div className="active-session-header">
        <div className="active-session-info">
          <h2>Workout in Progress</h2>
          <p className="active-session-duration">Duration: {formatDuration(currentSession.started_at)}</p>
        </div>
        <div className="active-session-actions">
          <Button variant="secondary" onClick={() => setShowRestTimer(!showRestTimer)}>
            {showRestTimer ? 'Hide' : 'Show'} Timer
          </Button>
          <Button variant="danger" onClick={handleEndWorkout}>
            End Workout
          </Button>
        </div>
      </div>

      {showRestTimer && (
        <div className="active-session-timer">
          <RestTimer defaultSeconds={60} onDismiss={() => setShowRestTimer(false)} />
        </div>
      )}

      <div className="active-session-exercises">
        {exercisesInSession.length === 0 ? (
          <Card>
            <div className="active-session-empty-exercises">
              <p>No exercises logged yet</p>
              <Button variant="primary" onClick={handleAddExercise}>
                Log First Set
              </Button>
            </div>
          </Card>
        ) : (
          exercisesInSession.map(({ exercise, sets }) => (
            <Card key={exercise.id} title={exercise.name}>
              <div className="active-session-sets">
                {sets.map((set) => (
                  <SetRow key={set.id} set={set} onEdit={handleEditSet} onDelete={handleDeleteSet} />
                ))}
              </div>
              <div className="active-session-exercise-actions">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setEditingSet(null);
                    setSetLoggerOpen(true);
                  }}
                >
                  Log Set
                </Button>
              </div>
            </Card>
          ))
        )}

        <Button variant="secondary" onClick={handleAddExercise}>
          + Add Set
        </Button>
      </div>

      <SetLogger
        isOpen={setLoggerOpen}
        onClose={() => {
          setSetLoggerOpen(false);
          setEditingSet(null);
        }}
        sessionId={currentSession.id}
        editingSet={editingSet}
        onSetLogged={handleSetLogged}
      />
    </div>
  );
}
