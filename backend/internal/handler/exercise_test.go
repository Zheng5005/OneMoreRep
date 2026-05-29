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

type mockExerciseService struct {
	createFunc func(ctx context.Context, name, targetMuscle, notes string) (queries.Exercise, error)
	listFunc   func(ctx context.Context, limit, offset int32, search string) (service.ExerciseListResult, error)
	getFunc    func(ctx context.Context, id uuid.UUID) (queries.Exercise, error)
	updateFunc func(ctx context.Context, id uuid.UUID, name, targetMuscle, notes string) (queries.Exercise, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockExerciseService) CreateExercise(ctx context.Context, name, targetMuscle, notes string) (queries.Exercise, error) {
	return m.createFunc(ctx, name, targetMuscle, notes)
}

func (m *mockExerciseService) ListExercises(ctx context.Context, limit, offset int32, search string) (service.ExerciseListResult, error) {
	return m.listFunc(ctx, limit, offset, search)
}

func (m *mockExerciseService) GetExercise(ctx context.Context, id uuid.UUID) (queries.Exercise, error) {
	return m.getFunc(ctx, id)
}

func (m *mockExerciseService) UpdateExercise(ctx context.Context, id uuid.UUID, name, targetMuscle, notes string) (queries.Exercise, error) {
	return m.updateFunc(ctx, id, name, targetMuscle, notes)
}

func (m *mockExerciseService) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

var testID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func TestExerciseHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func() *mockExerciseService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			body: `{"name":"Bench Press","target_muscle":"Chest","notes":"Keep back flat"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					createFunc: func(_ context.Context, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							Notes:        pgtype.Text{String: "Keep back flat", Valid: true},
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "validation error",
			body: `{"name":"","target_muscle":"Chest"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					createFunc: func(_ context.Context, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewValidationError("name", "cannot be blank")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "name",
		},
		{
			name: "duplicate conflict",
			body: `{"name":"Bench Press","target_muscle":"Chest"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					createFunc: func(_ context.Context, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewConflictError("exercise with same name and target muscle already exists")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name:       "invalid json",
			body:       `{"name":`,
			setupMock:  func() *mockExerciseService { return &mockExerciseService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewExercise(mock)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/exercises", strings.NewReader(tt.body))
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

func TestExerciseHandlerList(t *testing.T) {
	mock := &mockExerciseService{
		listFunc: func(_ context.Context, limit, offset int32, search string) (service.ExerciseListResult, error) {
			return service.ExerciseListResult{
				Data: []queries.Exercise{
					{
						ID:           testID,
						Name:         "Bench Press",
						TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
						CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				Limit:  20,
				Offset: 0,
				Total:  1,
			}, nil
		},
	}

	h := NewExercise(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises?limit=20&offset=0&search=bench", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp ExerciseListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Bench Press" {
		t.Errorf("name = %q, want %q", resp.Data[0].Name, "Bench Press")
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("total = %d, want %d", resp.Pagination.Total, 1)
	}
	if resp.Pagination.Limit != 20 {
		t.Errorf("limit = %d, want %d", resp.Pagination.Limit, 20)
	}
}

func TestExerciseHandlerGet(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockExerciseService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
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
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					getFunc: func(_ context.Context, _ uuid.UUID) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewNotFoundError("exercise not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/exercises/invalid-uuid",
			setupMock:  func() *mockExerciseService { return &mockExerciseService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewExercise(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/api/v1/exercises/{id}", h.Get)
			r.ServeHTTP(rec, req)

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

func TestExerciseHandlerUpdate(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockExerciseService
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name: "happy path",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			body: `{"name":"Bench Press","target_muscle":"Chest","notes":"Updated"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{
							ID:           testID,
							Name:         "Bench Press",
							TargetMuscle: pgtype.Text{String: "Chest", Valid: true},
							Notes:        pgtype.Text{String: "Updated", Valid: true},
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			body: `{"name":"Bench Press","target_muscle":"Chest"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewNotFoundError("exercise not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "conflict",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			body: `{"name":"Bench Press","target_muscle":"Chest"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewConflictError("exercise with same name and target muscle already exists")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name: "validation error",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			body: `{"name":"","target_muscle":"Chest"}`,
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					updateFunc: func(_ context.Context, _ uuid.UUID, _, _, _ string) (queries.Exercise, error) {
						return queries.Exercise{}, service.NewValidationError("name", "cannot be blank")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantField:  "name",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/exercises/bad-uuid",
			body:       `{"name":"Bench Press","target_muscle":"Chest"}`,
			setupMock:  func() *mockExerciseService { return &mockExerciseService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			body:       `{"name":`,
			setupMock:  func() *mockExerciseService { return &mockExerciseService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewExercise(mock)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Put("/api/v1/exercises/{id}", h.Update)
			r.ServeHTTP(rec, req)

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

func TestExerciseHandlerDelete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockExerciseService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return service.NewNotFoundError("exercise not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "referenced resource",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockExerciseService {
				return &mockExerciseService{
					deleteFunc: func(_ context.Context, _ uuid.UUID) error {
						return service.NewReferencedResourceError("exercise is referenced by routines")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "REFERENCED_RESOURCE",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/exercises/bad-uuid",
			setupMock:  func() *mockExerciseService { return &mockExerciseService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewExercise(mock)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Delete("/api/v1/exercises/{id}", h.Delete)
			r.ServeHTTP(rec, req)

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
