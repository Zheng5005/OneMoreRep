import { api } from './client';
import type { Routine, PaginatedResponse } from '../types';

export const routineApi = {
  list: (): Promise<PaginatedResponse<Routine>> =>
    api.get<PaginatedResponse<Routine>>('/routines'),

  get: (id: string): Promise<Routine> =>
    api.get<Routine>(`/routines/${id}`),

  create: (data: { name: string }): Promise<Routine> =>
    api.post<Routine>('/routines', data),

  update: (id: string, data: { name?: string }): Promise<Routine> =>
    api.put<Routine>(`/routines/${id}`, data),

  delete: (id: string): Promise<void> =>
    api.delete(`/routines/${id}`),

  addExercise: (routineId: string, exerciseId: string, order?: number): Promise<Routine> =>
    api.post<Routine>(`/routines/${routineId}/exercises`, { exercise_id: exerciseId, order }),

  removeExercise: (routineId: string, routineExerciseId: string): Promise<void> =>
    api.delete(`/routines/${routineId}/exercises/${routineExerciseId}`),

  reorderExercises: (routineId: string, exerciseIds: string[]): Promise<Routine> =>
    api.put<Routine>(`/routines/${routineId}/exercises/reorder`, { exercise_ids: exerciseIds }),
};
