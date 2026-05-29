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

type mockRoutineStore struct {
	getRoutineFunc                func(ctx context.Context, id uuid.UUID) (queries.Routine, error)
	listRoutinesPaginatedFunc     func(ctx context.Context, arg queries.ListRoutinesPaginatedParams) ([]queries.Routine, error)
	countRoutinesFunc             func(ctx context.Context) (int64, error)
	createRoutineFunc             func(ctx context.Context, name string) (queries.Routine, error)
	updateRoutineFunc             func(ctx context.Context, arg queries.UpdateRoutineParams) (queries.Routine, error)
	deleteRoutineFunc             func(ctx context.Context, id uuid.UUID) error
	getExerciseFunc               func(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	listRoutineExercisesFunc      func(ctx context.Context, routineID uuid.UUID) ([]queries.ListRoutineExercisesRow, error)
	getRoutineExerciseFunc        func(ctx context.Context, arg queries.GetRoutineExerciseParams) (queries.RoutineExercise, error)
	getRoutineExerciseByExerciseFunc func(ctx context.Context, arg queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error)
	countRoutineExercisesFunc     func(ctx context.Context, routineID uuid.UUID) (int64, error)
	createRoutineExerciseFunc     func(ctx context.Context, arg queries.CreateRoutineExerciseParams) (queries.RoutineExercise, error)
	updateRoutineExerciseOrderFunc func(ctx context.Context, arg queries.UpdateRoutineExerciseOrderParams) (queries.RoutineExercise, error)
	shiftRoutineExerciseOrderUpFunc   func(ctx context.Context, arg queries.ShiftRoutineExerciseOrderUpParams) error
	shiftRoutineExerciseOrderDownFunc func(ctx context.Context, arg queries.ShiftRoutineExerciseOrderDownParams) error
	reorderRoutineExerciseForwardFunc func(ctx context.Context, arg queries.ReorderRoutineExerciseForwardParams) error
	reorderRoutineExerciseBackwardFunc func(ctx context.Context, arg queries.ReorderRoutineExerciseBackwardParams) error
	deleteRoutineExerciseFunc     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRoutineStore) GetRoutine(ctx context.Context, id uuid.UUID) (queries.Routine, error) {
	return m.getRoutineFunc(ctx, id)
}
func (m *mockRoutineStore) ListRoutinesPaginated(ctx context.Context, arg queries.ListRoutinesPaginatedParams) ([]queries.Routine, error) {
	return m.listRoutinesPaginatedFunc(ctx, arg)
}
func (m *mockRoutineStore) CountRoutines(ctx context.Context) (int64, error) {
	return m.countRoutinesFunc(ctx)
}
func (m *mockRoutineStore) CreateRoutine(ctx context.Context, name string) (queries.Routine, error) {
	return m.createRoutineFunc(ctx, name)
}
func (m *mockRoutineStore) UpdateRoutine(ctx context.Context, arg queries.UpdateRoutineParams) (queries.Routine, error) {
	return m.updateRoutineFunc(ctx, arg)
}
func (m *mockRoutineStore) DeleteRoutine(ctx context.Context, id uuid.UUID) error {
	return m.deleteRoutineFunc(ctx, id)
}
func (m *mockRoutineStore) GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error) {
	return m.getExerciseFunc(ctx, id)
}
func (m *mockRoutineStore) ListRoutineExercises(ctx context.Context, routineID uuid.UUID) ([]queries.ListRoutineExercisesRow, error) {
	return m.listRoutineExercisesFunc(ctx, routineID)
}
func (m *mockRoutineStore) GetRoutineExercise(ctx context.Context, arg queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
	return m.getRoutineExerciseFunc(ctx, arg)
}
func (m *mockRoutineStore) GetRoutineExerciseByExercise(ctx context.Context, arg queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error) {
	return m.getRoutineExerciseByExerciseFunc(ctx, arg)
}
func (m *mockRoutineStore) CountRoutineExercises(ctx context.Context, routineID uuid.UUID) (int64, error) {
	return m.countRoutineExercisesFunc(ctx, routineID)
}
func (m *mockRoutineStore) CreateRoutineExercise(ctx context.Context, arg queries.CreateRoutineExerciseParams) (queries.RoutineExercise, error) {
	return m.createRoutineExerciseFunc(ctx, arg)
}
func (m *mockRoutineStore) UpdateRoutineExerciseOrder(ctx context.Context, arg queries.UpdateRoutineExerciseOrderParams) (queries.RoutineExercise, error) {
	return m.updateRoutineExerciseOrderFunc(ctx, arg)
}
func (m *mockRoutineStore) ShiftRoutineExerciseOrderUp(ctx context.Context, arg queries.ShiftRoutineExerciseOrderUpParams) error {
	return m.shiftRoutineExerciseOrderUpFunc(ctx, arg)
}
func (m *mockRoutineStore) ShiftRoutineExerciseOrderDown(ctx context.Context, arg queries.ShiftRoutineExerciseOrderDownParams) error {
	return m.shiftRoutineExerciseOrderDownFunc(ctx, arg)
}
func (m *mockRoutineStore) ReorderRoutineExerciseForward(ctx context.Context, arg queries.ReorderRoutineExerciseForwardParams) error {
	return m.reorderRoutineExerciseForwardFunc(ctx, arg)
}
func (m *mockRoutineStore) ReorderRoutineExerciseBackward(ctx context.Context, arg queries.ReorderRoutineExerciseBackwardParams) error {
	return m.reorderRoutineExerciseBackwardFunc(ctx, arg)
}
func (m *mockRoutineStore) DeleteRoutineExercise(ctx context.Context, id uuid.UUID) error {
	return m.deleteRoutineExerciseFunc(ctx, id)
}

var routineTestID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var exerciseTestID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
var routineExerciseTestID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

func newMockRoutineService(mock *mockRoutineStore) *RoutineService {
	return &RoutineService{
		store: mock,
		withTx: func(_ context.Context, fn func(store RoutineStore) error) error {
			return fn(mock)
		},
	}
}

func TestRoutineServiceCreate(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
		wantField string
	}{
		{
			name:      "happy path",
			inputName: "Push Day",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					createRoutineFunc: func(_ context.Context, name string) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: name, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "blank name",
			inputName: "   ",
			setupMock: func() *mockRoutineStore { return &mockRoutineStore{} },
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "name",
		},
		{
			name:      "name too long",
			inputName: string(make([]byte, 256)),
			setupMock: func() *mockRoutineStore { return &mockRoutineStore{} },
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := newMockRoutineService(mock)
			_, err := svc.CreateRoutine(context.Background(), tt.inputName)
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

func TestRoutineServiceList(t *testing.T) {
	mock := &mockRoutineStore{
		listRoutinesPaginatedFunc: func(_ context.Context, arg queries.ListRoutinesPaginatedParams) ([]queries.Routine, error) {
			if arg.Limit != 20 || arg.Offset != 0 {
				t.Errorf("expected limit=20 offset=0, got limit=%d offset=%d", arg.Limit, arg.Offset)
			}
			return []queries.Routine{{ID: routineTestID, Name: "Push Day"}}, nil
		},
		countRoutinesFunc: func(_ context.Context) (int64, error) {
			return 1, nil
		},
	}
	svc := newMockRoutineService(mock)
	result, err := svc.ListRoutines(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 routine, got %d", len(result.Data))
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want %d", result.Total, 1)
	}
}

func TestRoutineServiceGet(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
	}{
		{
			name: "happy path",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: "Push Day", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
					},
					listRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) ([]queries.ListRoutineExercisesRow, error) {
						return []queries.ListRoutineExercisesRow{
							{ID: routineExerciseTestID, RoutineID: routineTestID, ExerciseID: exerciseTestID, Order: 1, ExerciseName: "Bench Press", TargetMuscle: pgtype.Text{String: "Chest", Valid: true}},
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{}, pgx.ErrNoRows
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
			svc := newMockRoutineService(mock)
			detail, err := svc.GetRoutine(context.Background(), routineTestID)
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
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(detail.Exercises) != 1 {
					t.Errorf("expected 1 exercise, got %d", len(detail.Exercises))
				}
			}
		})
	}
}

func TestRoutineServiceUpdate(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
		wantField string
	}{
		{
			name:      "happy path",
			inputName: "Pull Day",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: "Push Day"}, nil
					},
					updateRoutineFunc: func(_ context.Context, arg queries.UpdateRoutineParams) (queries.Routine, error) {
						return queries.Routine{ID: arg.ID, Name: arg.Name, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "idempotent same name",
			inputName: "Push Day",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: "Push Day"}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:      "not found",
			inputName: "Leg Day",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:      "validation error",
			inputName: "",
			setupMock: func() *mockRoutineStore { return &mockRoutineStore{} },
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := newMockRoutineService(mock)
			_, err := svc.UpdateRoutine(context.Background(), routineTestID, tt.inputName)
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

func TestRoutineServiceDelete(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
	}{
		{
			name: "happy path",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					deleteRoutineFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{}, pgx.ErrNoRows
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
			svc := newMockRoutineService(mock)
			err := svc.DeleteRoutine(context.Background(), routineTestID)
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

func TestRoutineServiceAddRoutineExercise(t *testing.T) {
	validOrder := int32(1)
	invalidOrder := int32(5)

	tests := []struct {
		name        string
		exerciseID  string
		order       *int32
		setupMock   func() *mockRoutineStore
		wantErr     bool
		wantCode    string
		wantField   string
		wantShiftUp bool
	}{
		{
			name:       "happy path append",
			exerciseID: exerciseTestID.String(),
			order:      nil,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					getExerciseFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: exerciseTestID}, nil
					},
					getRoutineExerciseByExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, pgx.ErrNoRows
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 2, nil
					},
					createRoutineExerciseFunc: func(_ context.Context, arg queries.CreateRoutineExerciseParams) (queries.RoutineExercise, error) {
						if arg.Order != 3 {
							t.Errorf("expected order=3, got %d", arg.Order)
						}
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: arg.RoutineID, ExerciseID: arg.ExerciseID, Order: arg.Order}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:       "happy path insert at beginning",
			exerciseID: exerciseTestID.String(),
			order:      &validOrder,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					getExerciseFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: exerciseTestID}, nil
					},
					getRoutineExerciseByExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, pgx.ErrNoRows
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 2, nil
					},
					shiftRoutineExerciseOrderUpFunc: func(_ context.Context, arg queries.ShiftRoutineExerciseOrderUpParams) error {
						if arg.Order != 1 {
							t.Errorf("expected shift order=1, got %d", arg.Order)
						}
						return nil
					},
					createRoutineExerciseFunc: func(_ context.Context, arg queries.CreateRoutineExerciseParams) (queries.RoutineExercise, error) {
						if arg.Order != 1 {
							t.Errorf("expected order=1, got %d", arg.Order)
						}
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: arg.RoutineID, ExerciseID: arg.ExerciseID, Order: arg.Order}, nil
					},
				}
			},
			wantErr:     false,
			wantShiftUp: true,
		},
		{
			name:       "invalid exercise uuid",
			exerciseID: "not-a-uuid",
			order:      nil,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "exercise_id",
		},
		{
			name:       "routine not found",
			exerciseID: exerciseTestID.String(),
			order:      nil,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:       "exercise not found",
			exerciseID: exerciseTestID.String(),
			order:      nil,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					getExerciseFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "exercise_id",
		},
		{
			name:       "duplicate exercise",
			exerciseID: exerciseTestID.String(),
			order:      nil,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					getExerciseFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: exerciseTestID}, nil
					},
					getRoutineExerciseByExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: "CONFLICT",
		},
		{
			name:       "order exceeds max+1",
			exerciseID: exerciseTestID.String(),
			order:      &invalidOrder,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineFunc: func(_ context.Context, _ uuid.UUID) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID}, nil
					},
					getExerciseFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{ID: exerciseTestID}, nil
					},
					getRoutineExerciseByExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseByExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, pgx.ErrNoRows
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 2, nil
					},
				}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := newMockRoutineService(mock)
			_, err := svc.AddRoutineExercise(context.Background(), routineTestID, tt.exerciseID, tt.order)
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

func TestRoutineServiceUpdateRoutineExerciseOrder(t *testing.T) {
	tests := []struct {
		name      string
		newOrder  int32
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
		wantField string
	}{
		{
			name:     "happy path forward",
			newOrder: 4,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, Order: 2}, nil
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 5, nil
					},
					reorderRoutineExerciseForwardFunc: func(_ context.Context, arg queries.ReorderRoutineExerciseForwardParams) error {
						if arg.Order != 2 || arg.Order_2 != 4 {
							t.Errorf("expected forward old=2 new=4, got old=%d new=%d", arg.Order, arg.Order_2)
						}
						return nil
					},
					updateRoutineExerciseOrderFunc: func(_ context.Context, arg queries.UpdateRoutineExerciseOrderParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: arg.ID, Order: arg.Order}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "happy path backward",
			newOrder: 1,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, Order: 3}, nil
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 5, nil
					},
					reorderRoutineExerciseBackwardFunc: func(_ context.Context, arg queries.ReorderRoutineExerciseBackwardParams) error {
						if arg.Order != 1 || arg.Order_2 != 3 {
							t.Errorf("expected backward new=1 old=3, got new=%d old=%d", arg.Order, arg.Order_2)
						}
						return nil
					},
					updateRoutineExerciseOrderFunc: func(_ context.Context, arg queries.UpdateRoutineExerciseOrderParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: arg.ID, Order: arg.Order}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "idempotent same position",
			newOrder: 2,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, Order: 2}, nil
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 5, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "not found",
			newOrder: 1,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:  true,
			wantCode: "NOT_FOUND",
		},
		{
			name:     "order too high",
			newOrder: 6,
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, Order: 2}, nil
					},
					countRoutineExercisesFunc: func(_ context.Context, _ uuid.UUID) (int64, error) {
						return 5, nil
					},
				}
			},
			wantErr:   true,
			wantCode:  "VALIDATION_ERROR",
			wantField: "order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			svc := newMockRoutineService(mock)
			_, err := svc.UpdateRoutineExerciseOrder(context.Background(), routineTestID, routineExerciseTestID, tt.newOrder)
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

func TestRoutineServiceDeleteRoutineExercise(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mockRoutineStore
		wantErr   bool
		wantCode  string
	}{
		{
			name: "happy path",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, Order: 2}, nil
					},
					deleteRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
					shiftRoutineExerciseOrderDownFunc: func(_ context.Context, arg queries.ShiftRoutineExerciseOrderDownParams) error {
						if arg.Order != 2 {
							t.Errorf("expected shift down order=2, got %d", arg.Order)
						}
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() *mockRoutineStore {
				return &mockRoutineStore{
					getRoutineExerciseFunc: func(_ context.Context, _ queries.GetRoutineExerciseParams) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, pgx.ErrNoRows
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
			svc := newMockRoutineService(mock)
			err := svc.DeleteRoutineExercise(context.Background(), routineTestID, routineExerciseTestID)
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
