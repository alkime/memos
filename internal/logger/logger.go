package logger

import (
	"log/slog"
	"os"

	"github.com/alkime/memos/internal/config"
)

// SetupLogger configures structured logging based on environment.
func SetupLogger(cfg *config.Config) *slog.Logger {
	// Determine log level
	logLevel := slog.LevelInfo
	if cfg.Env == "development" {
		logLevel = slog.LevelDebug
	}
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	// Create JSON handler for structured logging
	//nolint:exhaustruct // Using default values for other HandlerOptions fields
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger
}
