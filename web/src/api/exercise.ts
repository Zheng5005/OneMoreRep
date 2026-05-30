import { api } from './client';
import type { Exercise, LastValues, PaginatedResponse } from '../types';

export const exerciseApi = {
  list: (): Promise<PaginatedResponse<Exercise>> =>
    api.get<PaginatedResponse<Exercise>>('/exercises'),

  get: (id: string): Promise<Exercise> =>
    api.get<Exercise>(`/exercises/${id}`),

  create: (data: { name: string; target_muscle: string; notes?: string }): Promise<Exercise> =>
    api.post<Exercise>('/exercises', data),

  update: (id: string, data: { name?: string; target_muscle?: string; notes?: string }): Promise<Exercise> =>
    api.put<Exercise>(`/exercises/${id}`, data),

  delete: (id: string): Promise<void> =>
    api.delete(`/exercises/${id}`),

  lastValues: (id: string): Promise<LastValues> =>
    api.get<LastValues>(`/exercises/${id}/last-values`),
};
