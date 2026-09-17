package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/maxon2034/trainee-go-cart-api/migrations"
	"github.com/pressly/goose/v3"
)

func New(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN or open driver: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgresql: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(migrations.EmbedFS)

	stdDB := db.DB
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect error: %w", err)
	}

	// Накатываем миграции из папки migrations
	if err := goose.Up(stdDB, "/../../migrations"); err != nil {
		return fmt.Errorf("goose up error: %w", err)
	}

	return nil
}
