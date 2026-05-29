package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Zheng5005/onemorerep/internal/config"
	"github.com/Zheng5005/onemorerep/internal/store"
	"github.com/go-chi/chi/v5"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	mux        *chi.Mux
	cfg        config.Config
	db         *store.DB
}

// New creates a new Server instance with configured routes and middleware.
func New(cfg config.Config, db *store.DB) *Server {
	mux := chi.NewRouter()

	attachMiddleware(mux, cfg)

	srv := &Server{
		mux: mux,
		cfg: cfg,
		db:  db,
		httpServer: &http.Server{
			Addr:         ":8080",
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	registerRoutes(mux, db)

	return srv
}

// Start begins listening for incoming HTTP requests.
func (s *Server) Start(addr string) error {
	s.httpServer.Addr = addr
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
