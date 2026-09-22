package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/handlers"
	"github.com/maxon2034/trainee-go-cart-api/internal/repository"
	"github.com/maxon2034/trainee-go-cart-api/internal/server"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
	"github.com/maxon2034/trainee-go-cart-api/pkg/db/postgres"
	"github.com/maxon2034/trainee-go-cart-api/pkg/logger"
)

var cfgPath string = "config/"

func Run(ctx context.Context) {
	logger := logger.New()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("errs in loading config", slog.Any("errs", err))
		return
	}

	DB, err := postgres.New(ctx, cfg.DB.DSN)
	defer DB.Close()
	if err != nil {
		logger.Error("errs in connecting to database", slog.Any("errs", err))
		return
	}

	err = postgres.RunMigrations(DB)
	if err != nil {
		logger.Error("errs in running migrations", slog.Any("errs", err))
		return
	}

	repo := repository.New(DB)

	service := service.New(repo)

	handler := handlers.NewCartHandler(service, logger)

	server := server.New(cfg.Server, logger)

	server.RegisterRoutes(handler)

	err = server.Run(ctx)
	defer server.Close(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("server closed")
			return
		}
		logger.Error("errs in running server", slog.Any("errs", err))
	}

}
