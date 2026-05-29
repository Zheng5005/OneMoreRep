package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockProgressService struct {
	getExerciseLastValuesFunc func(ctx context.Context, exerciseID uuid.UUID) (*service.LastValues, error)
	getSessionSummaryFunc     func(ctx context.Context, sessionID uuid.UUID) (*service.SessionSummary, error)
}

func (m *mockProgressService) GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (*service.LastValues, error) {
	return m.getExerciseLastValuesFunc(ctx, exerciseID)
}

func (m *mockProgressService) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*service.SessionSummary, error) {
	return m.getSessionSummaryFunc(ctx, sessionID)
}

func TestProgressHandlerGetLastValues(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockProgressService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path returns values",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111/last-values",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getExerciseLastValuesFunc: func(_ context.Context, _ uuid.UUID) (*service.LastValues, error) {
						w := 135.0
						r := 10.0
						return &service.LastValues{Weight: &w, Reps: &r}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns null fields when no history",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111/last-values",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getExerciseLastValuesFunc: func(_ context.Context, _ uuid.UUID) (*service.LastValues, error) {
						return nil, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/exercises/invalid-uuid/last-values",
			setupMock:  func() *mockProgressService { return &mockProgressService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewProgress(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/api/v1/exercises/{id}/last-values", h.GetLastValues)
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

func TestProgressHandlerGetSessionSummary(t *testing.T) {
	sessionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockProgressService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path for ended session",
			path: "/api/v1/sessions/33333333-3333-3333-3333-333333333333/summary",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getSessionSummaryFunc: func(_ context.Context, _ uuid.UUID) (*service.SessionSummary, error) {
						return &service.SessionSummary{
							SessionID:     sessionID,
							StartedAt:     "2024-01-01T10:00:00Z",
							DurationSecs:  3600,
							TotalVolume:   4050,
							ExerciseCount: 2,
							TotalSets:     6,
							Exercises:     []service.ExerciseBreakdown{},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "happy path for in-progress session",
			path: "/api/v1/sessions/33333333-3333-3333-3333-333333333333/summary",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getSessionSummaryFunc: func(_ context.Context, _ uuid.UUID) (*service.SessionSummary, error) {
						return &service.SessionSummary{
							SessionID:     sessionID,
							StartedAt:     "2024-01-01T10:00:00Z",
							DurationSecs:  1800,
							TotalVolume:   810,
							ExerciseCount: 1,
							TotalSets:     3,
							Exercises:     []service.ExerciseBreakdown{},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "session not found",
			path: "/api/v1/sessions/33333333-3333-3333-3333-333333333333/summary",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getSessionSummaryFunc: func(_ context.Context, _ uuid.UUID) (*service.SessionSummary, error) {
						return nil, service.NewNotFoundError("session not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/sessions/invalid-uuid/summary",
			setupMock:  func() *mockProgressService { return &mockProgressService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewProgress(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/api/v1/sessions/{id}/summary", h.GetSessionSummary)
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