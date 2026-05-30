import { useState, useEffect } from 'react';
import { useStore } from '../../../stores';
import type { Routine } from '../../../types';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import { Spinner } from '../../ui/Spinner';
import { EmptyState } from '../../shared/EmptyState';
import { RoutineCard } from './RoutineCard';
import { RoutineBuilder } from './RoutineBuilder';
import './RoutineList.css';

export function RoutineList() {
  const { routines, loading, error, fetchRoutines, deleteRoutine, addToast } = useStore();

  const [isBuilderOpen, setIsBuilderOpen] = useState(false);
  const [editingRoutine, setEditingRoutine] = useState<Routine | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<Routine | null>(null);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    fetchRoutines();
  }, [fetchRoutines]);

  const handleCreate = () => {
    setEditingRoutine(null);
    setIsBuilderOpen(true);
  };

  const handleEdit = (routine: Routine) => {
    setEditingRoutine(routine);
    setIsBuilderOpen(true);
  };

  const handleDelete = (routine: Routine) => {
    setDeleteConfirm(routine);
  };

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return;
    setDeleting(true);
    try {
      await deleteRoutine(deleteConfirm.id);
      setDeleteConfirm(null);
      addToast('Routine deleted', 'success');
    } catch {
      addToast('Failed to delete routine', 'error');
    } finally {
      setDeleting(false);
    }
  };

  const handleBuilderSave = () => {
    setIsBuilderOpen(false);
    setEditingRoutine(null);
    addToast('Routine saved successfully', 'success');
  };

  const handleBuilderCancel = () => {
    setIsBuilderOpen(false);
    setEditingRoutine(null);
  };

  const handleRetry = () => {
    fetchRoutines();
  };

  if (isBuilderOpen) {
    return (
      <RoutineBuilder
        routine={editingRoutine ?? undefined}
        onSave={handleBuilderSave}
        onCancel={handleBuilderCancel}
      />
    );
  }

  return (
    <div className="routine-list">
      <div className="routine-list-header">
        <h2 className="routine-list-title">Routines</h2>
        <Button variant="primary" onClick={handleCreate}>
          + Create Routine
        </Button>
      </div>

      {loading && routines.length === 0 ? (
        <div className="routine-list-loading">
          <Spinner size="lg" />
          <p>Loading routines...</p>
        </div>
      ) : error ? (
        <div className="routine-list-error">
          <p>{error}</p>
          <Button variant="secondary" onClick={handleRetry}>
            Retry
          </Button>
        </div>
      ) : routines.length === 0 ? (
        <EmptyState
          title="No routines yet"
          description="Create your first workout plan to get started."
          actionLabel="+ Create Routine"
          onAction={handleCreate}
          icon="📋"
        />
      ) : (
        <div className="routine-grid">
          {routines.map((routine) => (
            <RoutineCard
              key={routine.id}
              routine={routine}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      <Modal
        isOpen={!!deleteConfirm}
        onClose={() => setDeleteConfirm(null)}
        title="Delete Routine"
      >
        <div className="delete-confirm">
          <p>
            Are you sure you want to delete <strong>{deleteConfirm?.name}</strong>?
          </p>
          <p className="delete-warning">
            This will remove all exercises from this routine. This action cannot be
            undone.
          </p>
          <div className="delete-confirm-actions">
            <Button variant="secondary" onClick={() => setDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleDeleteConfirm} loading={deleting}>
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
