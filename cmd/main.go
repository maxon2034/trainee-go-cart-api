package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/maxon2034/trainee-go-cart-api/internal/app"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/db"
)

var cfgPath string = "config/"

func main() {
	logger := app.NewLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(".env"); err != nil {
		logger.Error("Error loading .env file", slog.Any("Error", err))
		return
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("Error loading config", slog.Any("Error", err))
	}

	DB, err := db.NewPostgres(ctx, cfg)
	if err != nil {
		logger.Error("Error initializing DB", slog.Any("Error", err))
	}

	defer func() {
		if err := DB.Close(); err != nil {
			logger.Error("Error closing DB", slog.Any("Error", err))
		}
	}()

	if err := db.RunMigrations(DB); err != nil {
		logger.Error("Error running migrations", slog.Any("Error", err))
	}

	mux := http.NewServeMux()

	// handlers

	if err = http.ListenAndServe(cfg.Server.Port, mux); err != nil {
		logger.Error("Error starting server", slog.Any("Error", err))
	}
}
