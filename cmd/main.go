package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/db"
	"github.com/maxon2034/trainee-go-cart-api/internal/handlers"
	"github.com/maxon2034/trainee-go-cart-api/internal/repository"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
)

var cfgPath string = "config/"

func main() {
	// logger := app.Logger
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Print(err)
	}

	DB, err := db.NewPostgres(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := db.RunMigrations(DB); err != nil {
		log.Fatal(err)
	}

	repo := repository.New(DB)

	service := service.NewService(repo)

	server := handlers.NewServer(ctx, cfg, service)
	defer server.Shutdown(ctx)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
