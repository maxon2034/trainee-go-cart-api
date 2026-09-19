package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	h := slog.NewTextHandler(os.Stdout, nil)
	return slog.New(h)
}
