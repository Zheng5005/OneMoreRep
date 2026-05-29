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

type mockSessionStore struct {
	getWorkoutSessionFunc       func(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	getActiveWorkoutSessionFunc func(ctx context.Context) (queries.WorkoutSession, error)
	createWorkoutSessionFunc    func(ctx context.Context, routineID pgtype.UUID) (queries.WorkoutSession, error)
	endWorkoutSessionFunc       func(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
	getSessionWithSetsFunc      func(ctx context.Context, id uuid.UUID) (queries.GetSessionWithSetsRow, error)
}

func (m *mockSessionStore) GetWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	return m.getWorkoutSessionFunc(ctx, id)
}
func (m *mockSessionStore) GetActiveWorkoutSession(ctx context.Context) (queries.WorkoutSession, error) {
	return m.getActiveWorkoutSessionFunc(ctx)
}
func (m *mockSessionStore) CreateWorkoutSession(ctx context.Context, routineID pgtype.UUID) (queries.WorkoutSession, error) {
	return m.createWorkoutSessionFunc(ctx, routineID)
}
func (m *mockSessionStore) EndWorkoutSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	return m.endWorkoutSessionFunc(ctx, id)
}
func (m *mockSessionStore) GetSessionWithSets(ctx context.Context, id uuid.UUID) (queries.GetSessionWithSetsRow, error) {
	return m.getSessionWithSetsFunc(ctx, id)
}

var wsTestID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var wsRoutineTestID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

func TestSessionServiceCreateSession(t *testing.T) {
	tests := []struct {
		name      string
		routineID *uuid.UUID
		setupMock func() *mockSessionStore
		wantErr   bool
		wantCode  string
	}{
		{
			name:      "happy path without routine",
			routineID: nil,
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					createWorkoutSessionFunc: func(_ context.Context, routineID pgtype.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							RoutineID: routineID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "happy path with routine",
			routineID: &wsRoutineTestID,
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					createWorkoutSessionFunc: func(_ context.Context, routineID pgtype.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							RoutineID: routineID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := &SessionService{store: mock}

			_, err := svc.CreateSession(context.Background(), tt.routineID)
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

func TestSessionServiceGetSession(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockSessionStore
		wantErr  bool
		wantCode string
	}{
		{
			name: "happy path",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getSessionWithSetsFunc: func(_ context.Context, _ uuid.UUID) (queries.GetSessionWithSetsRow, error) {
						return queries.GetSessionWithSetsRow{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							Sets:      []interface{}{},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getSessionWithSetsFunc: func(_ context.Context, _ uuid.UUID) (queries.GetSessionWithSetsRow, error) {
						return queries.GetSessionWithSetsRow{}, pgx.ErrNoRows
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
			svc := &SessionService{store: mock}

			_, err := svc.GetSession(context.Background(), wsTestID)
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

func TestSessionServiceGetActiveSession(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockSessionStore
		wantNil  bool
		wantErr  bool
	}{
		{
			name: "happy path with active session",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getActiveWorkoutSessionFunc: func(_ context.Context) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
					getSessionWithSetsFunc: func(_ context.Context, _ uuid.UUID) (queries.GetSessionWithSetsRow, error) {
						return queries.GetSessionWithSetsRow{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							Sets:      []interface{}{},
						}, nil
					},
				}
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "no active session",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getActiveWorkoutSessionFunc: func(_ context.Context) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, pgx.ErrNoRows
					},
				}
			},
			wantNil: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := &SessionService{store: mock}

			result, err := svc.GetActiveSession(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("expected non-nil result, got nil")
			}
		})
	}
}

func TestSessionServiceEndSession(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockSessionStore
		wantErr  bool
		wantCode string
	}{
		{
			name: "happy path",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Valid: false},
						}, nil
					},
					endWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "session not found",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
					getWorkoutSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name: "already ended",
			setupMock: func() *mockSessionStore {
				return &mockSessionStore{
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
			svc := &SessionService{store: mock}

			_, err := svc.EndSession(context.Background(), wsTestID)
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