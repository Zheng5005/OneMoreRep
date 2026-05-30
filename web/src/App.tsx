import { useState, useEffect, Component, type ReactNode } from 'react';
import { ExerciseList } from './components/features/exercise/ExerciseList';
import { RoutineList } from './components/features/routine/RoutineList';
import { ActiveSession } from './components/features/workout/ActiveSession';
import { StartWorkoutModal } from './components/features/workout/StartWorkoutModal';
import { ProgressScreen } from './components/features/progress/ProgressScreen';
import { Button } from './components/ui/Button';
import { ToastContainer } from './components/ui/Toast';
import { useStore } from './stores';
import './App.css';

type Tab = 'exercises' | 'routines' | 'history' | 'workout' | 'progress';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

interface ErrorBoundaryProps {
  children: ReactNode;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('App Error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <h2>Something went wrong</h2>
          <p>We encountered an unexpected error. Please try refreshing the page.</p>
          <Button
            variant="primary"
            onClick={() => this.setState({ hasError: false, error: null })}
          >
            Try Again
          </Button>
        </div>
      );
    }
    return this.props.children;
  }
}

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('exercises');
  const [startWorkoutOpen, setStartWorkoutOpen] = useState(false);
  const { currentSession, hydrateActiveSession, sessionSummary, addToast } = useStore();

  useEffect(() => {
    hydrateActiveSession();
  }, [hydrateActiveSession]);

  const handleStartWorkout = () => {
    setActiveTab('workout');
  };

  const handleEndWorkout = () => {
    setActiveTab('history');
    addToast('Workout completed! Great job!', 'success');
  };

  const isWorkoutTab = activeTab === 'workout';
  const showWorkoutTab = isWorkoutTab || (currentSession !== null && !isWorkoutTab);

  return (
    <ErrorBoundary>
      <div className="app">
        <nav className="app-nav">
          <div className="nav-brand">OneMoreRep</div>
          <div className="nav-tabs">
            <button
              className={`nav-tab ${activeTab === 'exercises' ? 'nav-tab-active' : ''}`}
              onClick={() => setActiveTab('exercises')}
              aria-current={activeTab === 'exercises' ? 'page' : undefined}
            >
              Exercises
            </button>
            <button
              className={`nav-tab ${activeTab === 'routines' ? 'nav-tab-active' : ''}`}
              onClick={() => setActiveTab('routines')}
              aria-current={activeTab === 'routines' ? 'page' : undefined}
            >
              Routines
            </button>
            <button
              className={`nav-tab ${activeTab === 'workout' ? 'nav-tab-active' : ''}`}
              onClick={() => setActiveTab('workout')}
              aria-current={activeTab === 'workout' ? 'page' : undefined}
            >
              Workout
            </button>
            <button
              className={`nav-tab ${activeTab === 'history' ? 'nav-tab-active' : ''}`}
              onClick={() => setActiveTab('history')}
              aria-current={activeTab === 'history' ? 'page' : undefined}
            >
              History
            </button>
            <button
              className={`nav-tab ${activeTab === 'progress' ? 'nav-tab-active' : ''}`}
              onClick={() => setActiveTab('progress')}
              aria-current={activeTab === 'progress' ? 'page' : undefined}
            >
              Progress
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
          {activeTab === 'progress' && <ProgressScreen />}
        </main>

        <StartWorkoutModal
          isOpen={startWorkoutOpen}
          onClose={() => setStartWorkoutOpen(false)}
          onStart={handleStartWorkout}
        />

        <ToastContainer />
      </div>
    </ErrorBoundary>
  );
}

export default App;
