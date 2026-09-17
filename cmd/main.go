package main

import (
	"context"
	"log"
	"net/http"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/pkg/db/postgres"
)

var cfgPath string = "config/"

func main() {
	// logger := app.Logger
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Panic(err)

	}

	DB, err := postgres.New(ctx, cfg.DB.DSN)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := postgres.RunMigrations(DB); err != nil {
		log.Panic(err)
	}

	mux := http.NewServeMux()

	// handlers

	if err = http.ListenAndServe(cfg.Server.Port, mux); err != nil {
		log.Panic("error in starting server: ", err)
	}
}
