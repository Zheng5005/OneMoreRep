import { useState, useEffect } from 'react';
import { useStore } from '../../../stores';
import { Button } from '../../ui/Button';
import { Input } from '../../ui/Input';
import { Modal } from '../../ui/Modal';
import { Spinner } from '../../ui/Spinner';
import { EmptyState } from '../../shared/EmptyState';
import { ExerciseCard } from './ExerciseCard';
import { ExerciseForm } from './ExerciseForm';
import type { Exercise } from '../../../types';
import './ExerciseList.css';

const PAGE_SIZE = 10;

export function ExerciseList() {
  const { exercises, loading, error, fetchExercises, createExercise, updateExercise, deleteExercise, addToast } = useStore();

  const [search, setSearch] = useState('');
  const [page, setPage] = useState(0);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingExercise, setEditingExercise] = useState<Exercise | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<Exercise | null>(null);
  const [formError, setFormError] = useState<string | undefined>();
  const [formLoading, setFormLoading] = useState(false);

  useEffect(() => {
    fetchExercises();
  }, [fetchExercises]);

  const filteredExercises = exercises.filter((ex) =>
    ex.name.toLowerCase().includes(search.toLowerCase())
  );

  const totalPages = Math.ceil(filteredExercises.length / PAGE_SIZE);
  const paginatedExercises = filteredExercises.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
    setPage(0);
  };

  const handleCreate = () => {
    setEditingExercise(null);
    setFormError(undefined);
    setIsModalOpen(true);
  };

  const handleEdit = (exercise: Exercise) => {
    setEditingExercise(exercise);
    setFormError(undefined);
    setIsModalOpen(true);
  };

  const handleDelete = (exercise: Exercise) => {
    setDeleteConfirm(exercise);
  };

  const handleFormSubmit = async (data: { name: string; target_muscle: string; notes?: string }) => {
    setFormLoading(true);
    setFormError(undefined);
    try {
      if (editingExercise) {
        await updateExercise(editingExercise.id, data);
        addToast('Exercise updated successfully', 'success');
      } else {
        await createExercise(data);
        addToast('Exercise created successfully', 'success');
      }
      setIsModalOpen(false);
      setEditingExercise(null);
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'An error occurred');
      addToast(e instanceof Error ? e.message : 'Failed to save exercise', 'error');
    } finally {
      setFormLoading(false);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return;
    try {
      await deleteExercise(deleteConfirm.id);
      setDeleteConfirm(null);
      addToast('Exercise deleted', 'success');
    } catch {
      setDeleteConfirm(null);
      addToast('Failed to delete exercise', 'error');
    }
  };

  const handleRetry = () => {
    fetchExercises();
  };

  return (
    <div className="exercise-list">
      <div className="exercise-list-header">
        <h2 className="exercise-list-title">Exercise Library</h2>
        <Button variant="primary" onClick={handleCreate}>
          + Add Exercise
        </Button>
      </div>

      <div className="exercise-list-toolbar">
          <Input
          placeholder="Search exercises..."
          value={search}
          onChange={handleSearchChange}
          className="exercise-search"
        />
      </div>

      {loading && exercises.length === 0 ? (
        <div className="exercise-list-loading">
          <Spinner size="lg" />
          <p>Loading exercises...</p>
        </div>
      ) : error ? (
        <div className="exercise-list-error">
          <p>{error}</p>
          <Button variant="secondary" onClick={handleRetry}>
            Retry
          </Button>
        </div>
      ) : filteredExercises.length === 0 ? (
        <EmptyState
          title="No exercises yet"
          description="Build your library by adding your first exercise."
          actionLabel="+ Add Exercise"
          onAction={handleCreate}
          icon="🏋️"
        />
      ) : (
        <>
          <div className="exercise-grid">
            {paginatedExercises.map((exercise) => (
              <ExerciseCard
                key={exercise.id}
                exercise={exercise}
                onEdit={handleEdit}
                onDelete={handleDelete}
              />
            ))}
          </div>

          {totalPages > 1 && (
            <div className="exercise-pagination">
              <Button
                variant="secondary"
                size="sm"
                disabled={page === 0}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <span className="pagination-info">
                Page {page + 1} of {totalPages}
              </span>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages - 1}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingExercise ? 'Edit Exercise' : 'New Exercise'}
      >
        <ExerciseForm
          initialValues={editingExercise ? { name: editingExercise.name, target_muscle: editingExercise.target_muscle, notes: editingExercise.notes || undefined } : undefined}
          onSubmit={handleFormSubmit}
          onCancel={() => setIsModalOpen(false)}
          loading={formLoading}
          error={formError}
        />
      </Modal>

      <Modal
        isOpen={!!deleteConfirm}
        onClose={() => setDeleteConfirm(null)}
        title="Delete Exercise"
      >
        <div className="delete-confirm">
          <p>Are you sure you want to delete <strong>{deleteConfirm?.name}</strong>?</p>
          <p className="delete-warning">This action cannot be undone.</p>
          <div className="delete-confirm-actions">
            <Button variant="secondary" onClick={() => setDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleDeleteConfirm}>
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
