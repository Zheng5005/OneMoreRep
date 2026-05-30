import { useState } from 'react';
import { ExerciseList } from './components/features/exercise/ExerciseList';
import { RoutineList } from './components/features/routine/RoutineList';
import './App.css';

type Tab = 'exercises' | 'routines' | 'history';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('exercises');

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
        {activeTab === 'history' && (
          <div className="placeholder">
            <h2>History</h2>
            <p>Coming soon...</p>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
