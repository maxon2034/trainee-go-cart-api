package app

import (
	"log/slog"
	"os"
)

// TODO: Logger init
func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// TODO: Server struct init

// TODO: DB init with check and migrations
