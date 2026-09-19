package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

type Server struct {
	cfg    config.ServerConfig
	server *http.Server
	logger *slog.Logger
	router *http.ServeMux
}

func New(cfg config.ServerConfig, logger *slog.Logger) *Server {
	router := http.NewServeMux()
	server := &http.Server{
		Addr:         cfg.Port,
		WriteTimeout: cfg.WriteTimeout * time.Second,
		ReadTimeout:  cfg.ReadTimeout * time.Second,
		Handler:      router,
	}
	return &Server{
		cfg:    cfg,
		server: server,
		logger: logger,
		router: router,
	}
}

func (s *Server) Run() error {
	// TODO: init server running
	return s.server.ListenAndServe()
}

func (s *Server) RegisterRoutes(h Handler) *Server {
	s.router.HandleFunc("POST /api/v1/carts", h.Create)
	s.router.HandleFunc("GET /api/v1/carts/{id}", h.View)
	return s
}
