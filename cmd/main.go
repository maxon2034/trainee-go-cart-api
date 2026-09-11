package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/db"
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

	mux := http.NewServeMux()

	// handlers

	if err = http.ListenAndServe(cfg.Server.Port, mux); err != nil {
		log.Fatal("error in starting server: ", err)
	}
}
