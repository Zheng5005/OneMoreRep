import { type StateCreator } from 'zustand';
import { exerciseApi } from '../api/exercise';
import type { Exercise, LastValues } from '../types';
import type { AllSlices } from './index';

export interface ExerciseSlice {
  exercises: Exercise[];
  selectedExercise: Exercise | null;
  loading: boolean;
  error: string | null;
  fetchExercises: () => Promise<void>;
  getExercise: (id: string) => Promise<Exercise>;
  createExercise: (data: { name: string; target_muscle: string; notes?: string }) => Promise<Exercise>;
  updateExercise: (id: string, data: { name?: string; target_muscle?: string; notes?: string }) => Promise<Exercise>;
  deleteExercise: (id: string) => Promise<void>;
  selectExercise: (exercise: Exercise | null) => void;
  getLastValues: (id: string) => Promise<LastValues>;
}

export const createExerciseSlice: StateCreator<AllSlices, [], [], ExerciseSlice> = (set) => ({
  exercises: [],
  selectedExercise: null,
  loading: false,
  error: null,

  fetchExercises: async () => {
    set({ loading: true, error: null });
    try {
      const response = await exerciseApi.list();
      set({ exercises: response.data });
    } catch (e) {
      set({ error: e instanceof Error ? e.message : 'Failed to fetch exercises' });
    } finally {
      set({ loading: false });
    }
  },

  getExercise: async (id: string) => {
    return exerciseApi.get(id);
  },

  createExercise: async (data) => {
    const exercise = await exerciseApi.create(data);
    set((state) => ({ exercises: [...state.exercises, exercise] }));
    return exercise;
  },

  updateExercise: async (id, data) => {
    const exercise = await exerciseApi.update(id, data);
    set((state) => ({
      exercises: state.exercises.map((e) => (e.id === id ? exercise : e)),
      selectedExercise: state.selectedExercise?.id === id ? exercise : state.selectedExercise,
    }));
    return exercise;
  },

  deleteExercise: async (id) => {
    await exerciseApi.delete(id);
    set((state) => ({
      exercises: state.exercises.filter((e) => e.id !== id),
      selectedExercise: state.selectedExercise?.id === id ? null : state.selectedExercise,
    }));
  },

  selectExercise: (exercise) => {
    set({ selectedExercise: exercise });
  },

  getLastValues: async (id) => {
    return exerciseApi.lastValues(id);
  },
});
