package server

import (
	"github.com/Zheng5005/onemorerep/internal/handler"
	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/go-chi/chi/v5"
)

func registerRoutes(r chi.Router, db *store.DB) {
	h := handler.NewHealth(db)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)
	})
}
