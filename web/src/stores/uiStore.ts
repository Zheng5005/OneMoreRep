import { type StateCreator } from 'zustand';
import type { Quote } from '../types';
import type { AllSlices } from './index';

export interface UiSlice {
  restTimerSeconds: number;
  restTimerActive: boolean;
  quote: Quote | null;
  isFetchingQuote: boolean;
  setRestTimer: (seconds: number) => void;
  startRestTimer: (seconds: number) => void;
  stopRestTimer: () => void;
  decrementRestTimer: () => void;
  setQuote: (quote: Quote | null) => void;
  setFetchingQuote: (fetching: boolean) => void;
}

export const createUiSlice: StateCreator<AllSlices, [], [], UiSlice> = (set) => ({
  restTimerSeconds: 0,
  restTimerActive: false,
  quote: null,
  isFetchingQuote: false,

  setRestTimer: (seconds) => {
    set({ restTimerSeconds: seconds });
  },

  startRestTimer: (seconds) => {
    set({ restTimerSeconds: seconds, restTimerActive: true });
  },

  stopRestTimer: () => {
    set({ restTimerActive: false });
  },

  decrementRestTimer: () => {
    set((state) => {
      const newSeconds = state.restTimerSeconds - 1;
      if (newSeconds <= 0) {
        return { restTimerSeconds: 0, restTimerActive: false };
      }
      return { restTimerSeconds: newSeconds };
    });
  },

  setQuote: (quote) => {
    set({ quote });
  },

  setFetchingQuote: (fetching) => {
    set({ isFetchingQuote: fetching });
  },
});
