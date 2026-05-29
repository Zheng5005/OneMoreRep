package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockPinger struct {
	pingErr error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.pingErr
}

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantDB     string
	}{
		{
			name:       "healthy",
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantDB:     "ok",
		},
		{
			name:       "unhealthy db",
			pingErr:    errors.New("ping failed"),
			wantStatus: http.StatusOK,
			wantDB:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPinger{pingErr: tt.pingErr}
			h := NewHealth(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rec := httptest.NewRecorder()

			h.Health(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var resp HealthResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Status != "ok" {
				t.Errorf("status = %q, want %q", resp.Status, "ok")
			}
			if resp.DB != tt.wantDB {
				t.Errorf("db = %q, want %q", resp.DB, tt.wantDB)
			}
		})
	}
}
