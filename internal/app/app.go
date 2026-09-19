package app

import (
	"context"
	"errors"
	"fmt"
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
		logger.Error("error in loading config", slog.Any("error", err))
		return
	}

	DB, err := postgres.New(ctx, cfg.DB.DSN)
	if err != nil {
		logger.Error("error in connecting to database", slog.Any("error", err))
		return
	}

	err = postgres.RunMigrations(DB)
	if err != nil {
		logger.Error("error in running migrations", slog.Any("error", err))
	}

	repo := repository.New(DB)

	service := service.New(repo)

	handler := handlers.NewCartHandler(service, logger)

	server := server.New(cfg.Server, logger)

	server.RegisterRoutes(handler)

	logger.Info(fmt.Sprintf("app started on port %s", cfg.Server.Port))
	err = server.Run()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("server closed")
			return
		}
		logger.Error("error in running server", slog.Any("error", err))
	}

}
