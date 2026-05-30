import { useState, type FormEvent } from 'react';
import { Input, Textarea } from '../../ui/Input';
import { Button } from '../../ui/Button';
import './ExerciseForm.css';

interface ExerciseFormData {
  name: string;
  target_muscle: string;
  notes?: string;
}

interface ExerciseFormProps {
  initialValues?: Partial<ExerciseFormData>;
  onSubmit: (data: ExerciseFormData) => Promise<void>;
  onCancel: () => void;
  loading?: boolean;
  error?: string;
}

interface FormErrors {
  name?: string;
  target_muscle?: string;
}

export function ExerciseForm({
  initialValues,
  onSubmit,
  onCancel,
  loading,
  error,
}: ExerciseFormProps) {
  const [name, setName] = useState(initialValues?.name || '');
  const [target_muscle, setTargetMuscle] = useState(initialValues?.target_muscle || '');
  const [notes, setNotes] = useState(initialValues?.notes || '');
  const [errors, setErrors] = useState<FormErrors>({});

  const validate = (): boolean => {
    const newErrors: FormErrors = {};

    if (!name.trim()) {
      newErrors.name = 'Name is required';
    } else if (name.length > 100) {
      newErrors.name = 'Name must be 100 characters or less';
    }

    if (!target_muscle.trim()) {
      newErrors.target_muscle = 'Target muscle is required';
    } else if (target_muscle.length > 50) {
      newErrors.target_muscle = 'Target muscle must be 50 characters or less';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    await onSubmit({ name: name.trim(), target_muscle: target_muscle.trim(), notes: notes.trim() || undefined });
  };

  return (
    <form className="exercise-form" onSubmit={handleSubmit}>
      {error && <div className="form-error-alert">{error}</div>}

      <Input
        label="Name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="e.g., Bench Press"
        required
        error={errors.name}
        maxLength={100}
        disabled={loading}
      />

      <Input
        label="Target Muscle"
        value={target_muscle}
        onChange={(e) => setTargetMuscle(e.target.value)}
        placeholder="e.g., Chest"
        required
        error={errors.target_muscle}
        maxLength={50}
        disabled={loading}
      />

      <Textarea
        label="Notes"
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        placeholder="Optional notes about form, equipment, etc."
        disabled={loading}
        rows={3}
      />

      <div className="form-actions">
        <Button variant="secondary" type="button" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
        <Button variant="primary" type="submit" loading={loading}>
          {initialValues ? 'Update' : 'Create'}
        </Button>
      </div>
    </form>
  );
}
