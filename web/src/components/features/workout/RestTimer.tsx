import { Button } from '../../ui/Button';
import { useRestTimer } from '../../../hooks/useRestTimer';
import './RestTimer.css';

interface RestTimerProps {
  defaultSeconds?: number;
  onDismiss?: () => void;
}

export function RestTimer({ defaultSeconds = 60, onDismiss }: RestTimerProps) {
  const { seconds, isActive, quote, isFetchingQuote, start, stop } = useRestTimer();

  const handleStart = () => {
    start(defaultSeconds);
  };

  const formatTime = (secs: number) => {
    const mins = Math.floor(secs / 60);
    const remainingSecs = secs % 60;
    return `${mins}:${remainingSecs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="rest-timer">
      <div className="rest-timer-header">
        <h3>Rest Timer</h3>
        {onDismiss && (
          <Button variant="ghost" size="sm" onClick={onDismiss}>
            ✕
          </Button>
        )}
      </div>

      <div className="rest-timer-display">
        {isActive || seconds > 0 ? (
          <span className="rest-timer-time">{formatTime(seconds)}</span>
        ) : (
          <span className="rest-timer-idle">Ready</span>
        )}
      </div>

      <div className="rest-timer-controls">
        {!isActive && seconds === 0 ? (
          <Button variant="primary" onClick={handleStart}>
            Start ({defaultSeconds}s)
          </Button>
        ) : (
          <>
            <Button variant="secondary" onClick={stop}>
              {isActive ? 'Stop' : 'Reset'}
            </Button>
          </>
        )}
      </div>

      {quote && (
        <div className="rest-timer-quote">
          {isFetchingQuote ? (
            <p className="quote-loading">Loading inspiration...</p>
          ) : (
            <>
              <p className="quote-text">"{quote.text}"</p>
              <p className="quote-author">— {quote.author}</p>
            </>
          )}
        </div>
      )}
    </div>
  );
}
