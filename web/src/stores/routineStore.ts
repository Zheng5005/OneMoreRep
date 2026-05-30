import { type StateCreator } from 'zustand';
import { routineApi } from '../api/routine';
import type { Routine } from '../types';
import type { AllSlices } from './index';

export interface RoutineSlice {
  routines: Routine[];
  selectedRoutine: Routine | null;
  loading: boolean;
  error: string | null;
  fetchRoutines: () => Promise<void>;
  getRoutine: (id: string) => Promise<Routine>;
  createRoutine: (data: { name: string }) => Promise<Routine>;
  updateRoutine: (id: string, data: { name?: string }) => Promise<Routine>;
  deleteRoutine: (id: string) => Promise<void>;
  selectRoutine: (routine: Routine | null) => void;
  addExerciseToRoutine: (routineId: string, exerciseId: string, order?: number) => Promise<Routine>;
  removeExerciseFromRoutine: (routineId: string, routineExerciseId: string) => Promise<void>;
}

export const createRoutineSlice: StateCreator<AllSlices, [], [], RoutineSlice> = (set) => ({
  routines: [],
  selectedRoutine: null,
  loading: false,
  error: null,

  fetchRoutines: async () => {
    set({ loading: true, error: null });
    try {
      const response = await routineApi.list();
      set({ routines: response.data });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch routines' });
    } finally {
      set({ loading: false });
    }
  },

  getRoutine: async (id) => {
    return routineApi.get(id);
  },

  createRoutine: async (data) => {
    const routine = await routineApi.create(data);
    set((state) => ({ routines: [...state.routines, routine] }));
    return routine;
  },

  updateRoutine: async (id, data) => {
    const routine = await routineApi.update(id, data);
    set((state) => ({
      routines: state.routines.map((r) => (r.id === id ? routine : r)),
      selectedRoutine: state.selectedRoutine?.id === id ? routine : state.selectedRoutine,
    }));
    return routine;
  },

  deleteRoutine: async (id) => {
    await routineApi.delete(id);
    set((state) => ({
      routines: state.routines.filter((r) => r.id !== id),
      selectedRoutine: state.selectedRoutine?.id === id ? null : state.selectedRoutine,
    }));
  },

  selectRoutine: (routine) => {
    set({ selectedRoutine: routine });
  },

  addExerciseToRoutine: async (routineId, exerciseId, order) => {
    const routine = await routineApi.addExercise(routineId, exerciseId, order);
    set((state) => ({
      routines: state.routines.map((r) => (r.id === routineId ? routine : r)),
      selectedRoutine: state.selectedRoutine?.id === routineId ? routine : state.selectedRoutine,
    }));
    return routine;
  },

  removeExerciseFromRoutine: async (routineId, routineExerciseId) => {
    await routineApi.removeExercise(routineId, routineExerciseId);
    const routine = await routineApi.get(routineId);
    set((state) => ({
      routines: state.routines.map((r) => (r.id === routineId ? routine : r)),
      selectedRoutine: state.selectedRoutine?.id === routineId ? routine : state.selectedRoutine,
    }));
  },
});
