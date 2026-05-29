package service

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockProgressStore struct {
	getExerciseLastValuesFunc      func(ctx context.Context, exerciseID uuid.UUID) (queries.GetExerciseLastValuesRow, error)
	getSessionSummaryFunc         func(ctx context.Context, id uuid.UUID) (queries.GetSessionSummaryRow, error)
	getSessionExerciseBreakdownFunc func(ctx context.Context, sessionID uuid.UUID) ([]queries.GetSessionExerciseBreakdownRow, error)
	getWorkoutSessionFunc         func(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
}

func (m *mockProgressStore) GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (queries.GetExerciseLastValuesRow, error) {
	return m.getExerciseLastValuesFunc(ctx, exerciseID)
}

func (m *mockProgressStore) GetSessionSummary(ctx context.Context, id uuid.UUID) (queries.GetSessionSummaryRow, error) {
	return m.getSessionSummaryFunc(ctx, id)
}

func (m *mockProgressStore) GetSessionExerciseBreakdown(ctx context.Context, sessionID uuid.UUID) ([]queries.GetSessionExerciseBreakdownRow, error) {
	return m.getSessionExerciseBreakdownFunc(ctx, sessionID)
}

func (m *mockProgressStore) GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	return m.getWorkoutSessionFunc(ctx, id)
}

func TestProgressServiceGetExerciseLastValues(t *testing.T) {
	tests := []struct {
		name       string
		exerciseID uuid.UUID
		setupMock  func() *mockProgressStore
		wantNil    bool
		wantWeight *float64
		wantReps   *float64
	}{
		{
			name:       "returns values when history exists",
			exerciseID: testID,
			setupMock: func() *mockProgressStore {
				return &mockProgressStore{
					getExerciseLastValuesFunc: func(_ context.Context, _ uuid.UUID) (queries.GetExerciseLastValuesRow, error) {
						w := pgtype.Numeric{Int: big.NewInt(135), Exp: 0, Valid: true}
						return queries.GetExerciseLastValuesRow{Weight: w, Reps: 10}, nil
					},
				}
			},
			wantNil:    false,
			wantWeight: func() *float64 { v := 135.0; return &v }(),
			wantReps:   func() *float64 { v := 10.0; return &v }(),
		},
		{
			name:       "returns null when no history",
			exerciseID: testID,
			setupMock: func() *mockProgressStore {
				return &mockProgressStore{
					getExerciseLastValuesFunc: func(_ context.Context, _ uuid.UUID) (queries.GetExerciseLastValuesRow, error) {
						return queries.GetExerciseLastValuesRow{}, pgx.ErrNoRows
					},
				}
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewProgressService(mock)

			result, err := svc.GetExerciseLastValues(context.Background(), tt.exerciseID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil result, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected non-nil result")
			}

			if tt.wantWeight != nil && result.Weight == nil {
				t.Errorf("expected weight %v, got nil", *tt.wantWeight)
			} else if tt.wantWeight != nil && result.Weight != nil && *result.Weight != *tt.wantWeight {
				t.Errorf("weight = %v, want %v", *result.Weight, *tt.wantWeight)
			}

			if tt.wantReps != nil && result.Reps == nil {
				t.Errorf("expected reps %v, got nil", *tt.wantReps)
			} else if tt.wantReps != nil && result.Reps != nil && *result.Reps != *tt.wantReps {
				t.Errorf("reps = %v, want %v", *result.Reps, *tt.wantReps)
			}
		})
	}
}

func TestProgressServiceGetSessionSummary(t *testing.T) {
	sessionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	exerciseID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	tests := []struct {
		name      string
		sessionID uuid.UUID
		setupMock func() *mockProgressStore
		wantErr   bool
		wantCode  string
	}{
		{
			name:      "ended session returns correct summary",
			sessionID: sessionID,
			setupMock: func() *mockProgressStore {
				return &mockProgressStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        sessionID,
							StartedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Time: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Valid: true},
						}, nil
					},
					getSessionSummaryFunc: func(_ context.Context, _ uuid.UUID) (queries.GetSessionSummaryRow, error) {
						return queries.GetSessionSummaryRow{
							SessionID:     sessionID,
							StartedAt:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							EndedAt:       pgtype.Timestamptz{Time: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Valid: true},
							ExerciseCount: 2,
							TotalSets:     6,
							TotalVolume:   float64(4050),
						}, nil
					},
					getSessionExerciseBreakdownFunc: func(_ context.Context, _ uuid.UUID) ([]queries.GetSessionExerciseBreakdownRow, error) {
						return []queries.GetSessionExerciseBreakdownRow{
							{ExerciseID: exerciseID, ExerciseName: "Bench Press", SetsCount: 3, BestVolume: float64(1620), BestWeight: float64(135), BestReps: float64(12)},
							{ExerciseID: uuid.MustParse("55555555-5555-5555-5555-555555555555"), ExerciseName: "Squat", SetsCount: 3, BestVolume: float64(2430), BestWeight: float64(185), BestReps: float64(10)},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "in-progress session returns zero duration",
			sessionID: sessionID,
			setupMock: func() *mockProgressStore {
				return &mockProgressStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        sessionID,
							StartedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
					getSessionSummaryFunc: func(_ context.Context, _ uuid.UUID) (queries.GetSessionSummaryRow, error) {
						return queries.GetSessionSummaryRow{
							SessionID:     sessionID,
							StartedAt:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
							EndedAt:       pgtype.Timestamptz{Valid: false},
							ExerciseCount: 1,
							TotalSets:     3,
							TotalVolume:   float64(810),
						}, nil
					},
					getSessionExerciseBreakdownFunc: func(_ context.Context, _ uuid.UUID) ([]queries.GetSessionExerciseBreakdownRow, error) {
						return []queries.GetSessionExerciseBreakdownRow{
							{ExerciseID: exerciseID, ExerciseName: "Bench Press", SetsCount: 3, BestVolume: float64(810), BestWeight: float64(60), BestReps: float64(10)},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "session not found",
			sessionID: sessionID,
			setupMock: func() *mockProgressStore {
				return &mockProgressStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewProgressService(mock)

			result, err := svc.GetSessionSummary(context.Background(), tt.sessionID)
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
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatalf("expected non-nil result")
			}
			if result.SessionID != tt.sessionID {
				t.Errorf("session_id = %v, want %v", result.SessionID, tt.sessionID)
			}
		})
	}
}