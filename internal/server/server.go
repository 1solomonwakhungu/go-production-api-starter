package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/config"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/handler"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/middleware"
)

// Server encapsulates the HTTP server and its dependencies.
type Server struct {
	router *http.ServeMux
	server *http.Server
	logger *slog.Logger
}

// New creates a new Server, wiring up routes and middleware.
func New(cfg *config.Config, h *handler.Handler, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// --- Public routes ---
	mux.HandleFunc("GET /health", h.Health)

	// --- Protected routes (require JWT auth) ---
	// We wrap each handler with JWT auth middleware individually.
	auth := middleware.JWTAuth(cfg.JWTSecret)

	mux.Handle("POST /users", auth(http.HandlerFunc(h.CreateUser)))
	mux.Handle("GET /users", auth(http.HandlerFunc(h.ListUsers)))
	mux.Handle("GET /users/{id}", auth(http.HandlerFunc(h.GetUser)))
	mux.Handle("PUT /users/{id}", auth(http.HandlerFunc(h.UpdateUser)))
	mux.Handle("DELETE /users/{id}", auth(http.HandlerFunc(h.DeleteUser)))

	// --- Global middleware chain ---
	var handler http.Handler = mux
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Recover(logger)(handler)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router: mux,
		server: srv,
		logger: logger,
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	s.logger.Info("starting http server", slog.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}
