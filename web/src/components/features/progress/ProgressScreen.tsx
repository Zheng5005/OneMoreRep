import { useState, useEffect } from 'react';
import { useStore } from '../../../stores';
import { ExerciseHistory } from './ExerciseHistory';
import { VolumeChart } from './VolumeChart';
import { SessionSummary } from './SessionSummary';
import { Card } from '../../ui/Card';
import { EmptyState } from '../../shared/EmptyState';
import './ProgressScreen.css';

type Tab = 'history' | 'volume' | 'summary';

export function ProgressScreen() {
  const [activeTab, setActiveTab] = useState<Tab>('history');
  const { exercises, fetchExercises, selectedExercise, selectExercise } = useStore();
  const { sessionSummary, fetchSessionSummary } = useStore();

  useEffect(() => {
    fetchExercises();
  }, [fetchExercises]);

  useEffect(() => {
    if (sessionSummary) {
      fetchSessionSummary(sessionSummary.session_id);
    }
  }, [sessionSummary, fetchSessionSummary]);

  const handleExerciseSelect = (exerciseId: string) => {
    const exercise = exercises.find((e) => e.id === exerciseId);
    if (exercise) {
      selectExercise(exercise);
    }
  };

  return (
    <div className="progress-screen">
      <div className="progress-header">
        <h2>Progress</h2>
      </div>

      <div className="progress-tabs">
        <button
          className={`progress-tab ${activeTab === 'history' ? 'progress-tab-active' : ''}`}
          onClick={() => setActiveTab('history')}
        >
          Exercise History
        </button>
        <button
          className={`progress-tab ${activeTab === 'volume' ? 'progress-tab-active' : ''}`}
          onClick={() => setActiveTab('volume')}
        >
          Volume Chart
        </button>
        <button
          className={`progress-tab ${activeTab === 'summary' ? 'progress-tab-active' : ''}`}
          onClick={() => setActiveTab('summary')}
        >
          Session Summary
        </button>
      </div>

      <div className="progress-content">
        {activeTab === 'history' && (
          <div className="progress-history-tab">
            <Card>
              <div className="exercise-selector">
                <label htmlFor="exercise-select">Select Exercise</label>
                <select
                  id="exercise-select"
                  className="exercise-select"
                  value={selectedExercise?.id || ''}
                  onChange={(e) => handleExerciseSelect(e.target.value)}
                >
                  <option value="">Choose an exercise...</option>
                  {exercises.map((ex) => (
                    <option key={ex.id} value={ex.id}>
                      {ex.name}
                    </option>
                  ))}
                </select>
              </div>
            </Card>

            {selectedExercise ? (
              <ExerciseHistory exerciseId={selectedExercise.id} />
            ) : (
              <EmptyState
                title="Select an exercise"
                description="Choose an exercise above to view its history."
                icon="🏋️"
              />
            )}
          </div>
        )}

        {activeTab === 'volume' && (
          <div className="progress-volume-tab">
            <Card>
              <div className="volume-chart-wrapper">
                <VolumeChart exerciseId={selectedExercise?.id} />
              </div>
            </Card>
          </div>
        )}

        {activeTab === 'summary' && (
          <div className="progress-summary-tab">
            {sessionSummary ? (
              <SessionSummary sessionId={sessionSummary.session_id} />
            ) : (
              <EmptyState
                title="No recent session"
                description="Complete a workout to see your session summary here."
                icon="🏆"
              />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
