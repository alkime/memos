package main

import (
	"log/slog"
	"os"
)

// SetupLogger configures structured logging based on environment
func SetupLogger(config *Config) *slog.Logger {
	// Determine log level
	logLevel := slog.LevelInfo
	if config.Env == "development" {
		logLevel = slog.LevelDebug
	}
	if config.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger
}
