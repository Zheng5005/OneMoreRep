package handler

import (
	"context"
	"encoding/json"
	"math/big"
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

type mockSetServiceHandler struct {
	createSetFunc func(ctx context.Context, sessionID uuid.UUID, exerciseID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error)
	updateSetFunc func(ctx context.Context, setID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error)
	deleteSetFunc func(ctx context.Context, setID uuid.UUID) error
}

func (m *mockSetServiceHandler) CreateSet(ctx context.Context, sessionID uuid.UUID, exerciseID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error) {
	return m.createSetFunc(ctx, sessionID, exerciseID, weight, reps)
}
func (m *mockSetServiceHandler) UpdateSet(ctx context.Context, setID uuid.UUID, weight float64, reps int) (queries.WorkoutSet, error) {
	return m.updateSetFunc(ctx, setID, weight, reps)
}
func (m *mockSetServiceHandler) DeleteSet(ctx context.Context, setID uuid.UUID) error {
	return m.deleteSetFunc(ctx, setID)
}

var wsHSetTestID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
var wsHSessTestID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var wsHSetExerciseTestID = uuid.MustParse("44444444-4444-4444-4444-444444444444")

func TestSetHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockSetServiceHandler
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body: `{"exercise_id":"44444444-4444-4444-4444-444444444444","weight":135.5,"reps":10}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					createSetFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:         wsHSetTestID,
							SessionID:  wsHSessTestID,
							ExerciseID: wsHSetExerciseTestID,
							SetNumber:  1,
							Weight:     pgtype.Numeric{Int: big.NewInt(1355), Exp: -1, Valid: true},
							Reps:       10,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid session uuid",
			path:       "/api/v1/sessions/bad-uuid/sets",
			body:       `{"exercise_id":"44444444-4444-4444-4444-444444444444","weight":135.5,"reps":10}`,
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid exercise uuid",
			path:       "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body:       `{"exercise_id":"bad-uuid","weight":135.5,"reps":10}`,
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body:       `{"exercise_id":`,
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "session not found",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body: `{"exercise_id":"44444444-4444-4444-4444-444444444444","weight":135.5,"reps":10}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					createSetFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, service.NewNotFoundError("session not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "session ended",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body: `{"exercise_id":"44444444-4444-4444-4444-444444444444","weight":135.5,"reps":10}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					createSetFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, service.NewConflictError("session already ended")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name: "validation error",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets",
			body: `{"exercise_id":"44444444-4444-4444-4444-444444444444","weight":-10,"reps":10}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					createSetFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, service.NewValidationError("weight", "must be >= 0")
					},
				}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSet(mock)

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/api/v1/sessions/{id}/sets", h.Create)
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

func TestSetHandlerUpdate(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func() *mockSetServiceHandler
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			body: `{"weight":155.0,"reps":8}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					updateSetFunc: func(_ context.Context, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{
							ID:        wsHSetTestID,
							SessionID: wsHSessTestID,
							Weight:    pgtype.Numeric{Int: big.NewInt(155), Exp: 0, Valid: true},
							Reps:      8,
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid set uuid",
			path:       "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/bad-uuid",
			body:       `{"weight":155.0,"reps":8}`,
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid json",
			path:       "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			body:       `{"weight":`,
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "set not found",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			body: `{"weight":155.0,"reps":8}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					updateSetFunc: func(_ context.Context, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, service.NewNotFoundError("set not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "session ended",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			body: `{"weight":155.0,"reps":8}`,
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					updateSetFunc: func(_ context.Context, _ uuid.UUID, _ float64, _ int) (queries.WorkoutSet, error) {
						return queries.WorkoutSet{}, service.NewConflictError("session already ended")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSet(mock)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Put("/api/v1/sessions/{id}/sets/{setId}", h.Update)
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

func TestSetHandlerDelete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockSetServiceHandler
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					deleteSetFunc: func(_ context.Context, _ uuid.UUID) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid set uuid",
			path:       "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/bad-uuid",
			setupMock:  func() *mockSetServiceHandler { return &mockSetServiceHandler{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "set not found",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/sets/33333333-3333-3333-3333-333333333333",
			setupMock: func() *mockSetServiceHandler {
				return &mockSetServiceHandler{
					deleteSetFunc: func(_ context.Context, _ uuid.UUID) error {
						return service.NewNotFoundError("set not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSet(mock)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Delete("/api/v1/sessions/{id}/sets/{setId}", h.Delete)
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