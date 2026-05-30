import { useEffect, useState } from 'react';
import { useStore } from '../../../stores';
import { Button } from '../../ui/Button';
import type { HistorySession } from '../../../types';
import './ExerciseHistory.css';

type TimeFilter = 'all' | '30d' | '6m';

interface ExerciseHistoryProps {
  exerciseId: string;
}

export function ExerciseHistory({ exerciseId }: ExerciseHistoryProps) {
  const { history, loading, error, fetchHistory } = useStore();
  const [filter, setFilter] = useState<TimeFilter>('all');

  useEffect(() => {
    fetchHistory(exerciseId, filter);
  }, [exerciseId, filter, fetchHistory]);

  const handleFilterChange = (newFilter: TimeFilter) => {
    setFilter(newFilter);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const formatDuration = (start: string, end: string | null) => {
    if (!end) return 'In progress';
    const ms = new Date(end).getTime() - new Date(start).getTime();
    const mins = Math.floor(ms / 60000);
    if (mins < 60) return `${mins}m`;
    const hrs = Math.floor(mins / 60);
    const remainingMins = mins % 60;
    return `${hrs}h ${remainingMins}m`;
  };

  if (loading && history.length === 0) {
    return (
      <div className="exercise-history-loading">
        <div className="loading-spinner" />
        <p>Loading history...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="exercise-history-error">
        <p>{error}</p>
        <Button variant="secondary" onClick={() => fetchHistory(exerciseId, filter)}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="exercise-history">
      <div className="exercise-history-header">
        <h3>Exercise History</h3>
        <div className="exercise-history-filters">
          <Button
            variant={filter === 'all' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('all')}
          >
            All
          </Button>
          <Button
            variant={filter === '30d' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('30d')}
          >
            30 Days
          </Button>
          <Button
            variant={filter === '6m' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('6m')}
          >
            6 Months
          </Button>
        </div>
      </div>

      {history.length === 0 ? (
        <div className="exercise-history-empty">
          <div className="empty-icon">📊</div>
          <h4>No history yet</h4>
          <p>Complete some workouts with this exercise to see your history.</p>
        </div>
      ) : (
        <div className="exercise-history-list">
          {history.map((session) => (
            <SessionGroup key={session.session_id} session={session} formatDate={formatDate} formatDuration={formatDuration} />
          ))}
        </div>
      )}
    </div>
  );
}

interface SessionGroupProps {
  session: HistorySession;
  formatDate: (date: string) => string;
  formatDuration: (start: string, end: string | null) => string;
}

function SessionGroup({ session, formatDate, formatDuration }: SessionGroupProps) {
  return (
    <div className="history-session">
      <div className="history-session-header">
        <span className="history-session-date">{formatDate(session.started_at)}</span>
        <span className="history-session-duration">{formatDuration(session.started_at, session.ended_at)}</span>
      </div>
      <div className="history-sets">
        <div className="history-set-header">
          <span>Set</span>
          <span>Weight</span>
          <span>Reps</span>
          <span>Volume</span>
        </div>
        {session.sets.map((set) => (
          <div key={set.set_id} className={`history-set ${set.is_pr ? 'history-set-pr' : ''}`}>
            <span className="history-set-number">
              {set.is_pr && <span className="pr-badge">PR</span>}
              {set.set_number}
            </span>
            <span className="history-set-weight">{set.weight} lbs</span>
            <span className="history-set-reps">{set.reps}</span>
            <span className="history-set-volume">{set.volume}</span>
          </div>
        ))}
      </div>
    </div>
  );
}