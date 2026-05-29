package service

import (
	"context"
	"testing"
	"time"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockSetStore struct {
	getWorkoutSessionFunc   func(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	getWorkoutSetFunc       func(ctx context.Context, id uuid.UUID) (queries.WorkoutSet, error)
	createWorkoutSetFunc    func(ctx context.Context, arg queries.CreateWorkoutSetParams) (queries.WorkoutSet, error)
	updateWorkoutSetFunc    func(ctx context.Context, arg queries.UpdateWorkoutSetParams) (queries.WorkoutSet, error)
	deleteWorkoutSetFunc    func(ctx context.Context, id uuid.UUID) error
	getMaxSetNumberFunc     func(ctx context.Context, arg queries.GetMaxSetNumberParams) (interface{}, error)
	renumberWorkoutSetsFunc func(ctx context.Context, arg queries.RenumberWorkoutSetsParams) error
}

func (m *mockSetStore) GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	return m.getWorkoutSessionFunc(ctx, id)
}
func (m *mockSetStore) GetWorkoutSet(ctx context.Context, id uuid.UUID) (queries.WorkoutSet, error) {
	return m.getWorkoutSetFunc(ctx, id)
}
func (m *mockSetStore) CreateWorkoutSet(ctx context.Context, arg queries.CreateWorkoutSetParams) (queries.WorkoutSet, error) {
	return m.createWorkoutSetFunc(ctx, arg)
}
func (m *mockSetStore) UpdateWorkoutSet(ctx context.Context, arg queries.UpdateWorkoutSetParams) (queries.WorkoutSet, error) {
	return m.updateWorkoutSetFunc(ctx, arg)
}
func (m *mockSetStore) DeleteWorkoutSet(ctx context.Context, id uuid.UUID) error {
	return m.deleteWorkoutSetFunc(ctx, id)
}
func (m *mockSetStore) GetMaxSetNumber(ctx context.Context, arg queries.GetMaxSetNumberParams) (interface{}, error) {
	return m.getMaxSetNumberFunc(ctx, arg)
}
func (m *mockSetStore) RenumberWorkoutSets(ctx context.Context, arg queries.RenumberWorkoutSetsParams) error {
	return m.renumberWorkoutSetsFunc(ctx, arg)
}

var wsSetTestID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
var wsExerciseTestID = uuid.MustParse("44444444-4444-4444-4444-444444444444")

func TestSetServiceCreateSet(t *testing.T) {
	tests := []struct {
		name       string
		sessionID  uuid.UUID
		exerciseID uuid.UUID
		weight     float64
		reps       int
		setupMock  func() *mockSetStore
		wantErr    bool
		wantCode   string
	}{
		{
			name:       "happy path",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     135.0,
			reps:       10,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
					getMaxSetNumberFunc: func(_ context.Context, _ queries.GetMaxSetNumberParams) (interface{}, error) {
						return int64(0), nil
					},
					createWorkoutSetFunc: func(_ context.Context, arg queries.CreateWorkoutSetParams) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:         wsSetTestID,
							SessionID:  arg.SessionID,
							ExerciseID: arg.ExerciseID,
							SetNumber:  arg.SetNumber,
							Weight:     arg.Weight,
							Reps:       arg.Reps,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:       "auto set number increments",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     135.0,
			reps:       10,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
					getMaxSetNumberFunc: func(_ context.Context, _ queries.GetMaxSetNumberParams) (interface{}, error) {
						return int64(2), nil
					},
					createWorkoutSetFunc: func(_ context.Context, arg queries.CreateWorkoutSetParams) (queries.WorkoutSet, error) {
						if arg.SetNumber != 3 {
							t.Errorf("expected SetNumber=3, got %d", arg.SetNumber)
						}
						return queries.WorkoutSet{
							ID:         wsSetTestID,
							SessionID:  arg.SessionID,
							ExerciseID: arg.ExerciseID,
							SetNumber:  arg.SetNumber,
							Weight:     arg.Weight,
							Reps:       arg.Reps,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:       "session not found",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     135.0,
			reps:       10,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:       "session ended",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     135.0,
			reps:       10,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "CONFLICT",
		},
		{
			name:       "negative weight validation error",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     -10.0,
			reps:       10,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:       "zero reps validation error",
			sessionID:  wsTestID,
			exerciseID: wsExerciseTestID,
			weight:     135.0,
			reps:       0,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := &SetService{store: mock}

			_, err := svc.CreateSet(context.Background(), tt.sessionID, tt.exerciseID, tt.weight, tt.reps)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				appErr, ok := err.(*AppError)
				if !ok {
					t.Fatalf("expected *AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", appErr.Code, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetServiceUpdateSet(t *testing.T) {
	tests := []struct {
		name     string
		setID    uuid.UUID
		weight   float64
		reps     int
		setupMock func() *mockSetStore
		wantErr  bool
		wantCode string
	}{
		{
			name:   "happy path",
			setID:  wsSetTestID,
			weight: 155.0,
			reps:   8,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSetFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:         wsSetTestID,
							SessionID:  wsTestID,
							ExerciseID: wsExerciseTestID,
							SetNumber:  1,
							Weight:    floatToNumeric(135.0),
							Reps:      10,
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
					updateWorkoutSetFunc: func(_ context.Context, arg queries.UpdateWorkoutSetParams) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:        wsSetTestID,
							SessionID: wsTestID,
							Weight:    arg.Weight,
							Reps:      arg.Reps,
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "set not found",
			setID:  wsSetTestID,
			weight: 155.0,
			reps:   8,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSetFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:   "session ended",
			setID:  wsSetTestID,
			weight: 155.0,
			reps:   8,
			setupMock: func() *mockSetStore {
				return &mockSetStore{
					getWorkoutSetFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:         wsSetTestID,
							SessionID:  wsTestID,
							ExerciseID: wsExerciseTestID,
							SetNumber:  1,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "CONFLICT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := &SetService{store: mock}

			_, err := svc.UpdateSet(context.Background(), tt.setID, tt.weight, tt.reps)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				appErr, ok := err.(*AppError)
				if !ok {
					t.Fatalf("expected *AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", appErr.Code, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetServiceDeleteSet(t *testing.T) {
	t.Run("set not found", func(t *testing.T) {
		mock := &mockSetStore{
			getWorkoutSetFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSet, error) {
				return queries.WorkoutSet{}, pgx.ErrNoRows
			},
		}
		svc := &SetService{store: mock}

		err := svc.DeleteSet(context.Background(), wsSetTestID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		appErr, ok := err.(*AppError)
		if !ok {
			t.Fatalf("expected *AppError, got %T", err)
		}
		if appErr.Code != "NOT_FOUND" {
			t.Errorf("code = %q, want %q", appErr.Code, "NOT_FOUND")
		}
	})
}