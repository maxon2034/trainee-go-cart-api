package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

func TestOpen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	cfg, err := config.Load("../../config")
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewPostgres(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cfg)
}
