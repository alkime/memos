// Package main is the entry point for the memos web server.
package main

import (
	"log"

	"github.com/alkime/memos/internal/config"
	"github.com/alkime/memos/internal/logger"
	"github.com/alkime/memos/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup structured logging
	appLogger := logger.SetupLogger(cfg)

	// Log startup information
	appLogger.Info("Starting Memos server",
		"env", cfg.Env,
		"port", cfg.Port,
	)

	// Create and start server
	srv := server.New(cfg, appLogger)
	if err := server.Run(srv); err != nil {
		appLogger.Error("Failed to start server", "error", err)
		log.Fatalf("Fatal: %v", err)
	}
}
