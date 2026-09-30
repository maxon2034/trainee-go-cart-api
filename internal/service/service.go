package service

import "log/slog"

type CartService struct {
	repo   Repository
	logger *slog.Logger
}

func New(rep Repository, log *slog.Logger) *CartService { return &CartService{repo: rep, logger: log} }
