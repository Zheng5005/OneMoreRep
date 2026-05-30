import { api } from './client';
import type { VolumePoint, PaginatedResponse, HistorySession, SessionSummary, LastValues } from '../types';

export const progressApi = {
  volume: (period: string): Promise<PaginatedResponse<VolumePoint>> =>
    api.get<PaginatedResponse<VolumePoint>>(`/progress/volume?period=${period}`),

  personalRecords: (exerciseId?: string): Promise<{ exercise_id: string; weight: number; reps: number; achieved_at: string }[]> => {
    const query = exerciseId ? `?exercise_id=${exerciseId}` : '';
    return api.get(`/progress/personal-records${query}`);
  },

  getLastValues: (exerciseId: string): Promise<LastValues> =>
    api.get<LastValues>(`/exercises/${exerciseId}/last-values`),

  getSessionSummary: (sessionId: string): Promise<SessionSummary> =>
    api.get<SessionSummary>(`/sessions/${sessionId}/summary`),

  getExerciseHistory: (exerciseId: string, filter?: 'all' | '30d' | '6m'): Promise<HistorySession[]> =>
    api.get<HistorySession[]>(`/exercises/${exerciseId}/history?filter=${filter || 'all'}`),

  getVolume: (groupBy: 'session' | 'week' | 'month', exerciseId?: string): Promise<VolumePoint[]> => {
    const params = new URLSearchParams({ group_by: groupBy });
    if (exerciseId) params.append('exercise_id', exerciseId);
    return api.get<VolumePoint[]>(`/progress/volume?${params.toString()}`);
  },
};
