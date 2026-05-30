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
	getExerciseLastValuesFunc  func(ctx context.Context, exerciseID uuid.UUID) (*service.LastValues, error)
	getSessionSummaryFunc      func(ctx context.Context, sessionID uuid.UUID) (*service.SessionSummary, error)
	getExerciseHistoryFunc     func(ctx context.Context, exerciseID uuid.UUID, filter string) (*service.ExerciseHistory, error)
	getVolumeAggregationFunc   func(ctx context.Context, groupBy string, exerciseID *uuid.UUID) ([]service.VolumePeriod, error)
}

func (m *mockProgressService) GetExerciseLastValues(ctx context.Context, exerciseID uuid.UUID) (*service.LastValues, error) {
	return m.getExerciseLastValuesFunc(ctx, exerciseID)
}

func (m *mockProgressService) GetSessionSummary(ctx context.Context, sessionID uuid.UUID) (*service.SessionSummary, error) {
	return m.getSessionSummaryFunc(ctx, sessionID)
}

func (m *mockProgressService) GetExerciseHistory(ctx context.Context, exerciseID uuid.UUID, filter string) (*service.ExerciseHistory, error) {
	return m.getExerciseHistoryFunc(ctx, exerciseID, filter)
}

func (m *mockProgressService) GetVolumeAggregation(ctx context.Context, groupBy string, exerciseID *uuid.UUID) ([]service.VolumePeriod, error) {
	return m.getVolumeAggregationFunc(ctx, groupBy, exerciseID)
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

func TestProgressHandlerGetExerciseHistory(t *testing.T) {
	sessionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockProgressService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path returns history",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111/history",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getExerciseHistoryFunc: func(_ context.Context, _ uuid.UUID, _ string) (*service.ExerciseHistory, error) {
						return &service.ExerciseHistory{
							Sessions: []service.ExerciseHistorySession{
								{
									SessionID: sessionID,
									StartedAt: "2024-01-01T09:00:00Z",
									Sets: []service.ExerciseHistorySet{
										{SetID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), SetNumber: 1, Weight: 135, Reps: 10, Volume: 1350, IsPR: true},
									},
								},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty history returns empty sessions",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111/history",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getExerciseHistoryFunc: func(_ context.Context, _ uuid.UUID, _ string) (*service.ExerciseHistory, error) {
						return &service.ExerciseHistory{Sessions: []service.ExerciseHistorySession{}}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid filter returns 400",
			path: "/api/v1/exercises/11111111-1111-1111-1111-111111111111/history?filter=invalid",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getExerciseHistoryFunc: func(_ context.Context, _ uuid.UUID, _ string) (*service.ExerciseHistory, error) {
						return nil, service.NewBadRequestError("invalid filter value: must be all, 30d, or 6m")
					},
				}
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/exercises/invalid-uuid/history",
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
			r.Get("/api/v1/exercises/{id}/history", h.GetExerciseHistory)
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

func TestProgressHandlerGetVolumeAggregation(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockProgressService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path with group_by session",
			path: "/api/v1/progress/volume?group_by=session",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getVolumeAggregationFunc: func(_ context.Context, _ string, _ *uuid.UUID) ([]service.VolumePeriod, error) {
						return []service.VolumePeriod{
							{Period: "session-1", TotalVolume: 5000},
							{Period: "session-2", TotalVolume: 6000},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "happy path with group_by week",
			path: "/api/v1/progress/volume?group_by=week",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getVolumeAggregationFunc: func(_ context.Context, _ string, _ *uuid.UUID) ([]service.VolumePeriod, error) {
						return []service.VolumePeriod{
							{Period: "2024-W01", TotalVolume: 15000},
							{Period: "2024-W02", TotalVolume: 18000},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "happy path with group_by month",
			path: "/api/v1/progress/volume?group_by=month",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getVolumeAggregationFunc: func(_ context.Context, _ string, _ *uuid.UUID) ([]service.VolumePeriod, error) {
						return []service.VolumePeriod{
							{Period: "2024-01", TotalVolume: 50000},
							{Period: "2024-02", TotalVolume: 65000},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing group_by returns 400",
			path:       "/api/v1/progress/volume",
			setupMock:  func() *mockProgressService { return &mockProgressService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "invalid group_by returns 400",
			path: "/api/v1/progress/volume?group_by=invalid",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getVolumeAggregationFunc: func(_ context.Context, _ string, _ *uuid.UUID) ([]service.VolumePeriod, error) {
						return nil, service.NewBadRequestError("invalid group_by value: must be session, week, or month")
					},
				}
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid exercise_id returns 400",
			path:       "/api/v1/progress/volume?group_by=session&exercise_id=invalid",
			setupMock:  func() *mockProgressService { return &mockProgressService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name: "exercise_id filter works",
			path: "/api/v1/progress/volume?group_by=session&exercise_id=11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockProgressService {
				return &mockProgressService{
					getVolumeAggregationFunc: func(_ context.Context, _ string, _ *uuid.UUID) ([]service.VolumePeriod, error) {
						return []service.VolumePeriod{
							{Period: "session-1", TotalVolume: 1350},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewProgress(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/api/v1/progress/volume", h.GetVolumeAggregation)
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