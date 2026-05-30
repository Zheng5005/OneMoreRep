import { useEffect } from 'react';
import { useToast, type ToastItem } from '../../hooks/useToast';
import './Toast.css';

export function ToastContainer() {
  const { toasts, dismissToast } = useToast();

  return (
    <div className="toast-container" role="region" aria-label="Notifications">
      {toasts.map((toast) => (
        <Toast key={toast.id} toast={toast} onDismiss={() => dismissToast(toast.id)} />
      ))}
    </div>
  );
}

interface ToastProps {
  toast: ToastItem;
  onDismiss: () => void;
}

function Toast({ toast, onDismiss }: ToastProps) {
  useEffect(() => {
    if (toast.duration > 0) {
      const timer = setTimeout(onDismiss, toast.duration);
      return () => clearTimeout(timer);
    }
  }, [toast.duration, onDismiss]);

  const iconMap = {
    success: '✓',
    error: '✕',
    info: 'ℹ',
  };

  return (
    <div className={`toast toast-${toast.type}`} role="alert">
      <span className="toast-icon">{iconMap[toast.type]}</span>
      <span className="toast-message">{toast.message}</span>
      <button className="toast-dismiss" onClick={onDismiss} aria-label="Dismiss">
        ✕
      </button>
    </div>
  );
}
