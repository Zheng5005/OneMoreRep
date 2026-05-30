import { api } from './client';
import type { VolumePoint, PaginatedResponse } from '../types';

export const progressApi = {
  volume: (period: string): Promise<PaginatedResponse<VolumePoint>> =>
    api.get<PaginatedResponse<VolumePoint>>(`/progress/volume?period=${period}`),

  personalRecords: (exerciseId?: string): Promise<{ exercise_id: string; weight: number; reps: number; achieved_at: string }[]> => {
    const query = exerciseId ? `?exercise_id=${exerciseId}` : '';
    return api.get(`/progress/personal-records${query}`);
  },
};
