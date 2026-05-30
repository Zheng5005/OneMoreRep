import { useEffect, useCallback, useRef } from 'react';
import { useStore } from '../stores';

export type ToastType = 'success' | 'error' | 'info';

export interface ToastItem {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
}

interface ToastSlice {
  toasts: ToastItem[];
  addToast: (message: string, type: ToastType, duration?: number) => void;
  removeToast: (id: string) => void;
  clearToasts: () => void;
}

let toastIdCounter = 0;

export function useToast() {
  const { toasts, addToast, removeToast, clearToasts } = useStore();

  const showToast = useCallback(
    (message: string, type: ToastType = 'info', duration = 4000) => {
      addToast(message, type, duration);
    },
    [addToast],
  );

  const dismissToast = useCallback(
    (id: string) => {
      removeToast(id);
    },
    [removeToast],
  );

  const dismissAll = useCallback(() => {
    clearToasts();
  }, [clearToasts]);

  return { toasts, showToast, dismissToast, dismissAll };
}

export function createToastSlice(): ToastSlice {
  return {
    toasts: [],
    addToast: (message, type, duration = 4000) => {
      const id = `toast-${++toastIdCounter}`;
      useStore.setState((state) => ({
        toasts: [...state.toasts, { id, message, type, duration }],
      }));
    },
    removeToast: (id) => {
      useStore.setState((state) => ({
        toasts: state.toasts.filter((t) => t.id !== id),
      }));
    },
    clearToasts: () => {
      useStore.setState({ toasts: [] });
    },
  };
}
