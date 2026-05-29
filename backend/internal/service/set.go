package service

import (
	"context"
	"math/big"

	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/Zheng5005/onemorerep/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SetStore interface {
	GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	GetWorkoutSet(ctx context.Context, id uuid.UUID) (queries.WorkoutSet, error)
	CreateWorkoutSet(ctx context.Context, arg queries.CreateWorkoutSetParams) (queries.WorkoutSet, error)
	UpdateWorkoutSet(ctx context.Context, arg queries.UpdateWorkoutSetParams) (queries.WorkoutSet, error)
	DeleteWorkoutSet(ctx context.Context, id uuid.UUID) error
	GetMaxSetNumber(ctx context.Context, arg queries.GetMaxSetNumberParams) (interface{}, error)
	RenumberWorkoutSets(ctx context.Context, arg queries.RenumberWorkoutSetsParams) error
}

type SetService struct {
	store  SetStore
	withTx func(ctx context.Context, fn func(store SetStore) error) error
}

func NewSetService(db *store.DB) *SetService {
	return &SetService{
		store: db.Queries(),
		withTx: func(ctx context.Context, fn func(store SetStore) error) error {
			return db.WithTx(ctx, func(tx pgx.Tx) error {
				return fn(queries.New(tx))
			})
		},
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	bf := big.NewFloat(f)
	ai, exp := bf.Int(nil)
	return pgtype.Numeric{Int: ai, Exp: int32(exp), Valid: true}
}

func setValidationErrorFromOzzo(ve validation.Errors) *AppError {
	if err, ok := ve["weight"]; ok {
		return NewValidationError("weight", err.Error())
	}
	if err, ok := ve["reps"]; ok {
		return NewValidationError("reps", err.Error())
	}
	for field, err := range ve {
		return NewValidationError(field, err.Error())
	}
	return NewValidationError("unknown", "validation failed")
}

func (s *SetService) CreateSet(ctx context.Context, sessionID uuid.UUID, exerciseID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error) {
	session, err := s.store.GetWorkoutSession(ctx, sessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.WorkoutSet{}, NewNotFoundError("session not found")
		}
		return queries.WorkoutSet{}, err
	}
	if session.EndedAt.Valid {
		return queries.WorkoutSet{}, NewConflictError("session already ended")
	}

	if err := validator.ValidateWorkoutSet(weight, reps); err != nil {
		if ve, ok := err.(validation.Errors); ok {
			return queries.WorkoutSet{}, setValidationErrorFromOzzo(ve)
		}
		return queries.WorkoutSet{}, NewValidationError("unknown", err.Error())
	}

	maxSet, err := s.store.GetMaxSetNumber(ctx, queries.GetMaxSetNumberParams{
		SessionID:  sessionID,
		ExerciseID: exerciseID,
	})
	if err != nil {
		return queries.WorkoutSet{}, err
	}
	var maxSetNumber int32
	switch v := maxSet.(type) {
	case int32:
		maxSetNumber = v
	case int64:
		maxSetNumber = int32(v)
	case float64:
		maxSetNumber = int32(v)
	}

	return s.store.CreateWorkoutSet(ctx, queries.CreateWorkoutSetParams{
		SessionID:  sessionID,
		ExerciseID: exerciseID,
		SetNumber:  maxSetNumber + 1,
		Weight:     floatToNumeric(weight),
		Reps:       int32(reps),
	})
}

func (s *SetService) UpdateSet(ctx context.Context, setID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error) {
	set, err := s.store.GetWorkoutSet(ctx, setID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.WorkoutSet{}, NewNotFoundError("set not found")
		}
		return queries.WorkoutSet{}, err
	}

	session, err := s.store.GetWorkoutSession(ctx, set.SessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.WorkoutSet{}, NewNotFoundError("session not found")
		}
		return queries.WorkoutSet{}, err
	}
	if session.EndedAt.Valid {
		return queries.WorkoutSet{}, NewConflictError("session already ended")
	}

	if err := validator.ValidateWorkoutSet(weight, reps); err != nil {
		if ve, ok := err.(validation.Errors); ok {
			return queries.WorkoutSet{}, setValidationErrorFromOzzo(ve)
		}
		return queries.WorkoutSet{}, NewValidationError("unknown", err.Error())
	}

	return s.store.UpdateWorkoutSet(ctx, queries.UpdateWorkoutSetParams{
		ID:     setID,
		Weight: floatToNumeric(weight),
		Reps:   int32(reps),
	})
}

func (s *SetService) DeleteSet(ctx context.Context, setID uuid.UUID) error {
	set, err := s.store.GetWorkoutSet(ctx, setID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NewNotFoundError("set not found")
		}
		return err
	}

	return s.withTx(ctx, func(txStore SetStore) error {
		if err := txStore.DeleteWorkoutSet(ctx, setID); err != nil {
			return err
		}
		return txStore.RenumberWorkoutSets(ctx, queries.RenumberWorkoutSetsParams{
			SessionID:  set.SessionID,
			ExerciseID: set.ExerciseID,
			SetNumber:  set.SetNumber,
		})
	})
}

var _ = big.Int{}