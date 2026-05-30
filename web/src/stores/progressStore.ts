import { type StateCreator } from 'zustand';
import { progressApi } from '../api/progress';
import type { VolumePoint } from '../types';
import type { AllSlices } from './index';

export interface ProgressSlice {
  volumeData: VolumePoint[];
  personalRecords: { exercise_id: string; weight: number; reps: number; achieved_at: string }[];
  loading: boolean;
  error: string | null;
  fetchVolume: (period: string) => Promise<void>;
  fetchPersonalRecords: (exerciseId?: string) => Promise<void>;
}

export const createProgressSlice: StateCreator<AllSlices, [], [], ProgressSlice> = (set) => ({
  volumeData: [],
  personalRecords: [],
  loading: false,
  error: null,

  fetchVolume: async (period) => {
    set({ loading: true, error: null });
    try {
      const response = await progressApi.volume(period);
      set({ volumeData: response.data });
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
});
