package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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
	errsChan := make(chan error, 1)

	go func() {
		s.logger.Info("starting server", slog.String("port", s.cfg.Port))
		errsChan <- s.server.ListenAndServe()

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errsChan <- err
		}
		close(errsChan)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		return nil
	case err := <-errsChan:
		return err
	}
}

func (s Server) Close(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.CtxDefaultTimeout*time.Second)
	defer cancel()
	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) RegisterRoutes(h Handler) *Server {
	s.router.HandleFunc("POST /api/v1/carts", h.Create)
	s.router.HandleFunc("GET /api/v1/carts/{id}", h.View)
	return s
}
