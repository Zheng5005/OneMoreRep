package server

import (
	"github.com/Zheng5005/onemorerep/internal/handler"
	"github.com/Zheng5005/onemorerep/internal/service"
	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/go-chi/chi/v5"
)

func registerRoutes(r chi.Router, db *store.DB) {
	h := handler.NewHealth(db)

	q := db.Queries()
	exerciseSvc := service.NewExerciseService(q)
	exerciseHandler := handler.NewExercise(exerciseSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)

		r.Route("/exercises", func(r chi.Router) {
			r.Post("/", exerciseHandler.Create)
			r.Get("/", exerciseHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", exerciseHandler.Get)
				r.Put("/", exerciseHandler.Update)
				r.Delete("/", exerciseHandler.Delete)
			})
		})
	})
}
