import { useEffect } from 'react';
import { Card } from '../../ui/Card';
import { Button } from '../../ui/Button';
import { useStore } from '../../../stores';
import './SessionSummary.css';

interface SessionSummaryProps {
  sessionId?: string;
}

export function SessionSummary({ sessionId }: SessionSummaryProps) {
  const { sessionSummary, loading, error, fetchSessionSummary } = useStore();

  useEffect(() => {
    if (sessionId) {
      fetchSessionSummary(sessionId);
    }
  }, [sessionId, fetchSessionSummary]);

  const formatDuration = (seconds: number) => {
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    if (hrs > 0) {
      return `${hrs}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'long',
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    });
  };

  if (loading && !sessionSummary) {
    return (
      <div className="session-summary-loading">
        <div className="loading-spinner" />
        <p>Loading session summary...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="session-summary-error">
        <p>{error}</p>
        {sessionId && (
          <Button variant="secondary" onClick={() => fetchSessionSummary(sessionId)}>
            Retry
          </Button>
        )}
      </div>
    );
  }

  if (!sessionSummary) {
    return (
      <div className="session-summary-empty">
        <div className="empty-icon">🏆</div>
        <h4>No session summary</h4>
        <p>Complete a workout to see your session summary here.</p>
      </div>
    );
  }

  return (
    <div className="session-summary">
      <Card title="Session Summary">
        <div className="session-summary-content">
          <div className="session-summary-header">
            <span className="session-date">{formatDate(sessionSummary.started_at)}</span>
          </div>

          <div className="session-stats">
            <div className="stat">
              <span className="stat-value">{formatDuration(sessionSummary.duration_seconds)}</span>
              <span className="stat-label">Duration</span>
            </div>
            <div className="stat">
              <span className="stat-value">{sessionSummary.total_volume.toLocaleString()}</span>
              <span className="stat-label">Total Volume (lbs)</span>
            </div>
            <div className="stat">
              <span className="stat-value">{sessionSummary.exercise_count}</span>
              <span className="stat-label">Exercises</span>
            </div>
            <div className="stat">
              <span className="stat-value">{sessionSummary.total_sets}</span>
              <span className="stat-label">Total Sets</span>
            </div>
          </div>

          <div className="exercise-breakdown">
            <h4>Exercise Breakdown</h4>
            <table className="exercise-breakdown-table">
              <thead>
                <tr>
                  <th>Exercise</th>
                  <th>Sets</th>
                  <th>Best Vol</th>
                  <th>Best Wgt</th>
                  <th>Best Reps</th>
                </tr>
              </thead>
              <tbody>
                {sessionSummary.exercises.map((ex) => (
                  <tr key={ex.exercise_id}>
                    <td className="exercise-name">{ex.exercise_name}</td>
                    <td>{ex.sets_count}</td>
                    <td>{ex.best_volume}</td>
                    <td>{ex.best_weight} lbs</td>
                    <td>{ex.best_reps}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Card>
    </div>
  );
}