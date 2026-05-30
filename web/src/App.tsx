import { useState, useEffect } from 'react';
import { ExerciseList } from './components/features/exercise/ExerciseList';
import { RoutineList } from './components/features/routine/RoutineList';
import { ActiveSession } from './components/features/workout/ActiveSession';
import { StartWorkoutModal } from './components/features/workout/StartWorkoutModal';
import { Button } from './components/ui/Button';
import { useStore } from './stores';
import './App.css';

type Tab = 'exercises' | 'routines' | 'history' | 'workout';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('exercises');
  const [startWorkoutOpen, setStartWorkoutOpen] = useState(false);
  const { currentSession, hydrateActiveSession, sessionSummary } = useStore();

  useEffect(() => {
    hydrateActiveSession();
  }, [hydrateActiveSession]);

  const handleStartWorkout = () => {
    setActiveTab('workout');
  };

  const handleEndWorkout = () => {
    setActiveTab('history');
  };

  const showWorkoutTab = activeTab === 'workout' || (currentSession !== null && activeTab !== 'workout');

  return (
    <div className="app">
      <nav className="app-nav">
        <div className="nav-brand">OneMoreRep</div>
        <div className="nav-tabs">
          <button
            className={`nav-tab ${activeTab === 'exercises' ? 'nav-tab-active' : ''}`}
            onClick={() => setActiveTab('exercises')}
          >
            Exercises
          </button>
          <button
            className={`nav-tab ${activeTab === 'routines' ? 'nav-tab-active' : ''}`}
            onClick={() => setActiveTab('routines')}
          >
            Routines
          </button>
          <button
            className={`nav-tab ${activeTab === 'workout' ? 'nav-tab-active' : ''}`}
            onClick={() => setActiveTab('workout')}
          >
            Workout
          </button>
          <button
            className={`nav-tab ${activeTab === 'history' ? 'nav-tab-active' : ''}`}
            onClick={() => setActiveTab('history')}
          >
            History
          </button>
        </div>
      </nav>

      <main className="app-main">
        {activeTab === 'exercises' && <ExerciseList />}
        {activeTab === 'routines' && <RoutineList />}
        {showWorkoutTab && (
          currentSession ? (
            <ActiveSession onEnd={handleEndWorkout} />
          ) : (
            <div className="workout-placeholder">
              <h2>Ready to Train?</h2>
              <p>Start a workout to begin logging your sets.</p>
              <Button variant="primary" onClick={() => setStartWorkoutOpen(true)}>
                Start Workout
              </Button>
            </div>
          )
        )}
        {activeTab === 'history' && (
          sessionSummary ? (
            <div className="workout-summary">
              <h2>Workout Complete!</h2>
              <div className="summary-stats">
                <p>Duration: {Math.floor(sessionSummary.duration_seconds / 60)} minutes</p>
                <p>Total Volume: {sessionSummary.total_volume.toLocaleString()} lbs</p>
                <p>Exercises: {sessionSummary.exercise_count}</p>
                <p>Sets: {sessionSummary.total_sets}</p>
              </div>
            </div>
          ) : (
            <div className="placeholder">
              <h2>History</h2>
              <p>Complete a workout to see your summary here.</p>
            </div>
          )
        )}
      </main>

      <StartWorkoutModal
        isOpen={startWorkoutOpen}
        onClose={() => setStartWorkoutOpen(false)}
        onStart={handleStartWorkout}
      />
    </div>
  );
}

export default App;
