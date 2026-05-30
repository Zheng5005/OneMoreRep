import { type StateCreator } from 'zustand';
import { progressApi } from '../api/progress';
import type { VolumePoint, HistorySession, LastValues } from '../types';
import type { AllSlices } from './index';

export interface ProgressSlice {
  history: HistorySession[];
  volumeData: VolumePoint[];
  personalRecords: { exercise_id: string; weight: number; reps: number; achieved_at: string }[];
  loading: boolean;
  error: string | null;
  lastValues: LastValues | null;
  fetchVolume: (groupBy: 'session' | 'week' | 'month', exerciseId?: string) => Promise<void>;
  fetchPersonalRecords: (exerciseId?: string) => Promise<void>;
  fetchHistory: (exerciseId: string, filter?: 'all' | '30d' | '6m') => Promise<void>;
  fetchLastValues: (exerciseId: string) => Promise<void>;
}

export const createProgressSlice: StateCreator<AllSlices, [], [], ProgressSlice> = (set) => ({
  history: [],
  volumeData: [],
  personalRecords: [],
  loading: false,
  error: null,
  lastValues: null,

  fetchVolume: async (groupBy, exerciseId) => {
    set({ loading: true, error: null });
    try {
      const data = await progressApi.getVolume(groupBy, exerciseId);
      set({ volumeData: data });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch volume data' });
    } finally {
      set({ loading: false });
    }
  },

  fetchPersonalRecords: async (exerciseId) => {
    set({ loading: true, error: null });
    try {
      const records = await progressApi.personalRecords(exerciseId);
      set({ personalRecords: records });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch personal records' });
    } finally {
      set({ loading: false });
    }
  },

  fetchHistory: async (exerciseId, filter) => {
    set({ loading: true, error: null });
    try {
      const history = await progressApi.getExerciseHistory(exerciseId, filter);
      set({ history });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch history' });
    } finally {
      set({ loading: false });
    }
  },

  fetchLastValues: async (exerciseId) => {
    set({ loading: true, error: null });
    try {
      const values = await progressApi.getLastValues(exerciseId);
      set({ lastValues: values });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch last values' });
    } finally {
      set({ loading: false });
    }
  },
});