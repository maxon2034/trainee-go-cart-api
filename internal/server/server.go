package server

import (
	"log/slog"
	"net/http"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

type Server struct {
	cfg    config.Config
	server *http.Server
	logger *slog.Logger
	mux    http.ServeMux
}

func (s *Server) Run() {
	// TODO: init server running
}

func (s *Server) ResolveHandlers(h Handler) *Server {
	s.mux.HandleFunc("POST /api/v1/carts", h.Create)
	s.mux.HandleFunc("GET /api/v1/carts/{id}", h.View)
	return s
}
