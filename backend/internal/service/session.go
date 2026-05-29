package service

import (
	"context"

	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionStore interface {
	GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	GetActiveWorkoutSession(ctx context.Context) (queries.WorkoutSession, error)
	CreateWorkoutSession(ctx context.Context, routineID pgtype.UUID) (queries.WorkoutSession, error)
	EndWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	GetSessionWithSets(ctx context.Context, id uuid.UUID) (queries.GetSessionWithSetsRow, error)
}

type SessionDetail struct {
	Session queries.GetSessionWithSetsRow
}

type SessionService struct {
	store  SessionStore
	withTx func(ctx context.Context, fn func(store SessionStore) error) error
}

func NewSessionService(db *store.DB) *SessionService {
	return &SessionService{
		store: db.Queries(),
		withTx: func(ctx context.Context, fn func(store SessionStore) error) error {
			return db.WithTx(ctx, func(tx pgx.Tx) error {
				return fn(queries.New(tx))
			})
		},
	}
}

func (s *SessionService) CreateSession(ctx context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error) {
	var routineIDParam pgtype.UUID
	if routineID != nil {
		routineIDParam = pgtype.UUID{Bytes: *routineID, Valid: true}
	}
	return s.store.CreateWorkoutSession(ctx, routineIDParam)
}

func (s *SessionService) GetSession(ctx context.Context, id uuid.UUID) (SessionDetail, error) {
	session, err := s.store.GetSessionWithSets(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SessionDetail{}, NewNotFoundError("session not found")
		}
		return SessionDetail{}, err
	}
	return SessionDetail{Session: session}, nil
}

func (s *SessionService) GetActiveSession(ctx context.Context) (*SessionDetail, error) {
	session, err := s.store.GetActiveWorkoutSession(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	detail, err := s.store.GetSessionWithSets(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	return &SessionDetail{Session: detail}, nil
}

func (s *SessionService) EndSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	session, err := s.store.GetWorkoutSession(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return queries.WorkoutSession{}, NewNotFoundError("session not found")
		}
		return queries.WorkoutSession{}, err
	}
	if session.EndedAt.Valid {
		return queries.WorkoutSession{}, NewConflictError("session already ended")
	}
	return s.store.EndWorkoutSession(ctx, id)
}