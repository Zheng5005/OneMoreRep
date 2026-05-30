import { useEffect, useRef, useCallback } from 'react';
import { useStore } from '../stores';
import { quoteApi } from '../api/quote';

export function useRestTimer() {
  const {
    restTimerSeconds,
    restTimerActive,
    quote,
    isFetchingQuote,
    startRestTimer,
    stopRestTimer,
    decrementRestTimer,
    setQuote,
    setFetchingQuote,
  } = useStore();

  const intervalRef = useRef<number | null>(null);

  const fetchNewQuote = useCallback(async () => {
    setFetchingQuote(true);
    try {
      const newQuote = await quoteApi.random();
      setQuote(newQuote);
    } catch {
      setQuote(null);
    } finally {
      setFetchingQuote(false);
    }
  }, [setQuote, setFetchingQuote]);

  const start = useCallback(
    (seconds: number) => {
      startRestTimer(seconds);
    },
    [startRestTimer],
  );

  const stop = useCallback(() => {
    stopRestTimer();
    if (intervalRef.current !== null) {
      window.clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, [stopRestTimer]);

  useEffect(() => {
    if (restTimerActive && restTimerSeconds > 0) {
      intervalRef.current = window.setInterval(() => {
        decrementRestTimer();
      }, 1000);
    } else if (restTimerActive && restTimerSeconds === 0) {
      stopRestTimer();
      fetchNewQuote();
      if (intervalRef.current !== null) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }

    return () => {
      if (intervalRef.current !== null) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [restTimerActive, restTimerSeconds, decrementRestTimer, stopRestTimer, fetchNewQuote]);

  return {
    seconds: restTimerSeconds,
    isActive: restTimerActive,
    quote,
    isFetchingQuote,
    start,
    stop,
  };
}
