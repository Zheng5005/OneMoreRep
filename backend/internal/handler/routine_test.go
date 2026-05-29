package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store/queries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockRoutineService struct {
	createFunc                func(ctx context.Context, name string) (queries.Routine, error)
	listFunc                  func(ctx context.Context, limit, offset int32) (service.RoutineListResult, error)
	getFunc                   func(ctx context.Context, id uuid.UUID) (service.RoutineDetail, error)
	updateFunc                func(ctx context.Context, id uuid.UUID, name string) (queries.Routine, error)
	deleteFunc                func(ctx context.Context, id uuid.UUID) error
	addRoutineExerciseFunc    func(ctx context.Context, routineID uuid.UUID, exerciseIDStr string, order *int32) (queries.RoutineExercise, error)
	updateExerciseOrderFunc   func(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID, newOrder int32) (queries.RoutineExercise, error)
	deleteRoutineExerciseFunc func(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID) error
}

func (m *mockRoutineService) CreateRoutine(ctx context.Context, name string) (queries.Routine, error) {
	return m.createFunc(ctx, name)
}
func (m *mockRoutineService) ListRoutines(ctx context.Context, limit, offset int32) (service.RoutineListResult, error) {
	return m.listFunc(ctx, limit, offset)
}
func (m *mockRoutineService) GetRoutine(ctx context.Context, id uuid.UUID) (service.RoutineDetail, error) {
	return m.getFunc(ctx, id)
}
func (m *mockRoutineService) UpdateRoutine(ctx context.Context, id uuid.UUID, name string) (queries.Routine, error) {
	return m.updateFunc(ctx, id, name)
}
func (m *mockRoutineService) DeleteRoutine(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}
func (m *mockRoutineService) AddRoutineExercise(ctx context.Context, routineID uuid.UUID, exerciseIDStr string, order *int32) (queries.RoutineExercise, error) {
	return m.addRoutineExerciseFunc(ctx, routineID, exerciseIDStr, order)
}
func (m *mockRoutineService) UpdateRoutineExerciseOrder(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID, newOrder int32) (queries.RoutineExercise, error) {
	return m.updateExerciseOrderFunc(ctx, routineID, routineExerciseID, newOrder)
}
func (m *mockRoutineService) DeleteRoutineExercise(ctx context.Context, routineID uuid.UUID, routineExerciseID uuid.UUID) error {
	return m.deleteRoutineExerciseFunc(ctx, routineID, routineExerciseID)
}

var routineTestID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var exerciseTestID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
var routineExerciseTestID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

func TestRoutineHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			body: `{"name":"Push Day"}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					createFunc: func(_ context.Context, name string) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: name, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "validation error",
			body: `{"name":""}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					createFunc: func(_ context.Context, _ string) (queries.Routine, error) {
						return queries.Routine{}, service.NewValidationError("name", "cannot be blank")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "name",
		},
		{
			name:       "invalid json",
			body:       `{"name":`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/routines", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
				if tt.wantField != "" && resp.Error.Field != tt.wantField {
					t.Errorf("field = %q, want %q", resp.Error.Field, tt.wantField)
				}
			}
		})
	}
}

func TestRoutineHandlerList(t *testing.T) {
	mock := &mockRoutineService{
		listFunc: func(_ context.Context, limit, offset int32) (service.RoutineListResult, error) {
			return service.RoutineListResult{
				Data: []queries.Routine{
					{ID: routineTestID, Name: "Push Day", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				Limit:  20,
				Offset: 0,
				Total:  1,
			}, nil
		},
	}

	h := NewRoutine(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/routines?limit=20&offset=0", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp RoutineListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 routine, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Push Day" {
		t.Errorf("name = %q, want %q", resp.Data[0].Name, "Push Day")
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("total = %d, want %d", resp.Pagination.Total, 1)
	}
}

func TestRoutineHandlerGet(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					getFunc: func(_ context.Context, _ uuid.UUID) (service.RoutineDetail, error) {
						return service.RoutineDetail{
							Routine: queries.Routine{ID: routineTestID, Name: "Push Day", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
							Exercises: []queries.ListRoutineExercisesRow{
								{ID: routineExerciseTestID, RoutineID: routineTestID, ExerciseID: exerciseTestID, Order: 1, ExerciseName: "Bench Press", TargetMuscle: pgtype.Text{String: "Chest", Valid: true}},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					getFunc: func(_ context.Context, _ uuid.UUID) (service.RoutineDetail, error) {
						return service.RoutineDetail{}, service.NewNotFoundError("routine not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/routines/invalid-uuid",
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get("/api/v1/routines/{id}", h.Get)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestRoutineHandlerUpdate(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			body: `{"name":"Pull Day"}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateFunc: func(_ context.Context, _ uuid.UUID, name string) (queries.Routine, error) {
						return queries.Routine{ID: routineTestID, Name: name, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			body: `{"name":"Pull Day"}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _ string) (queries.Routine, error) {
						return queries.Routine{}, service.NewNotFoundError("routine not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "validation error",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			body: `{"name":""}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _ string) (queries.Routine, error) {
						return queries.Routine{}, service.NewValidationError("name", "cannot be blank")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "name",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/routines/bad-uuid",
			body:       `{"name":"Pull Day"}`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			body:       `{"name":`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Put("/api/v1/routines/{id}", h.Update)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
				if tt.wantField != "" && resp.Error.Field != tt.wantField {
					t.Errorf("field = %q, want %q", resp.Error.Field, tt.wantField)
				}
			}
		})
	}
}

func TestRoutineHandlerDelete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return service.NewNotFoundError("routine not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/routines/bad-uuid",
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Delete("/api/v1/routines/{id}", h.Delete)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestRoutineHandlerAddExercise(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises",
			body: `{"exercise_id":"22222222-2222-2222-2222-222222222222","order":1}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					addRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, ExerciseID: exerciseTestID, Order: 1}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "duplicate exercise conflict",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises",
			body: `{"exercise_id":"22222222-2222-2222-2222-222222222222"}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					addRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, service.NewConflictError("exercise already in routine")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name: "order validation error",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises",
			body: `{"exercise_id":"22222222-2222-2222-2222-222222222222","order":5}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					addRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, service.NewValidationError("order", "order must be between 1 and max+1")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "order",
		},
		{
			name:       "invalid routine uuid",
			path:       "/api/v1/routines/bad-uuid/exercises",
			body:       `{"exercise_id":"22222222-2222-2222-2222-222222222222"}`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises",
			body:       `{"exercise_id":`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/api/v1/routines/{id}/exercises", h.AddExercise)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
				if tt.wantField != "" && resp.Error.Field != tt.wantField {
					t.Errorf("field = %q, want %q", resp.Error.Field, tt.wantField)
				}
			}
		})
	}
}

func TestRoutineHandlerUpdateExerciseOrder(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			body: `{"order":3}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateExerciseOrderFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, newOrder int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{ID: routineExerciseTestID, RoutineID: routineTestID, ExerciseID: exerciseTestID, Order: newOrder}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			body: `{"order":3}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateExerciseOrderFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, service.NewNotFoundError("routine exercise not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "order validation",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			body: `{"order":10}`,
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					updateExerciseOrderFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32) (queries.RoutineExercise, error) {
						return queries.RoutineExercise{}, service.NewValidationError("order", "order must be between 1 and max")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "order",
		},
		{
			name:       "invalid routine uuid",
			path:       "/api/v1/routines/bad-uuid/exercises/33333333-3333-3333-3333-333333333333",
			body:       `{"order":3}`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid routine exercise uuid",
			path:       "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/bad-uuid",
			body:       `{"order":3}`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			body:       `{"order":`,
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Put("/api/v1/routines/{id}/exercises/{routineExerciseId}", h.UpdateExerciseOrder)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
				if tt.wantField != "" && resp.Error.Field != tt.wantField {
					t.Errorf("field = %q, want %q", resp.Error.Field, tt.wantField)
				}
			}
		})
	}
}

func TestRoutineHandlerDeleteExercise(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockRoutineService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					deleteRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			path: "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/33333333-3333-3333-3333-333333333333",
			setupMock: func() *mockRoutineService {
				return &mockRoutineService{
					deleteRoutineExerciseFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
						return service.NewNotFoundError("routine exercise not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid routine uuid",
			path:       "/api/v1/routines/bad-uuid/exercises/33333333-3333-3333-3333-333333333333",
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid routine exercise uuid",
			path:       "/api/v1/routines/11111111-1111-1111-1111-111111111111/exercises/bad-uuid",
			setupMock:  func() *mockRoutineService { return &mockRoutineService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewRoutine(mock)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Delete("/api/v1/routines/{id}/exercises/{routineExerciseId}", h.DeleteExercise)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
				}
			}
		})
	}
}
