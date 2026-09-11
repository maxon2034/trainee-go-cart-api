package db

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

func TestOpen(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	cfg, err := config.LoadConfig("../../config")
	if err != nil {
		t.Fatal(err)
	}

	db, err := Open(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}
