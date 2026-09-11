package handlers

import (
	"context"
	"net"
	"net/http"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

func NewServer(ctx context.Context, cfg config.Config, service Service) *http.Server {
	mux := http.NewServeMux()

	cartHandler := NewCartHandler(service)

	mux.HandleFunc("POST /api/v1/carts", cartHandler.Create)

	return &http.Server{
		Addr:    cfg.Server.Port,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
}
