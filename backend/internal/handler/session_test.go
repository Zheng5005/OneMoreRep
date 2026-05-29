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

type mockSessionService struct {
	createSessionFunc   func(ctx context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error)
	getSessionFunc      func(ctx context.Context, id uuid.UUID) (service.SessionDetail, error)
	getActiveSessionFunc func(ctx context.Context) (*service.SessionDetail, error)
	endSessionFunc      func(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error)
}

func (m *mockSessionService) CreateSession(ctx context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error) {
	return m.createSessionFunc(ctx, routineID)
}
func (m *mockSessionService) GetSession(ctx context.Context, id uuid.UUID) (service.SessionDetail, error) {
	return m.getSessionFunc(ctx, id)
}
func (m *mockSessionService) GetActiveSession(ctx context.Context) (*service.SessionDetail, error) {
	return m.getActiveSessionFunc(ctx)
}
func (m *mockSessionService) EndSession(ctx context.Context, id uuid.UUID) (queries.WorkoutSession, error) {
	return m.endSessionFunc(ctx, id)
}

var wsSessTestID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var wsSessRoutineTestID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

func TestSessionHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func() *mockSessionService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path without routine_id",
			body: `{}`,
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					createSessionFunc: func(_ context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsSessTestID,
							RoutineID: pgtype.UUID{Valid: routineID != nil, Bytes: func() uuid.UUID { if routineID != nil { return *routineID }; return uuid.Nil }()},
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "happy path with routine_id",
			body: `{"routine_id":"22222222-2222-2222-2222-222222222222"}`,
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					createSessionFunc: func(_ context.Context, routineID *uuid.UUID) (queries.WorkoutSession, error) {
						if routineID == nil || *routineID != wsSessRoutineTestID {
							t.Errorf("expected routineID=%v, got %v", wsSessRoutineTestID, routineID)
						}
						return queries.WorkoutSession{
							ID:        wsSessTestID,
							RoutineID: pgtype.UUID{Bytes: wsSessRoutineTestID, Valid: true},
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{"routine_id":`,
			setupMock:  func() *mockSessionService { return &mockSessionService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid routine_id",
			body:       `{"routine_id":"not-a-uuid"}`,
			setupMock:  func() *mockSessionService { return &mockSessionService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSession(mock)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", strings.NewReader(tt.body))
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
			}
		})
	}
}

func TestSessionHandlerGet(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockSessionService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					getSessionFunc: func(_ context.Context, _ uuid.UUID) (service.SessionDetail, error) {
						return service.SessionDetail{
							Session: queries.GetSessionWithSetsRow{
								ID:        wsSessTestID,
								StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
								Sets:      []interface{}{},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					getSessionFunc: func(_ context.Context, _ uuid.UUID) (service.SessionDetail, error) {
						return service.SessionDetail{}, service.NewNotFoundError("session not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/sessions/not-a-uuid",
			setupMock:  func() *mockSessionService { return &mockSessionService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSession(mock)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get("/api/v1/sessions/{id}", h.Get)
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

func TestSessionHandlerGetActive(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func() *mockSessionService
		wantStatus int
	}{
		{
			name: "happy path with active session",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					getActiveSessionFunc: func(_ context.Context) (*service.SessionDetail, error) {
						return &service.SessionDetail{
							Session: queries.GetSessionWithSetsRow{
								ID:        wsSessTestID,
								StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
								Sets:      []interface{}{},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "no active session returns null",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					getActiveSessionFunc: func(_ context.Context) (*service.SessionDetail, error) {
						return nil, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSession(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/active", nil)
			rec := httptest.NewRecorder()

			h.GetActive(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestSessionHandlerEnd(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func() *mockSessionService
		wantStatus int
		wantCode   string
	}{
		{
			name: "happy path",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/end",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					endSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{
							ID:        wsSessTestID,
							StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							EndedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/end",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					endSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, service.NewNotFoundError("session not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "already ended",
			path: "/api/v1/sessions/11111111-1111-1111-1111-111111111111/end",
			setupMock: func() *mockSessionService {
				return &mockSessionService{
					endSessionFunc: func(_ context.Context, _ uuid.UUID) (queries.WorkoutSession, error) {
						return queries.WorkoutSession{}, service.NewConflictError("session already ended")
					},
				}
			},
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name:       "invalid uuid",
			path:       "/api/v1/sessions/not-a-uuid/end",
			setupMock:  func() *mockSessionService { return &mockSessionService{} },
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()
			h := NewSession(mock)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/api/v1/sessions/{id}/end", h.End)
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