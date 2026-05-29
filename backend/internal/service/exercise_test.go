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

type mockExerciseStore struct {
	createFunc             func(ctx context.Context, arg queries.CreateExerciseParams) (queries.Exercise, error)
	searchFunc             func(ctx context.Context, arg queries.SearchExercisesParams) ([]queries.Exercise, error)
	countFunc              func(ctx context.Context, search string) (int64, error)
	getFunc                func(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	getByNameAndMuscleFunc func(ctx context.Context, arg queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error)
	updateFunc             func(ctx context.Context, arg queries.UpdateExerciseParams) (queries.Exercise, error)
	deleteFunc             func(ctx context.Context, id uuid.UUID) error
	countREFunc            func(ctx context.Context, exerciseID uuid.UUID) (int64, error)
	countWSFunc            func(ctx context.Context, exerciseID uuid.UUID) (int64, error)
}

func (m *mockExerciseStore) CreateExercise(ctx context.Context, arg queries.CreateExerciseParams) (queries.Exercise, error) {
	return m.createFunc(ctx, arg)
}

func (m *mockExerciseStore) SearchExercises(ctx context.Context, arg queries.SearchExercisesParams) ([]queries.Exercise, error) {
	return m.searchFunc(ctx, arg)
}

func (m *mockExerciseStore) CountExercises(ctx context.Context, search string) (int64, error) {
	return m.countFunc(ctx, search)
}

func (m *mockExerciseStore) GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error) {
	return m.getFunc(ctx, id)
}

func (m *mockExerciseStore) GetExerciseByNameAndMuscle(ctx context.Context, arg queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error) {
	return m.getByNameAndMuscleFunc(ctx, arg)
}

func (m *mockExerciseStore) UpdateExercise(ctx context.Context, arg queries.UpdateExerciseParams) (queries.Exercise, error) {
	return m.updateFunc(ctx, arg)
}

func (m *mockExerciseStore) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func (m *mockExerciseStore) CountRoutineExercisesByExercise(ctx context.Context, exerciseID uuid.UUID) (int64, error) {
	return m.countREFunc(ctx, exerciseID)
}

func (m *mockExerciseStore) CountWorkoutSetsByExercise(ctx context.Context, exerciseID uuid.UUID) (int64, error) {
	return m.countWSFunc(ctx, exerciseID)
}

var testID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var otherID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

func TestExerciseServiceCreate(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputMuscle string
		inputNotes  string
		setupMock   func() *mockExerciseStore
		wantErr     bool
		wantCode    string
		wantField   string
	}{
		{
			name:        "happy path",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "Keep back flat",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getByNameAndMuscleFunc: func(_ context.Context, _ queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error) {
						return queries.Exercise{}, pgx.ErrNoRows
					},
					createFunc: func(_ context.Context, arg queries.CreateExerciseParams) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         arg.Name,
							TargetMuscle: arg.TargetMuscle,
							Notes:        arg.Notes,
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:        "blank name after trim",
			inputName:   "   ",
			inputMuscle: "Chest",
			inputNotes:  "",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "name",
		},
		{
			name:        "duplicate name and muscle",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getByNameAndMuscleFunc: func(_ context.Context, _ queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error) {
						return queries.Exercise{ID: otherID, Name: "Bench Press"}, nil
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
			svc := NewExerciseService(mock)

			_, err := svc.CreateExercise(context.Background(), tt.inputName, tt.inputMuscle, tt.inputNotes)
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
				if tt.wantField != "" && appErr.Field != tt.wantField {
					t.Errorf("field = %q, want %q", appErr.Field, tt.wantField)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExerciseServiceList(t *testing.T) {
	mock := &mockExerciseStore{
		searchFunc: func(_ context.Context, arg queries.SearchExercisesParams) ([]queries.Exercise, error) {
			if arg.Limit != 20 || arg.Offset != 0 {
				t.Errorf("expected limit=20 offset=0, got limit=%d offset=%d", arg.Limit, arg.Offset)
			}
			return []queries.Exercise{
				{ID: testID, Name: "Bench Press", TargetMuscle: pgtype.Text{String: "Chest", Valid: true}},
			}, nil
		},
		countFunc: func(_ context.Context, search string) (int64, error) {
			if search != "bench" {
				t.Errorf("expected search=bench, got %q", search)
			}
			return 1, nil
		},
	}

	svc := NewExerciseService(mock)
	result, err := svc.ListExercises(context.Background(), 20, 0, "bench")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(result.Data))
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want %d", result.Total, 1)
	}
	if result.Data[0].Name != "Bench Press" {
		t.Errorf("name = %q, want %q", result.Data[0].Name, "Bench Press")
	}
}

func TestExerciseServiceGet(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockExerciseStore
		wantErr  bool
		wantCode string
	}{
		{
			name: "happy path",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{}, pgx.ErrNoRows
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
			svc := NewExerciseService(mock)

			_, err := svc.GetExercise(context.Background(), testID)
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

func TestExerciseServiceUpdate(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputMuscle string
		inputNotes  string
		setupMock   func() *mockExerciseStore
		wantErr     bool
		wantCode    string
		wantField   string
	}{
		{
			name:        "happy path",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "Updated notes",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							Notes:        pgtype.Text{String: "Old notes", Valid: true},
						}, nil
					},
					getByNameAndMuscleFunc: func(_ context.Context, _ queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error) {
						return queries.Exercise{ID: testID}, nil
					},
					updateFunc: func(_ context.Context, arg queries.UpdateExerciseParams) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           arg.ID,
							Name:         arg.Name,
							TargetMuscle: arg.TargetMuscle,
							Notes:        arg.Notes,
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:        "idempotent same values",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "Same notes",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							Notes:        pgtype.Text{String: "Same notes", Valid: true},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:        "not found",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:        "duplicate conflict with different id",
			inputName:   "Bench Press",
			inputMuscle: "Chest",
			inputNotes:  "",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							Notes:        pgtype.Text{String: "Old notes", Valid: true},
						}, nil
					},
					getByNameAndMuscleFunc: func(_ context.Context, _ queries.GetExerciseByNameAndMuscleParams) (queries.Exercise, error) {
						return queries.Exercise{ID: otherID, Name: "Bench Press"}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "CONFLICT",
		},
		{
			name:        "validation error",
			inputName:   "",
			inputMuscle: "Chest",
			inputNotes:  "",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewExerciseService(mock)

			_, err := svc.UpdateExercise(context.Background(), testID, tt.inputName, tt.inputMuscle, tt.inputNotes)
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
				if tt.wantField != "" && appErr.Field != tt.wantField {
					t.Errorf("field = %q, want %q", appErr.Field, tt.wantField)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExerciseServiceDelete(t *testing.T) {
	tests := []struct {
		name     string
		setupMock func() *mockExerciseStore
		wantErr  bool
		wantCode string
	}{
		{
			name: "happy path",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: testID}, nil
					},
					countREFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 0, nil
					},
					countWSFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 0, nil
					},
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name: "referenced by routine",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: testID}, nil
					},
					countREFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 2, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "REFERENCED_RESOURCE",
		},
		{
			name: "referenced by workout sets",
			setupMock: func() *mockExerciseStore {
				return &mockExerciseStore{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: testID}, nil
					},
					countREFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 0, nil
					},
					countWSFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 3, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "REFERENCED_RESOURCE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := NewExerciseService(mock)

			err := svc.DeleteExercise(context.Background(), testID)
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
