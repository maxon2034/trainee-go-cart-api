package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

func New(cfg config.Config) (*slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(cfg.Logger.Level)); err != nil {
		return nil, fmt.Errorf("error parsing level \"%s\": %v", cfg.Logger.Level, err)
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h), nil
}
