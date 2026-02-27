// Package logger ...
package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

// NewLogger ...
func NewLogger(databaseURL string) *slog.Logger {
	var lvl slog.LevelVar

	if err := lvl.UnmarshalText([]byte(databaseURL)); err != nil {
		lvl.Set(slog.LevelInfo)
	}

	handler := console.NewHandler(
		os.Stderr,
		&console.HandlerOptions{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		},
	)

	return slog.New(handler)
}
