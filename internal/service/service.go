package service

import (
	"log/slog"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

type CartService struct {
	repo   Repository
	logger *slog.Logger
	cfg    config.CartConfig
}

func New(rep Repository, log *slog.Logger, cfg config.CartConfig) *CartService {
	return &CartService{repo: rep, logger: log, cfg: cfg}
}
