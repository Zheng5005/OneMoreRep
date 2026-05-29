package service

import (
	"context"
	"strings"

	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/Zheng5005/onemorerep/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RoutineStore defines the store operations required by RoutineService.
type RoutineStore interface {
	GetRoutine(ctx context.Context, id uuid.UUID) (queries.Routine, error)
	ListRoutinesPaginated(ctx context.Context, arg queries.ListRoutinesPaginatedParams) ([]queries.Routine, error)
	CountRoutines(ctx context.Context) (int64, error)
	CreateRoutine(ctx context.Context, name string) (queries.Routine, error)
	UpdateRoutine(ctx context.Context, arg queries.UpdateRoutineParams) (queries.Routine, error)
	DeleteRoutine(ctx context.Context, id uuid.UUID) error

	GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	ListRoutineExercises(ctx context.Context, routineID uuid.UUID) ([]queries.ListRoutineExercisesRow, error)
	GetRoutineExercise(ctx context.Context, arg queries.GetRoutineExerciseParams) (queries.RoutineExercise, error)
	GetRoutineExerciseByExercise(ctx context.Context, arg queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error)
	CountRoutineExercises(ctx context.Context, routineID uuid.UUID) (int64, error)
	CreateRoutineExercise(ctx context.Context, arg queries.CreateRoutineExerciseParams) (queries.RoutineExercise, error)
	UpdateRoutineExerciseOrder(ctx context.Context, arg queries.UpdateRoutineExerciseOrderParams) (queries.RoutineExercise, error)
	ShiftRoutineExerciseOrderUp(ctx context.Context, arg queries.ShiftRoutineExerciseOrderUpParams) error
	ShiftRoutineExerciseOrderDown(ctx context.Context, arg queries.ShiftRoutineExerciseOrderDownParams) error
	ReorderRoutineExerciseForward(ctx context.Context, arg queries.ReorderRoutineExerciseForwardParams) error
	ReorderRoutineExerciseBackward(ctx context.Context, arg queries.ReorderRoutineExerciseBackwardParams) error
	DeleteRoutineExercise(ctx context.Context, id uuid.UUID) error
}

// RoutineListResult wraps the list response.
type RoutineListResult struct {
	Data   []queries.Routine
	Limit  int32
	Offset int32
	Total  int64
}

// RoutineDetail wraps a routine with its exercises.
type RoutineDetail struct {
	Routine   queries.Routine
	Exercises []queries.ListRoutineExercisesRow
}

// RoutineService provides business logic for routines.
type RoutineService struct {
	store  RoutineStore
	withTx func(ctx context.Context, fn func(store RoutineStore) error) error
}

// NewRoutineService creates a new RoutineService backed by db.
func NewRoutineService(db *store.DB) *RoutineService {
	return &RoutineService{
		store: db.Queries(),
		withTx: func(ctx context.Context, fn func(store RoutineStore) error) error {
			return db.WithTx(ctx, func(tx pgx.Tx) error {
				return fn(queries.New(tx))
			})
		},
	}
}

// CreateRoutine creates a new routine after validation.
func (s *RoutineService) CreateRoutine(ctx context.Context, name string) (queries.Routine, error) {
	name = strings.TrimSpace(name)
	if err := validator.ValidateRoutineName(name); err != nil {
		return queries.Routine{}, NewValidationError("name", err.Error())
	}
	return s.store.CreateRoutine(ctx, name)
}

// ListRoutines returns a paginated list of routines.
func (s *RoutineService) ListRoutines(ctx context.Context, limit, offset int32) (RoutineListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	routines, err := s.store.ListRoutinesPaginated(ctx, queries.ListRoutinesPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return RoutineListResult{}, err
	}

	total, err := s.store.CountRoutines(ctx)
	if err != nil {
		return RoutineListResult{}, err
	}

	return RoutineListResult{
		Data:   routines,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}, nil
}

// GetRoutine returns a routine by ID with its exercises.
func (s *RoutineService) GetRoutine(ctx context.Context, id uuid.UUID) (RoutineDetail, error) {
	routine, err := s.store.GetRoutine(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return RoutineDetail{}, NewNotFoundError("routine not found")
		}
		return RoutineDetail{}, err
	}

	exercises, err := s.store.ListRoutineExercises(ctx, id)
	if err != nil {
		return RoutineDetail{}, err
	}

	return RoutineDetail{
		Routine:   routine,
		Exercises: exercises,
	}, nil
}

// UpdateRoutine updates a routine name after validation.
func (s *RoutineService) UpdateRoutine(ctx context.Context, id uuid.UUID, name string) (queries.Routine, error) {
	name = strings.TrimSpace(name)
	if err := validator.ValidateRoutineName(name); err != nil {
		return queries.Routine{}, NewValidationError("name", err.Error())
	}

	existing, err := s.store.GetRoutine(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.Routine{}, NewNotFoundError("routine not found")
		}
		return queries.Routine{}, err
	}

	if existing.Name == name {
		return existing, nil
	}

	updated, err := s.store.UpdateRoutine(ctx, queries.UpdateRoutineParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.Routine{}, NewNotFoundError("routine not found")
		}
		return queries.Routine{}, err
	}
	return updated, nil
}

// DeleteRoutine deletes a routine by ID.
func (s *RoutineService) DeleteRoutine(ctx context.Context, id uuid.UUID) error {
	_, err := s.store.GetRoutine(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NewNotFoundError("routine not found")
		}
		return err
	}

	return s.store.DeleteRoutine(ctx, id)
}

// AddRoutineExercise adds an exercise to a routine with optional order.
func (s *RoutineService) AddRoutineExercise(ctx context.Context, routineID uuid.UUID, exerciseIDStr string, order *int32) (queries.RoutineExercise, error) {
	exerciseID, err := uuid.Parse(exerciseIDStr)
	if err != nil {
		return queries.RoutineExercise{}, NewValidationError("exercise_id", "invalid UUID")
	}

	_, err = s.store.GetRoutine(ctx, routineID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.RoutineExercise{}, NewNotFoundError("routine not found")
		}
		return queries.RoutineExercise{}, err
	}

	_, err = s.store.GetExercise(ctx, exerciseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.RoutineExercise{}, NewValidationError("exercise_id", "exercise not found")
		}
		return queries.RoutineExercise{}, err
	}

	_, err = s.store.GetRoutineExerciseByExercise(ctx, queries.GetRoutineExerciseByExerciseParams{
		RoutineID:  routineID,
		ExerciseID: exerciseID,
	})
	if err == nil {
		return queries.RoutineExercise{}, NewConflictError("exercise already in routine")
	}
	if err != pgx.ErrNoRows {
		return queries.RoutineExercise{}, err
	}

	count, err := s.store.CountRoutineExercises(ctx, routineID)
	if err != nil {
		return queries.RoutineExercise{}, err
	}
	maxOrder := int32(count)

	var targetOrder int32
	if order != nil {
		if *order < 1 || *order > maxOrder+1 {
			return queries.RoutineExercise{}, NewValidationError("order", "order must be between 1 and max+1")
		}
		targetOrder = *order
	} else {
		targetOrder = maxOrder + 1
	}

	var result queries.RoutineExercise
	err = s.withTx(ctx, func(txStore RoutineStore) error {
		if targetOrder <= maxOrder {
			if err := txStore.ShiftRoutineExerciseOrderUp(ctx, queries.ShiftRoutineExerciseOrderUpParams{
				RoutineID: routineID,
				Order:     targetOrder,
			}); err != nil {
				return err
			}
		}
		re, err := txStore.CreateRoutineExercise(ctx, queries.CreateRoutineExerciseParams{
			RoutineID:  routineID,
			ExerciseID: exerciseID,
			Order:      targetOrder,
		})
		result = re
		return err
	})
	return result, err
}

// UpdateRoutineExerciseOrder reorders an exercise within a routine.
func (s *RoutineService) UpdateRoutineExerciseOrder(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID, newOrder int32) (queries.RoutineExercise, error) {
	re, err := s.store.GetRoutineExercise(ctx, queries.GetRoutineExerciseParams{
		ID:        routineExerciseID,
		RoutineID: routineID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.RoutineExercise{}, NewNotFoundError("routine exercise not found")
		}
		return queries.RoutineExercise{}, err
	}

	count, err := s.store.CountRoutineExercises(ctx, routineID)
	if err != nil {
		return queries.RoutineExercise{}, err
	}
	maxOrder := int32(count)

	if newOrder < 1 || newOrder > maxOrder {
		return queries.RoutineExercise{}, NewValidationError("order", "order must be between 1 and max")
	}

	if re.Order == newOrder {
		return re, nil
	}

	var result queries.RoutineExercise
	err = s.withTx(ctx, func(txStore RoutineStore) error {
		if newOrder > re.Order {
			if err := txStore.ReorderRoutineExerciseForward(ctx, queries.ReorderRoutineExerciseForwardParams{
				RoutineID: routineID,
				Order:     re.Order,
				Order_2:   newOrder,
			}); err != nil {
				return err
			}
		} else {
			if err := txStore.ReorderRoutineExerciseBackward(ctx, queries.ReorderRoutineExerciseBackwardParams{
				RoutineID: routineID,
				Order:     newOrder,
				Order_2:   re.Order,
			}); err != nil {
				return err
			}
		}
		updated, err := txStore.UpdateRoutineExerciseOrder(ctx, queries.UpdateRoutineExerciseOrderParams{
			ID:    routineExerciseID,
			Order: newOrder,
		})
		result = updated
		return err
	})
	return result, err
}

// DeleteRoutineExercise removes an exercise from a routine and closes the gap.
func (s *RoutineService) DeleteRoutineExercise(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID) error {
	re, err := s.store.GetRoutineExercise(ctx, queries.GetRoutineExerciseParams{
		ID:        routineExerciseID,
		RoutineID: routineID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NewNotFoundError("routine exercise not found")
		}
		return err
	}

	return s.withTx(ctx, func(txStore RoutineStore) error {
		if err := txStore.DeleteRoutineExercise(ctx, routineExerciseID); err != nil {
			return err
		}
		return txStore.ShiftRoutineExerciseOrderDown(ctx, queries.ShiftRoutineExerciseOrderDownParams{
			RoutineID: routineID,
			Order:     re.Order,
		})
	})
}
