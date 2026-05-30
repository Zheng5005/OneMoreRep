import { useState, useCallback, useEffect } from 'react';
import { exerciseApi } from '../api/exercise';
import type { LastValues } from '../types';

interface UseAutoFillResult {
  lastValues: LastValues | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useAutoFill(exerciseId: string | null): UseAutoFillResult {
  const [lastValues, setLastValues] = useState<LastValues | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLastValues = useCallback(async () => {
    if (!exerciseId) {
      setLastValues(null);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const values = await exerciseApi.lastValues(exerciseId);
      setLastValues(values);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch last values');
      setLastValues(null);
    } finally {
      setLoading(false);
    }
  }, [exerciseId]);

  useEffect(() => {
    fetchLastValues();
  }, [fetchLastValues]);

  return {
    lastValues,
    loading,
    error,
    refetch: fetchLastValues,
  };
}
