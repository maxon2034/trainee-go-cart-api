package app

import (
	"context"
	"errors"
	"log"
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
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Print("error in loading config", slog.Any("error", err))
		return
	}

	logger, err := logger.New(cfg)
	if err != nil {
		log.Print("error in initializing logger", slog.Any("error", err))
		return
	}

	DB, err := postgres.New(ctx, cfg.DB.DSN)
	if err != nil {
		logger.Error("error in connecting to database", slog.Any("error", err))
		return
	}
	defer DB.Close()

	err = postgres.RunMigrations(DB)
	if err != nil {
		logger.Error("error in running migrations", slog.Any("error", err))
		return
	}

	repo := repository.New(DB)

	service := service.New(repo)

	handler := handlers.NewCartHandler(service, logger)

	server := server.New(cfg.Server, logger)
	server.RegisterRoutes(handler)
	defer server.Close(ctx)

	err = server.Run(ctx)

	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("server closed")
			return
		}
		logger.Error("errs in running server", slog.Any("errs", err))
	}

}
