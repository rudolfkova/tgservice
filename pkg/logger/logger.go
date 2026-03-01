// Package logger ...
package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

// NewLogger ...
func NewLogger() *slog.Logger {
	handler := console.NewHandler(
		os.Stderr,
		&console.HandlerOptions{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		},
	)

	return slog.New(handler)
}
