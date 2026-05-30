import { type StateCreator } from 'zustand';
import type { Quote } from '../types';
import type { AllSlices } from './index';

let toastIdCounter = 0;

export interface ToastItem {
  id: string;
  message: string;
  type: 'success' | 'error' | 'info';
  duration: number;
}

export interface UiSlice {
  restTimerSeconds: number;
  restTimerActive: boolean;
  quote: Quote | null;
  isFetchingQuote: boolean;
  toasts: ToastItem[];
  setRestTimer: (seconds: number) => void;
  startRestTimer: (seconds: number) => void;
  stopRestTimer: () => void;
  decrementRestTimer: () => void;
  setQuote: (quote: Quote | null) => void;
  setFetchingQuote: (fetching: boolean) => void;
  addToast: (message: string, type: 'success' | 'error' | 'info', duration?: number) => void;
  removeToast: (id: string) => void;
  clearToasts: () => void;
}

export const createUiSlice: StateCreator<AllSlices, [], [], UiSlice> = (set) => ({
  restTimerSeconds: 0,
  restTimerActive: false,
  quote: null,
  isFetchingQuote: false,
  toasts: [],

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

  addToast: (message, type, duration = 4000) => {
    const id = `toast-${++toastIdCounter}`;
    set((state) => ({
      toasts: [...state.toasts, { id, message, type, duration }],
    }));
  },

  removeToast: (id) => {
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    }));
  },

  clearToasts: () => {
    set({ toasts: [] });
  },
});
