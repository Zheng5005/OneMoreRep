import { create } from 'zustand';
import { createExerciseSlice, type ExerciseSlice } from './exerciseStore';
import { createRoutineSlice, type RoutineSlice } from './routineStore';
import { createSessionSlice, type SessionSlice } from './sessionStore';
import { createProgressSlice, type ProgressSlice } from './progressStore';
import { createUiSlice, type UiSlice } from './uiStore';

export type AllSlices = ExerciseSlice & RoutineSlice & SessionSlice & ProgressSlice & UiSlice;

export const useStore = create<AllSlices>()((...args) => ({
  ...createExerciseSlice(...args),
  ...createRoutineSlice(...args),
  ...createSessionSlice(...args),
  ...createProgressSlice(...args),
  ...createUiSlice(...args),
}));

export type Store = ReturnType<typeof useStore>;
