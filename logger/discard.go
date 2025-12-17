package logger

import (
	"log/slog"
)

// NewDiscardLogger creates and returns a new Logger instance that discards all log entries without processing them.
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
