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

	routineSvc := service.NewRoutineService(db)
	routineHandler := handler.NewRoutine(routineSvc)

	sessionSvc := service.NewSessionService(db)
	sessionHandler := handler.NewSession(sessionSvc)

	setSvc := service.NewSetService(db)
	setHandler := handler.NewSet(setSvc)

	progressSvc := service.NewProgressService(q)
	progressHandler := handler.NewProgress(progressSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)

		r.Route("/exercises", func(r chi.Router) {
			r.Post("/", exerciseHandler.Create)
			r.Get("/", exerciseHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", exerciseHandler.Get)
				r.Get("/last-values", progressHandler.GetLastValues)
				r.Put("/", exerciseHandler.Update)
				r.Delete("/", exerciseHandler.Delete)
			})
		})

		r.Route("/routines", func(r chi.Router) {
			r.Post("/", routineHandler.Create)
			r.Get("/", routineHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", routineHandler.Get)
				r.Put("/", routineHandler.Update)
				r.Delete("/", routineHandler.Delete)

				r.Route("/exercises", func(r chi.Router) {
					r.Post("/", routineHandler.AddExercise)
					r.Route("/{routineExerciseId}", func(r chi.Router) {
						r.Put("/", routineHandler.UpdateExerciseOrder)
						r.Delete("/", routineHandler.DeleteExercise)
					})
				})
			})
		})

		r.Route("/sessions", func(r chi.Router) {
			r.Post("/", sessionHandler.Create)
			r.Get("/active", sessionHandler.GetActive)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", sessionHandler.Get)
				r.Post("/end", sessionHandler.End)
				r.Get("/summary", progressHandler.GetSessionSummary)
				r.Post("/sets", setHandler.Create)
				r.Route("/sets/{setId}", func(r chi.Router) {
					r.Put("/", setHandler.Update)
					r.Delete("/", setHandler.Delete)
				})
			})
		})
	})
}
