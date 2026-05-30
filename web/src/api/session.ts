import { api } from './client';
import type { Session, SetResponse, SessionSummary, HistorySession, PaginatedResponse } from '../types';

export const sessionApi = {
  getActive: (): Promise<{ data: Session | null }> =>
    api.get<{ data: Session | null }>('/sessions/active'),

  getWithSets: (id: string): Promise<Session & { sets: SetResponse[] }> =>
    api.get<Session & { sets: SetResponse[] }>(`/sessions/${id}/with-sets`),

  list: (params?: { limit?: number; offset?: number }): Promise<PaginatedResponse<Session>> => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    const qs = query.toString();
    return api.get<PaginatedResponse<Session>>(`/sessions${qs ? `?${qs}` : ''}`);
  },

  get: (id: string): Promise<Session> =>
    api.get<Session>(`/sessions/${id}`),

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

  history: (params?: { limit?: number; offset?: number }): Promise<PaginatedResponse<HistorySession>> => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    const qs = query.toString();
    return api.get<PaginatedResponse<HistorySession>>(`/sessions/history${qs ? `?${qs}` : ''}`);
  },
};
