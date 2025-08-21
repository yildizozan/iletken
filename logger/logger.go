package logger

import (
	"log/slog"
	"os"
	"strings"

	"iletken/config"
)

// NewLogger creates a new logger
func NewLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	
	return slog.New(handler)
}
