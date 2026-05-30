import { api } from './client';
import type { Session, SetResponse, SessionSummary } from '../types';

export const sessionApi = {
  getActive: (): Promise<Session & { sets: SetResponse[] } | null> =>
    api.get<Session & { sets: SetResponse[] } | null>('/sessions/active'),

  get: (id: string): Promise<Session & { sets: SetResponse[] }> =>
    api.get<Session & { sets: SetResponse[] }>(`/sessions/${id}`),

  start: (routineId?: string): Promise<Session> =>
    api.post<Session>('/sessions', { routine_id: routineId }),

  end: (id: string): Promise<Session> =>
    api.post<Session>(`/sessions/${id}/end`),

  addSet: (sessionId: string, data: { exercise_id: string; weight: number; reps: number }): Promise<SetResponse> =>
    api.post<SetResponse>(`/sessions/${sessionId}/sets`, data),

  updateSet: (sessionId: string, setId: string, data: { weight?: number; reps?: number }): Promise<SetResponse> =>
    api.put<SetResponse>(`/sessions/${sessionId}/sets/${setId}`, data),

  deleteSet: (sessionId: string, setId: string): Promise<void> =>
    api.delete(`/sessions/${sessionId}/sets/${setId}`),

  summary: (id: string): Promise<SessionSummary> =>
    api.get<SessionSummary>(`/sessions/${id}/summary`),
};
