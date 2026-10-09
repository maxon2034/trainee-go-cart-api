//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

func TestOpen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	cfg, err := config.Load("../../../config")
	if err != nil {
		t.Fatal(err)
	}

	db, err := New(ctx, cfg.DB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cfg)
}
