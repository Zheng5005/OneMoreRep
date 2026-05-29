package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is the interface required by the Health handler.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Health handles health check requests.
type Health struct {
	db Pinger
}

// NewHealth creates a new Health handler.
func NewHealth(db Pinger) *Health {
	return &Health{db: db}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// Health responds with the service health status.
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := HealthResponse{Status: "ok"}

	if err := h.db.Ping(ctx); err != nil {
		resp.DB = "error"
	} else {
		resp.DB = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
