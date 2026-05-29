package service

import (
	"context"
	"strings"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/Zheng5005/onemorerep/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// ExerciseStore defines the store operations required by ExerciseService.
type ExerciseStore interface {
	CreateExercise(ctx context.Context, arg queries.CreateExerciseParams) (queries.Exercise, error)
	SearchExercises(ctx context.Context, arg queries.SearchExercisesParams) ([]queries.Exercise, error)
	CountExercises(ctx context.Context, search string) (int64, error)
	GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	GetExerciseByNameAndMuscle(ctx context.Context, arg queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error)
	UpdateExercise(ctx context.Context, arg queries.UpdateExerciseParams) (queries.Exercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID) error
	CountRoutineExercisesByExercise(ctx context.Context, exerciseID uuid.UUID) (int64, error)
	CountWorkoutSetsByExercise(ctx context.Context, exerciseID uuid.UUID) (int64, error)
}

// ExerciseListResult wraps the list response.
type ExerciseListResult struct {
	Data   []queries.Exercise
	Limit  int32
	Offset int32
	Total  int64
}

// ExerciseService provides business logic for exercises.
type ExerciseService struct {
	store ExerciseStore
}

// NewExerciseService creates a new ExerciseService.
func NewExerciseService(store ExerciseStore) *ExerciseService {
	return &ExerciseService{store: store}
}

func validationErrorFromOzzo(ve validation.Errors) *AppError {
	if err, ok := ve["name"]; ok {
		return NewValidationError("name", err.Error())
	}
	if err, ok := ve["target_muscle"]; ok {
		return NewValidationError("target_muscle", err.Error())
	}
	if err, ok := ve["notes"]; ok {
		return NewValidationError("notes", err.Error())
	}
	for field, err := range ve {
		return NewValidationError(field, err.Error())
	}
	return NewValidationError("unknown", "validation failed")
}

// CreateExercise creates a new exercise after validation and duplicate checks.
func (s *ExerciseService) CreateExercise(ctx context.Context, name, targetMuscle, notes string) (queries.Exercise, error) {
	name = strings.TrimSpace(name)
	targetMuscle = strings.TrimSpace(targetMuscle)
	notes = strings.TrimSpace(notes)

	if err := validator.ValidateExercise(name, targetMuscle, notes); err != nil {
		if ve, ok := err.(validation.Errors); ok {
			return queries.Exercise{}, validationErrorFromOzzo(ve)
		}
		return queries.Exercise{}, NewValidationError("unknown", err.Error())
	}

	_, err := s.store.GetExerciseByNameAndMuscle(ctx, queries.GetExerciseByNameAndMuscleParams{
		Name:         name,
		TargetMuscle: pgtype.Text{String: targetMuscle, Valid: true},
	})
	if err == nil {
		return queries.Exercise{}, NewConflictError("exercise with same name and target muscle already exists")
	}
	if err != pgx.ErrNoRows {
		return queries.Exercise{}, err
	}

	ex, err := s.store.CreateExercise(ctx, queries.CreateExerciseParams{
		Name:         name,
		TargetMuscle: pgtype.Text{String: targetMuscle, Valid: true},
		Notes:        pgtype.Text{String: notes, Valid: notes != ""},
	})
	if err != nil {
		return queries.Exercise{}, err
	}
	return ex, nil
}

// ListExercises returns a paginated list of exercises with optional search.
func (s *ExerciseService) ListExercises(ctx context.Context, limit, offset int32, search string) (ExerciseListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	search = strings.TrimSpace(search)

	exercises, err := s.store.SearchExercises(ctx, queries.SearchExercisesParams{
		Column1: search,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return ExerciseListResult{}, err
	}

	total, err := s.store.CountExercises(ctx, search)
	if err != nil {
		return ExerciseListResult{}, err
	}

	return ExerciseListResult{
		Data:   exercises,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}, nil
}

// GetExercise returns an exercise by ID.
func (s *ExerciseService) GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error) {
	ex, err := s.store.GetExercise(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.Exercise{}, NewNotFoundError("exercise not found")
		}
		return queries.Exercise{}, err
	}
	return ex, nil
}

// UpdateExercise updates an exercise after validation and duplicate checks.
func (s *ExerciseService) UpdateExercise(ctx context.Context, id uuid.UUID, name, targetMuscle, notes string) (queries.Exercise, error) {
	name = strings.TrimSpace(name)
	targetMuscle = strings.TrimSpace(targetMuscle)
	notes = strings.TrimSpace(notes)

	if err := validator.ValidateExercise(name, targetMuscle, notes); err != nil {
		if ve, ok := err.(validation.Errors); ok {
			return queries.Exercise{}, validationErrorFromOzzo(ve)
		}
		return queries.Exercise{}, NewValidationError("unknown", err.Error())
	}

	existing, err := s.store.GetExercise(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.Exercise{}, NewNotFoundError("exercise not found")
		}
		return queries.Exercise{}, err
	}

	if existing.Name == name && existing.TargetMuscle.String == targetMuscle && existing.Notes.String == notes {
		return existing, nil
	}

	dup, err := s.store.GetExerciseByNameAndMuscle(ctx, queries.GetExerciseByNameAndMuscleParams{
		Name:         name,
		TargetMuscle: pgtype.Text{String: targetMuscle, Valid: true},
	})
	if err == nil && dup.ID != id {
		return queries.Exercise{}, NewConflictError("exercise with same name and target muscle already exists")
	}
	if err != nil && err != pgx.ErrNoRows {
		return queries.Exercise{}, err
	}

	ex, err := s.store.UpdateExercise(ctx, queries.UpdateExerciseParams{
		ID:           id,
		Name:         name,
		TargetMuscle: pgtype.Text{String: targetMuscle, Valid: true},
		Notes:        pgtype.Text{String: notes, Valid: notes != ""},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.Exercise{}, NewNotFoundError("exercise not found")
		}
		return queries.Exercise{}, err
	}
	return ex, nil
}

// DeleteExercise deletes an exercise if it's not referenced.
func (s *ExerciseService) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	_, err := s.store.GetExercise(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NewNotFoundError("exercise not found")
		}
		return err
	}

	reCount, err := s.store.CountRoutineExercisesByExercise(ctx, id)
	if err != nil {
		return err
	}
	if reCount > 0 {
		return NewReferencedResourceError("exercise is referenced by routines")
	}

	wsCount, err := s.store.CountWorkoutSetsByExercise(ctx, id)
	if err != nil {
		return err
	}
	if wsCount > 0 {
		return NewReferencedResourceError("exercise is referenced by workout sets")
	}

	return s.store.DeleteExercise(ctx, id)
}
